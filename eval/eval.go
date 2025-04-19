package eval

import (
	"fmt"

	"github.com/ChaosNyaruko/monkey/ast"
	"github.com/ChaosNyaruko/monkey/object"
	"github.com/ChaosNyaruko/monkey/token"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func boolToBoolean(value bool) *object.Boolean {
	if value {
		return TRUE
	}
	return FALSE
}

// isTrue convert others to real bools.
// i.e. condition -> true/false
// - integer -> false
// - boolean -> true/false
// - null -> false
func isTrue(obj object.Object) bool {
	switch obj {
	case TRUE:
		return true
	case FALSE:
		return false
	case NULL:
		return false
	}
	return true
}

func evalIfElse(node *ast.IfExpression, env *object.Environment) (object.Object, error) {
	condition, err := Eval(node.Condition, env)
	if err != nil {
		return nil, err
	}
	if isTrue(condition) {
		return Eval(node.If, env)
	} else if node.Else != nil {
		return Eval(node.Else, env)
	}
	// not hit if, but no else expression.
	return NULL, nil
}

func Eval(node ast.Node, env *object.Environment) (object.Object, error) {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.BlockStatement:
		return evalBlockStatements(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IfExpression:
		return evalIfElse(node, env)
	case *ast.ArrayLiteral:
		a := &object.Array{
			Elements: []object.Object{},
		}
		e, err := evalExpressions(node.Elements, env)
		if err != nil {
			return nil, err
		}
		a.Elements = e
		return a, nil
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	case *ast.IndexExpression:
		return evalIndexExpression(node, env)
	case *ast.StringLiteral:
		return &object.String{
			Value: node.Value,
		}, nil
	case *ast.IntegerLiteral:
		return &object.Integer{
			Value: node.Value,
		}, nil
	case *ast.BooleanExpression:
		return boolToBoolean(node.Value), nil
	case *ast.NullExpression:
		return NULL, nil
	case *ast.PrefixExpression:
		rhs, err := Eval(node.Rhs, env)
		if err != nil {
			return nil, err
		}
		res, err := evalPrefixExpression(node.Op, rhs)
		return res, err
	case *ast.InfixExpression:
		lhs, err := Eval(node.Lhs, env)
		if err != nil {
			return nil, err
		}
		rhs, err := Eval(node.Rhs, env)
		if err != nil {
			return nil, err
		}
		res, err := evalInfixExpression(node.Op, lhs, rhs)
		return res, err
	case *ast.LetStatement:
		val, err := Eval(node.Value, env)
		if err != nil {
			return nil, err
		}
		_, err = env.Set(node.Name.Value, val)
		return NULL, err
	case *ast.Identifier:
		// TODO: let x = (let c = 1);
		return evalIdentifier(node, env)
	case *ast.ReturnStatement: // return's value if the expression after the "return".
		// return 2;
		rValue, err := Eval(node.ReturnValue, env) // rValue -> Integar
		return &object.ReturnValue{
			Value: rValue,
		}, err
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{
			Parameters: params,
			Body:       body,
			Env:        env,
		}, nil
	case *ast.CallExpression:
		if node.F.TokenLiteral() == "eval" {
			if len(node.Arguments) != 1 {
				return nil, fmt.Errorf("eval should and only should have one argument\n")
			}
			return evalLiteral(node.Arguments[0], env)
		}
		if node.F.TokenLiteral() == "quote" {
			if len(node.Arguments) != 1 {
				return nil, fmt.Errorf("quote should and only should have one argument\n")
			}
			return quote(node.Arguments[0], env)
		}
		f, err := Eval(node.F, env)
		if err != nil {
			return nil, err
		}
		// eval arguments
		args, err := evalExpressions(node.Arguments, env)
		if err != nil {
			return nil, err
		}
		return callFunction(f, args)
	}
	return nil, fmt.Errorf("unsupported object type: %T\n", node)
}

func callFunction(fn object.Object, args []object.Object) (object.Object, error) {
	switch f := fn.(type) {
	case *object.Function:
		newEnv := object.NewEnvironment(f.Env)
		for i, p := range f.Parameters {
			newEnv.Set(p.Value, args[i])
		}

		val, err := Eval(f.Body, newEnv)
		if err != nil {
			return nil, err
		}
		if v, ok := val.(*object.ReturnValue); ok {
			return v.Value, nil
		}
		return val, nil
	case *object.Builtin:
		return f.Fn(args...)
	}
	return nil, fmt.Errorf("%v is not callable", fn.Inspect())
}

func evalExpressions(args []ast.Expression, env *object.Environment) ([]object.Object, error) {
	var res = make([]object.Object, 0, len(args))
	for _, a := range args {
		v, err := Eval(a, env)
		if err != nil {
			return nil, fmt.Errorf("passing exp error: [%v]%v", a, err)
		}
		res = append(res, v)
	}
	return res, nil
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) (object.Object, error) {
	if obj, err := env.Get(node.Value); err == nil {
		return obj, nil
	}

	bti, ok := builtins[node.Value]
	if ok {
		return bti, nil
	}

	return nil, fmt.Errorf("undefined identifier: %s\n", node.Value)
}

func evalInfixString(op string, l, r *object.String) (object.Object, error) {
	switch op {
	case "+":
		return &object.String{
			Value: l.Value + r.Value,
		}, nil
	case "==":
		return &object.Boolean{
			Value: l.Value == r.Value,
		}, nil
	case "!=":
		return &object.Boolean{
			Value: l.Value != r.Value,
		}, nil
	}
	return nil, fmt.Errorf("unsupported infix operator for strings: %q %s %q\n", l.Inspect(), op, r.Inspect())
}

func evalInfixInteger(op string, l, r *object.Integer) (object.Object, error) {
	switch op {
	case "+":
		return &object.Integer{
			Value: l.Value + r.Value,
		}, nil
	case "-":
		return &object.Integer{
			Value: l.Value - r.Value,
		}, nil
	case "*":
		return &object.Integer{
			Value: l.Value * r.Value,
		}, nil
	case "/":
		return &object.Integer{
			Value: l.Value / r.Value,
		}, nil
	case "==":
		return boolToBoolean(l.Value == r.Value), nil
	case "!=":
		return boolToBoolean(l.Value != r.Value), nil
	case "<":
		return boolToBoolean(l.Value < r.Value), nil
	case ">":
		return boolToBoolean(l.Value > r.Value), nil
	}
	return nil, fmt.Errorf("unsupported infix operator for integers: %q\n", op)
}

func evalInfixExpression(op string, lhs, rhs object.Object) (object.Object, error) {
	lType, rType := lhs.Type(), rhs.Type()

	if lType == object.INTEGER_OBJ && rType == object.INTEGER_OBJ {
		l, r := lhs.(*object.Integer), rhs.(*object.Integer)
		return evalInfixInteger(op, l, r)
	} else if lType == object.STRING_OBJ && rType == object.STRING_OBJ {
		l, r := lhs.(*object.String), rhs.(*object.String)
		return evalInfixString(op, l, r)
	}

	if lType == object.BOOLEAN_OBJ && rType == object.BOOLEAN_OBJ {
		switch op {
		case "==":
			return boolToBoolean(lhs == rhs), nil
		case "!=":
			return boolToBoolean(lhs != rhs), nil
		}
		return nil, fmt.Errorf("illegal operands for %q, lhs: %q, rhs: %q\n", op, lhs.Inspect(), rhs.Inspect())
	}
	return nil, fmt.Errorf("illegal operands for %q, lhs: %q, rhs: %q\n", op, lhs.Inspect(), rhs.Inspect())
}

func evalPrefixExpression(op string, rhs object.Object) (object.Object, error) {
	if op == "!" {
		switch rhs {
		case TRUE:
			return FALSE, nil
		case FALSE:
			return TRUE, nil
		case NULL:
			return TRUE, nil
		default: // TODO: 0 shoud be considered as true or false?
			return FALSE, nil
		}
	} else if op == "-" {
		value, ok := rhs.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf("expected integer after '-', but got %v\n", rhs.Type())
		}
		return &object.Integer{
			Value: -value.Value,
		}, nil
	}
	return nil, fmt.Errorf("unsupported prefix operator: %q\n", op)
}

