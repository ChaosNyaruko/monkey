package eval

import (
	"fmt"

	"github.com/ChaosNyaruko/monkey/ast"
	"github.com/ChaosNyaruko/monkey/object"
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
		f, err := Eval(node.F, env)
		if err != nil {
			return nil, err
		}
		// eval arguments
		args, err := evalArguments(node.Arguments, env)
		return callFunction(f, args)
	}
	return nil, fmt.Errorf("unsupported object type: %T\n", node)
}

func callFunction(fn object.Object, args []object.Object) (object.Object, error) {
	f, ok := fn.(*object.Function)
	if !ok {
		return nil, fmt.Errorf("%v is not a Function Object", f.Inspect())
	}
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
}

func evalArguments(args []ast.Expression, env *object.Environment) ([]object.Object, error) {
	var res = make([]object.Object, 0, len(args))
	for _, a := range args {
		v, err := Eval(a, env)
		if err != nil {
			return nil, fmt.Errorf("passing arguments error: [%v]%v", a, err)
		}
		res = append(res, v)
	}
	return res, nil
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) (object.Object, error) {
	return env.Get(node.Value)
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
			return nil, fmt.Errorf("expected integer after '-', but got %T\n", rhs)
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
