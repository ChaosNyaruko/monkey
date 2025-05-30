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

func TestHashIndex(t *testing.T) {
	type testcase struct {
		input    string
		expected any
		err      error
	}
	for _, tc := range []testcase{
		{
			input: `let key = "john";
			{key: "john smith", "one": 4-3, true: "TRUE!!!", false: 100000, 2 + 5: "7"}["one"]
			`,
			expected: 1,
			err:      nil,
		},
		{
			input: `let key = "john";
			{key: "john smith", "one": 4-3, true: "TRUE!!!", false: 100000, 2 + 5: "7"}[key]
			`,
			expected: "john smith",
			err:      nil,
		},
		{
			input:    `{fn() {return 1;} : 1}`,
			expected: NULL,
			err:      fmt.Errorf("not hashable"),
		},
		{
			input:    `{}[fn() {1 + 2}]`,
			expected: NULL,
			err:      fmt.Errorf("not hashable"),
		},
		{
			input:    `{}[2]`,
			expected: NULL,
			err:      nil,
		},
	} {
		input := tc.input
		got, err := stringToObject(input)
		if err != nil {
			assert.NotNil(t, tc.err, "got err: %v, input: %v", err, tc.input)
			require.Conditionf(t, func() bool { return strings.Contains(err.Error(), tc.err.Error()) },
				"input: %v, expected err: %v, but got %v", tc.input, tc.err, err)
			continue
		}
		switch v := tc.expected.(type) {
		case int:
			testIntegerObject(t, tc.input, got, v)
		case bool:
			testBooleanObject(t, tc.input, got, v)
		case string:
			assert.Equal(t, tc.expected, got.Inspect(), "input: %v", tc.input)
		case []int:
			g := got.(*object.Array)
			assert.Equal(t, len(v), len(g.Elements))
			for i, o := range g.Elements {
				testIntegerObject(t, o.Inspect(), o, v[i])
			}
		default:
			testNull(t, tc.input, got)
		}
	}
}

func TestHashLiteral(t *testing.T) {
	type testcase struct {
		input string
		err   error
	}
	for _, tc := range []testcase{
		{
			input: `let key = "john";
			{key: "john smith", "one": 4-3, true: "TRUE!!!", false: 100000, 2 + 5: "7"}
			`,
			err: nil,
		},
		{
			input: `{fn() {return 1;} : 1}`,
			err:   fmt.Errorf("not hashable"),
		},
	} {
		input := tc.input
		got, err := stringToObject(input)
		if err != nil {
			assert.NotNil(t, tc.err)
			require.Conditionf(t, func() bool { return strings.Contains(err.Error(), tc.err.Error()) },
				"input: %v, expected err: %v, but got %v", tc.input, tc.err, err)
			continue
		}
		h, ok := got.(*object.Hash)
		assert.True(t, ok, "expected hash object, but got: %v", got.Type())

		expected := map[object.HashKey]object.Object{
			(&object.String{
				Value: "john",
			}).HashKey(): &object.String{
				Value: "john smith",
			},
			(&object.String{
				Value: "one",
			}).HashKey(): &object.Integer{
				Value: 1,
			},
			(&object.Boolean{
				Value: true,
			}).HashKey(): &object.String{
				Value: "TRUE!!!",
			},
			(&object.Boolean{
				Value: false,
			}).HashKey(): &object.Integer{
				Value: 100000,
			},
			(&object.Integer{
				Value: 7,
			}).HashKey(): &object.String{
				Value: "7",
			},
		}
		assert.Equal(t, len(expected), len(h.Pairs))
		for k, v := range expected {
			vgot, ok := h.Pairs[k]
			assert.True(t, ok, "%v/%v not found in got", k, v.Inspect())
			vg := vgot.Value
			assert.Equal(t, v.Type(), vg.Type())
			assert.Equal(t, v.Inspect(), vg.Inspect())
		}
	}
}

