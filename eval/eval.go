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
		if node.Value {
			return TRUE, nil
		}
		return FALSE, nil
	case *ast.LetStatement:
		return NULL, nil
	case *ast.ReturnStatement: // TODO: should it have a "value" itself?
		return NULL, nil
	}
	return nil, fmt.Errorf("unsupported object type: %T\n", node)
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
