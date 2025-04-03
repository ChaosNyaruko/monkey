package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ChaosNyaruko/monkey/ast"
	"github.com/ChaosNyaruko/monkey/lexer"
	"github.com/ChaosNyaruko/monkey/object"
	"github.com/ChaosNyaruko/monkey/parser"
)

func TestEvalBoolean(t *testing.T) {
	type testcase struct {
		input    string
		expected bool
		hasError bool
	}
	tests := []testcase{
		{"true", true, false},
		{"false", false, false},
		{"TRUE", false, true}, // should report an error
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if !assert.Equal(t, tc.hasError, err != nil, "err: %v", err) {
			t.Fatalf("input to object err: %v", err)
		}
		if err == nil {
			testBooleanObject(t, got, tc.expected)
		}
	}
}

func TestEvalInteger(t *testing.T) {
	type testcase struct {
		input    string
		expected int
		hasError bool
	}
	tests := []testcase{
		{"5", 5, false},
		{"123", 123, false},
		{"111111111111111111111111111111111111", 0, true}, // should report an error
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if !assert.Equal(t, tc.hasError, err != nil, "err: %v", err) {
			t.Fatalf("input to object err: %v", err)
		}
		if err == nil {
			testIntegerObject(t, got, tc.expected)
		}
	}
}

func stringToAst(input string) (ast.Node, error) {
	l := lexer.New(input)
	parser := parser.New(l)
	return parser.ParseProgram(), parser.Error()
}

func stringToObject(input string) (object.Object, error) {
	if ob, err := stringToAst(input); err != nil {
		return nil, err
	} else {
		return Eval(ob)
	}

}

func testIntegerObject(t *testing.T, got object.Object, expected int) {
	i, ok := got.(*object.Integer)
	assert.True(t, ok, "expected an integer object, but got: %T", got)
	assert.Equal(t, expected, i.Value)
}

func testBooleanObject(t *testing.T, got object.Object, expected bool) {
	i, ok := got.(*object.Boolean)
	assert.True(t, ok, "expected a boolean object, but got: %T", got)
	assert.Equal(t, expected, i.Value)
}