func TestIndex(t *testing.T) {
	type testcase struct {
		input    string
		expected any
		err      error
	}
	tests := []testcase{
		{`[1,2,3][2]`, 3, nil},
		{`1[1]`, 0, fmt.Errorf("index INTEGER on INTEGER is not supported")},
		{`[1,2,3][true]`, 0, fmt.Errorf("index BOOLEAN on ARRAY is not supported")},
		{`[1,2,3][3]`, 0, fmt.Errorf("out of bounds, len:3, visit:3")},
		{`[1,2*3, 5+1][1]`, 6, nil},
		{`let a = [1,2,3,4,[5,6]]; a[4][1]`, 6, nil},
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if err != nil {
			assert.NotNil(t, tc.err, "input: %v, actual: %v", tc.input, err)
			require.Conditionf(t, func() bool { return strings.Contains(err.Error(), tc.err.Error()) },
				"input: %v, expected err: %v, but got %v", tc.input, tc.err, err)
			continue
		}

		switch v := tc.expected.(type) {
		case int:
			testIntegerObject(t, tc.input, got, v)
		case bool:
			testBooleanObject(t, tc.input, got, v)
		case string:
			assert.Equal(t, tc.expected, got.Inspect(), "input: %v", tc.input)
		case []int:
			g := got.(*object.Array)
			assert.Equal(t, len(v), len(g.Elements))
			for i, o := range g.Elements {
				testIntegerObject(t, o.Inspect(), o, v[i])
			}
		default:
			testNull(t, tc.input, got)
		}
	}
}

func TestArray(t *testing.T) {
	type testcase struct {
		input    string
		expected any
		err      error
	}
	tests := []testcase{
		{`[1,2*3, 5+1]`, []int{1, 6, 6}, nil},
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if err != nil {
			assert.NotNil(t, tc.err, "input: %v, actual: %v", tc.input, err)
			require.Conditionf(t, func() bool { return strings.Contains(err.Error(), tc.err.Error()) },
				"input: %v, expected err: %v, but got %v", tc.input, tc.err, err)
			continue
		}

		switch v := tc.expected.(type) {
		case int:
			testIntegerObject(t, tc.input, got, v)
		case bool:
			testBooleanObject(t, tc.input, got, v)
		case string:
			assert.Equal(t, tc.expected, got.Inspect(), "input: %v", tc.input)
		case []int:
			g := got.(*object.Array)
			assert.Equal(t, len(v), len(g.Elements))
			for i, o := range g.Elements {
				testIntegerObject(t, o.Inspect(), o, v[i])
			}
		default:
			testNull(t, tc.input, got)
		}
	}
}

func TestBuiltin(t *testing.T) {
	type testcase struct {
		input    string
		expected any
		err      error
	}
	tests := []testcase{
		{`len("")`, 0, nil},
		{`len("hello")`, 5, nil},
		{`len("hello\n")`, 7, nil},
		{`len(1)`, 0, fmt.Errorf("not supported on INTEGER")},
		{`len("one", "two")`, 0, fmt.Errorf("wrong number of arguments, expected 1, but got 2")},
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if err != nil {
			assert.NotNil(t, tc.err, "input: %v, actual: %v", tc.input, err)
			require.Conditionf(t, func() bool { return strings.Contains(err.Error(), tc.err.Error()) },
				"input: %v, expected err: %v, but got %v", tc.input, tc.err, err)
			continue
		}

		switch v := tc.expected.(type) {
		case int:
			testIntegerObject(t, tc.input, got, v)
		case bool:
			testBooleanObject(t, tc.input, got, v)
		case string:
			assert.Equal(t, tc.expected, got.Inspect(), "input: %v", tc.input)
		default:
			testNull(t, tc.input, got)
		}
	}
}

