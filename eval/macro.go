package eval

import (
	"fmt"
	"slices"

	"github.com/ChaosNyaruko/monkey/ast"
	"github.com/ChaosNyaruko/monkey/object"
)

// DefineMacros walks through the program ast, extract all macro(s) into env, and remove them from the ast.
// For simplicity, we only allow to define macro at the top layer of the progrom.
func DefineMacros(p *ast.Program, env *object.Environment) error {
	removed := []int{}
	// extract all the defined macros
	for i, stmt := range p.Statements {
		if isMacroDef(stmt) {
			if err := addMacro(stmt, env); err != nil {
				return err
			}
			removed = append(removed, i)
		}
	}
	// remove the macros from the original ast
	// []{x, y, z, a, b,c } -> []{x, z, a, c}
	for i := len(removed) - 1; i >= 0; i-- {
		p.Statements = slices.Delete(p.Statements, removed[i], removed[i]+1)
	}
	return nil
}

func addMacro(node ast.Statement, env *object.Environment) error {
	let, ok := node.(*ast.LetStatement)
	if !ok {
		return fmt.Errorf("should be a let statement, but got: %T", node)
	}
	v, ok := let.Value.(*ast.MacroLiteral)
	if !ok {
		return fmt.Errorf("should be assigned with a macro, but got: %T", let.Value)
	}

	macro := &object.Macro{
		Parameters: v.Parameters,
		Body:       v.Body,
		Env:        env,
	}
	_, err := env.Set(let.Name.String(), macro)
	return err
}

func isMacroDef(node ast.Statement) bool {
	let, ok := node.(*ast.LetStatement)
	if !ok {
		return false
	}
	_, ok = let.Value.(*ast.MacroLiteral)
	if !ok {
		return false
	}

	return true
}

func isMacroCall(node *ast.CallExpression, env *object.Environment) (*object.Macro, bool) {
	id, ok := node.F.(*ast.Identifier)
	if !ok {
		return nil, false
	}
	obj, err := env.Get(id.String())
	if err != nil {
		return nil, false
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		return nil, false
	}

	return macro, true
}

// ExpandMacros reads the macro literal by name in the environment, and "expand" it into a real AST(before evaluation).
func ExpandMacros(p *ast.Program, env *object.Environment) ast.Node {
	f := func(node ast.Node) ast.Node {
		call, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}
		m, ok := isMacroCall(call, env)
		if !ok {
			return node
		}
		newEnv := object.NewEnvironment(env)
		// log.Printf("expand %v", m.Inspect())
		// process args list
		// pass the "quoted" ast, to make the args not be evaluated before body evaluation.
		for i, a := range call.Arguments {
			qa := &object.Quote{Node: a}
			// log.Printf("qa: %s", a)
			// ea, _ := Eval(a, env)
			// store the args into env(for the individual macro call)
			newEnv.Set(m.Parameters[i].Value, qa)
		}
		// process body
		// log.Println(fmt.Sprintf("before expandedNode: %s", m.Body.String()))
		expandedNode, err := Eval(m.Body, newEnv)
		// log.Println(fmt.Sprintf("after expandedNode: %s, %s", expandedNode.Type(), expandedNode.Inspect()))
		if err != nil {
			return node
		}
		quote, ok := expandedNode.(*object.Quote)
		if !ok {
			panic("macros should only return QUOTEs(AST-nodes)")
		}
		return quote.Node
	}
	return ast.Modify(p, f)
}
