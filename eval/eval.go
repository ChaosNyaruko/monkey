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

func Eval(node ast.Node) (object.Object, error) {
	switch node := node.(type) {
	case *ast.Program:
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.IntegerLiteral:
		return &object.Integer{
			Value: node.Value,
		}, nil
	case *ast.BooleanExpression:
		return boolToBoolean(node.Value), nil
	case *ast.PrefixExpression:
		rhs, err := Eval(node.Rhs)
		if err != nil {
			return nil, err
		}
		res, err := evalPrefixExpression(node.Op, rhs)
		return res, err
	case *ast.InfixExpression:
		lhs, err := Eval(node.Lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := Eval(node.Rhs)
		if err != nil {
			return nil, err
		}
		res, err := evalInfixExpression(node.Op, lhs, rhs)
		return res, err
	case *ast.LetStatement:
		return NULL, nil
	case *ast.ReturnStatement: // TODO: should it have a "value" itself?
		return NULL, nil
	}
	return nil, fmt.Errorf("unsupported object type: %T\n", node)
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
		return &object.Boolean{
			Value: l.Value == r.Value,
		}, nil
	case "!=":
		return &object.Boolean{
			Value: l.Value != r.Value,
		}, nil
	case "<":
		return &object.Boolean{
			Value: l.Value < r.Value,
		}, nil
	case ">":
		return &object.Boolean{
			Value: l.Value > r.Value,
		}, nil
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

func evalStatements(stmts []ast.Statement) (object.Object, error) {
	var res object.Object
	var err error
	for _, s := range stmts {
		res, err = Eval(s)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