func TestStringConcat(t *testing.T) {
	type testcase struct {
		input    string
		expected any
		err      error
	}
	tests := []testcase{
		{`"hello "+"world"`, "hello world", nil},
		{`"hello "-"world"`, "", fmt.Errorf(`unsupported infix operator for strings`)},
		{`"hello" == "world"`, false, nil},
		{`"hello" == "hello"`, true, nil},
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if err != nil {
			assert.NotNil(t, tc.err, "input: %v, actual: %v", tc.input, err)
			require.Conditionf(t, func() bool { return strings.Contains(err.Error(), tc.err.Error()) },
				"input: %v, expected err: %v, but got %v", tc.input, tc.err, err)
			continue
		}

		switch v := tc.expected.(type) {
		case int:
			testIntegerObject(t, tc.input, got, v)
		case bool:
			testBooleanObject(t, tc.input, got, v)
		case string:
			assert.Equal(t, tc.expected, got.Inspect(), "input: %v", tc.input)
		default:
			testNull(t, tc.input, got)
		}
	}
}

func TestStringLiteral(t *testing.T) {
	type testcase struct {
		input    string
		expected any
		err      error
	}
	tests := []testcase{
		// {`"hello world"`, `hello world`, nil},
		// {`"hello world\n"`, `hello world\n`, nil},
		{`"hello world`, "hello world", nil},
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if err != nil {
			assert.NotNil(t, tc.err, "input: %v, actual: %v", tc.input, err)
			require.Conditionf(t, func() bool { return strings.Contains(err.Error(), tc.err.Error()) },
				"input: %v, expected err: %v, but got %v", tc.input, tc.err, err)
			continue
		}

		switch v := tc.expected.(type) {
		case int:
			testIntegerObject(t, tc.input, got, v)
		case bool:
			testBooleanObject(t, tc.input, got, v)
		case string:
			assert.Equal(t, tc.expected, got.Inspect(), "input: %v", tc.input)
		default:
			testNull(t, tc.input, got)
		}
	}
}

func TestCallFunction(t *testing.T) {
	type testcase struct {
		input    string
		expected any
		err      error
	}
	tests := []testcase{
		{`
			let add = fn(x) {
				fn(y) {
					x + y
				}
			};

			let addTwo = add(2);
			let c = addTwo(8);
			return c;
`,
			10,
			nil,
		},
		{"let add = fn(x, y, c) { return fn() {return x + y + c; }();}; add(1,2,5)", 8, nil},
		{"let c = 5; let add = fn(x, y) {return x + y + c;}; add(1,2)", 8, nil},
		// {"let c = 5; let add5 = fn(x, y) {return x + y + c;}; add(1,2)", 8, error},
		{"let add = fn(x, y) {x + y;}; add(1,2)", 3, nil},
		{"let add = fn(x, y) {x + y;}; add(1,add(2,3))", 6, nil},
	}
	for _, tc := range tests {
		got, err := stringToObject(tc.input)
		if err != nil {
			assert.NotNil(t, tc.err, "input: %v, actual: %v", tc.input, err)
			require.Conditionf(t, func() bool { return strings.Contains(err.Error(), tc.err.Error()) },
				"input: %v, expected err: %v, but got %v", tc.input, tc.err, err)
			continue
		}

		switch v := tc.expected.(type) {
		case int:
			testIntegerObject(t, tc.input, got, v)
		case bool:
			testBooleanObject(t, tc.input, got, v)
		case string:
			assert.Equal(t, tc.expected, got.Inspect(), "input: %v", tc.input)
		default:
			testNull(t, tc.input, got)
		}
	}
}

func TestEvalFunction(t *testing.T) {
	type testcase struct {
		input    string
		expected any
		err      error
	}
	tests := []testcase{
		{"fn(x, y) {return x + y;}", "fn(x,y){return (x+y);}\n", nil},
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
		case string:
			assert.Equal(t, tc.expected, got.Inspect(), "input: %v", tc.input)
		default:
			testNull(t, tc.input, got)
		}
	}
}

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
		env := object.NewEnvironment(nil)
		return Eval(ob, env)
	}

}

func testIntegerObject(t *testing.T, input string, got object.Object, expected int) {
	i, ok := got.(*object.Integer)
	assert.True(t, ok, "expected an integer object, but got: %T, input: %v", got, input)
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
