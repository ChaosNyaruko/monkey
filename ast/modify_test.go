package ast

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModify(t *testing.T) {
	one := func() Expression { return &IntegerLiteral{Value: 1} }
	two := func() Expression { return &IntegerLiteral{Value: 2} }

	turnOneIntoTwo := func(node Node) Node {
		i := node.(*IntegerLiteral)
		if i.Value != 1 {
			return node
		}
		i.Value = 2
		return i
	}

	tests := []struct {
		input    Node
		expected Node
	}{
		{
			one(),
			two(),
		},
		{
			&Program{
				Statements: []Statement{&ExpressionStatement{Expression: one()}},
			},
			&Program{
				Statements: []Statement{&ExpressionStatement{Expression: two()}},
			},
		},
		{
			&InfixExpression{
				Lhs: one(),
				Op:  "+",
				Rhs: two(),
			},
			&InfixExpression{
				Lhs: two(),
				Op:  "+",
				Rhs: two(),
			},
		},
		{
			&PrefixExpression{
				Op:  "-",
				Rhs: one(),
			},
			&PrefixExpression{
				Op:  "-",
				Rhs: two(),
			},
		},
		{
			&IndexExpression{
				Left:  one(),
				Index: one(),
			},
			&IndexExpression{
				Left:  two(),
				Index: two(),
			},
		},
		{
			&IfExpression{
				Condition: one(),
				If: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{
							Expression: one(),
						},
					},
				},
				Else: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{
							Expression: one(),
						},
					},
				},
			},
			&IfExpression{
				Condition: two(),
				If: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{
							Expression: two(),
						},
					},
				},
				Else: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{
							Expression: two(),
						},
					},
				},
			},
		},
		{
			&ReturnStatement{
				ReturnValue: one(),
			},
			&ReturnStatement{
				ReturnValue: two(),
			},
		},
		{
			&LetStatement{
				Name:  &Identifier{},
				Value: one(),
			},
			&LetStatement{
				Name:  &Identifier{},
				Value: two(),
			},
		},
		{
			&FunctionLiteral{
				Body: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{
							Expression: one(),
						},
					},
				},
			},
			&FunctionLiteral{
				Body: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{
							Expression: two(),
						},
					},
				},
			},
		},
		{
			&ArrayLiteral{
				Elements: []Expression{one()},
			},
			&ArrayLiteral{
				Elements: []Expression{two()},
			},
		},
		{
			&HashLiteral{
				Pairs: map[Expression]Expression{one(): one()},
			},
			&HashLiteral{
				Pairs: map[Expression]Expression{two(): two()},
			},
		},
	}
	for i, tc := range tests {
		modified := Modify(tc.input, turnOneIntoTwo)
		// assert.Equal(t, modified, tc.expected, "%d: %v", i, tc.input)
		switch h := modified.(type) {
		// NOTE: ugly test :(
		case *HashLiteral:
			for k, v := range h.Pairs {
				k := k.(*IntegerLiteral)
				assert.Equal(t, two(), k)
				v := v.(*IntegerLiteral)
				assert.Equal(t, two(), v)
			}
		default:
			assert.True(t, reflect.DeepEqual(modified, tc.expected), "%d: %v", i, tc.input)
		}
	}
}
