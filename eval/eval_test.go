package eval

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ChaosNyaruko/monkey/ast"
	"github.com/ChaosNyaruko/monkey/lexer"
	"github.com/ChaosNyaruko/monkey/object"
	"github.com/ChaosNyaruko/monkey/parser"
)

func TestEvalLetStatement(t *testing.T) {
	type testcase struct {
		input    string
		expected any
		err      error
	}
	tests := []testcase{
		{"let a = 5; a", 5, nil},
		{"let a = 5; a;", 5, nil},
		{"let a = 5; let b = a; let c = (a + b)*2; c", 20, nil},
		{"let a = x;", 0, fmt.Errorf("undefined identifier")},
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if err != nil {
			t.Logf("err: %v", err)
			assert.NotNil(t, tc.err, "input: %v", tc.input)
			require.Conditionf(t, func() bool { return strings.Contains(err.Error(), tc.err.Error()) },
				"input: %v, expected err: %v, but got %v", tc.input, tc.err, err)
			continue
		}

		switch v := tc.expected.(type) {
		case int:
			testIntegerObject(t, tc.input, got, v)
		case bool:
			testBooleanObject(t, tc.input, got, v)
		default:
			testNull(t, tc.input, got)
		}
	}
}

func TestEvalIfElse(t *testing.T) {
	type testcase struct {
		input    string
		expected any
		hasError bool
	}
	tests := []testcase{
		{"if (true) {100}", 100, false},
		{"if (false) {100}", NULL, false},
		{"if (1) {100}", 100, false},
		{"if (0) {100}", 100, false},
		{"if (null) {100}", NULL, false},
		{"if (1<2) {1} else {2}", 1, false},
		{"if (1>2) {1} else {2}", 2, false},
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if !assert.Equal(t, tc.hasError, err != nil, "err: %v", err) {
			t.Fatalf("input %v, to object err: %v", tc.input, err)
		}
		if err == nil {
			switch v := tc.expected.(type) {
			case int:
				testIntegerObject(t, tc.input, got, v)
			case bool:
				testBooleanObject(t, tc.input, got, v)
			default:
				testNull(t, tc.input, got)
			}
		}
	}
}

func TestEvalBang(t *testing.T) {
	type testcase struct {
		input    string
		expected bool
		hasError bool
	}
	tests := []testcase{
		{"!true", false, false},
		{"!false", true, false},
		{"!5", false, false},
		{"!0", false, false},
		{"!!5", true, false},
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if !assert.Equal(t, tc.hasError, err != nil, "err: %v", err) {
			t.Fatalf("input to object err: %v", err)
		}
		if err == nil {
			testBooleanObject(t, tc.input, got, tc.expected)
		}
	}
}

func TestEvalBoolean(t *testing.T) {
	type testcase struct {
		input    string
		expected bool
		hasError bool
	}
	tests := []testcase{
		{"true", true, false},
		{"false", false, false},
		{"true == true", true, false},
		{"false == false", true, false},
		{"true != false", true, false},
		{"false != true", true, false},
		{"true != true", false, false},
		{"false != false", false, false},
		{"true == false", false, false},
		{"false == true", false, false},
		{"1 < 2", true, false},
		{"1 > 2", false, false},
		{"2 == 2", true, false},
		{"2 != 2", false, false},
		{"2 == (1+1)", true, false},
		{"3 == 2 * (1+1)", false, false},
		{"3 != 2 * (1+1)", true, false},
		{"TRUE", false, true}, // should report an error
		{"false < true", false, true},
		{"false > true", false, true},
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if !assert.Equal(t, tc.hasError, err != nil, "err: %v", err) {
			t.Fatalf("input to object err: %v", err)
		}
		if err == nil {
			testBooleanObject(t, tc.input, got, tc.expected)
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
		{"-5", -5, false},
		{"-10", -10, false},
		{"1+2+3+4-10", 0, false},
		{"-1+2*-2", -5, false},
		{"10/2*3", 15, false},
		{"(1+3)*-4", -16, false},
		{"(4+3)*(4)+-29", -1, false},
		{"111111111111111111111111111111111111", 0, true}, // should report an error
		{"-true", 0, true},                                // should report an error
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if !assert.Equal(t, tc.hasError, err != nil, "err: %v", err) {
			t.Fatalf("input to object err: %v", err)
		}
		if err == nil {
			testIntegerObject(t, tc.input, got, tc.expected)
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
		env := object.NewEnvironment()
		return Eval(ob, env)
	}

}

func testIntegerObject(t *testing.T, input string, got object.Object, expected int) {
	i, ok := got.(*object.Integer)
	assert.True(t, ok, "expected an integer object, but got: %T", got)
	assert.Equal(t, expected, i.Value, input)
}

func testBooleanObject(t *testing.T, input string, got object.Object, expected bool) {
	i, ok := got.(*object.Boolean)
	assert.True(t, ok, "expected a boolean object, but got: %T", got)
	assert.Equal(t, expected, i.Value, input)
}

func testNull(t *testing.T, input string, got object.Object) {
	assert.Equal(t, NULL, got, "input: %v, expected 'null', but got: %T", input, got)
}