func evalProgram(stmts []ast.Statement, env *object.Environment) (object.Object, error) {
	var res object.Object
	var err error
	for _, s := range stmts {
		res, err = Eval(s, env)
		if err != nil {
			return nil, err
		}
		if r, ok := res.(*object.ReturnValue); ok {
			return r.Value, err
		}
	}
	return res, nil
}

func evalBlockStatements(stmts []ast.Statement, env *object.Environment) (object.Object, error) {
	/*
	   if (true) {
	       if (true) {
	           return 1;
	       } // -> blocks = object.ReturnValue{Value: object.Integer{1}}
	       return 2;
	   }
	*/

	/*
	   if (true) {
	     blocks // object.ReturnValue
	   } // -> blocks.Value -> object.Integer{1}
	*/
	var res object.Object
	var err error
	for _, s := range stmts {
		res, err = Eval(s, env)
		if err != nil {
			return nil, err
		}
		if r, ok := res.(*object.ReturnValue); ok {
			// fmt.Printf("return value in blockstatement: %v\n", s.String())
			return r, err
		}
	}
	return res, nil
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) (object.Object, error) {
	pairs := make(map[object.HashKey]object.HashPair)

	var err error
	for key, value := range node.Pairs {
		var k, v object.Object
		if k, err = Eval(key, env); err != nil {
			return nil, fmt.Errorf("eval key: %s err: %v\n", key.String(), err)
		}
		hk, ok := k.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("%v is not hashable\n", k.Type())
		}
		if v, err = Eval(value, env); err != nil {
			return nil, fmt.Errorf("eval key: %s err: %v\n", key.String(), err)
		}
		pairs[hk.HashKey()] = object.HashPair{
			Key:   k,
			Value: v,
		}
	}
	return &object.Hash{
		Pairs: pairs,
	}, nil
}

func evalIndexExpression(node *ast.IndexExpression, env *object.Environment) (object.Object, error) {
	left, err := Eval(node.Left, env)
	if err != nil {
		return nil, err
	}
	index, err := Eval(node.Index, env)
	if err != nil {
		return nil, err
	}
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	}
	return nil, fmt.Errorf("index %s on %s is not supported", index.Type(), left.Type())
}

func evalArrayIndexExpression(array, int object.Object) (object.Object, error) {
	a := array.(*object.Array)
	i := int.(*object.Integer)
	if i.Value >= len(a.Elements) || i.Value < 0 {
		return nil, fmt.Errorf("index out of bounds, len:%d, visit:%d", len(a.Elements), i.Value)
	}
	return a.Elements[i.Value], nil
}

func evalHashIndexExpression(hm, key object.Object) (object.Object, error) {
	a := hm.(*object.Hash)
	i, ok := key.(object.Hashable)
	if !ok {
		return nil, fmt.Errorf("%v is not hashable", key.Type())
	}
	res, ok := a.Pairs[i.HashKey()]
	if !ok {
		return NULL, nil
	}
	return res.Value, nil
}

func quote(node ast.Node, env *object.Environment) (object.Object, error) {
	node, err := evalUnquote(node, env)
	return &object.Quote{
		Node: node,
	}, err
}

func evalUnquote(quoted ast.Node, env *object.Environment) (ast.Node, error) {
	// (quote 1 2 (+ 3 4) unquote(2+3)) -> (quote 1 2 (+3 4) 5)
	f := func(node ast.Node) ast.Node {
		// node is unquote or not
		if !isUnquote(node) {
			return node
		}

		call := node.(*ast.CallExpression)
		n, err := evalNewAstNode(call.Arguments[0], env)
		if err != nil {
			return nil
		}
		return n
	}
	// TODO: better error process during modifying
	n := ast.Modify(quoted, f)
	if n == nil {
		return nil, fmt.Errorf("evalUnquote in quote err: %v", quoted.String())
	}
	return n, nil
}

func isUnquote(node ast.Node) bool {
	call, ok := node.(*ast.CallExpression)
	if !ok {
		return false
	}
	return call.F.TokenLiteral() == "unquote" && len(call.Arguments) == 1
}

func evalNewAstNode(node ast.Node, env *object.Environment) (ast.Node, error) {
	unquotedObj, err := Eval(node, env)
	if err != nil {
		return nil, err
	}
	switch obj := (unquotedObj).(type) {
	case *object.Integer:
		return &ast.IntegerLiteral{
			Token: token.Token{
				Type:    token.INT,
				Literal: fmt.Sprintf("%d", obj.Value),
			},
			Value: obj.Value,
		}, nil
	case *object.Quote:
		return obj.Node, nil

	case *object.Boolean:
		t := token.Token{
			Type:    token.FALSE,
			Literal: "false",
		}
		if obj.Value {
			t.Type = token.TRUE
			t.Literal = "true"
		}
		return &ast.BooleanExpression{
			Token: t,
			Value: obj.Value,
		}, nil
	}
	return nil, fmt.Errorf("TODO: cannot convert %s into ast", unquotedObj.Inspect())
}

func evalLiteral(node ast.Node, env *object.Environment) (object.Object, error) {
	q, err := Eval(node, env)
	if err != nil {
		return nil, err
	}
	e, ok := q.(*object.Quote)
	if !ok {
		return nil, fmt.Errorf("the 'eval' should be applied to a QUOTE, but got: %s", e.Inspect())
	}
	return Eval(e.Node, env)
}
