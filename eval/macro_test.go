package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ChaosNyaruko/monkey/lexer"
	"github.com/ChaosNyaruko/monkey/object"
	"github.com/ChaosNyaruko/monkey/parser"
)

func TestDefine(t *testing.T) {
	input := `
	let number = 1;
	let f = fn(x, y) { x + y };
	let m = macro(x, y) { x + y };
`
	env := object.NewEnvironment(nil)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	err := DefineMacros(program, env)
	assert.Equal(t, 2, len(program.Statements))

	_, err = env.Get("number")
	assert.NotNil(t, err)

	_, err = env.Get("f")
	assert.NotNil(t, err)

	m, err := env.Get("m")
	assert.Nil(t, err)

	mo, ok := m.(*object.Macro)
	assert.True(t, ok)

	assert.Equal(t, 2, len(mo.Parameters))
	assert.Equal(t, "x", mo.Parameters[0].String())
	assert.Equal(t, "y", mo.Parameters[1].String())
	assert.Equal(t, "(x+y)", mo.Body.String())
}

func TestExpand(t *testing.T) {
	type testcase struct {
		input    string
		expected string
	}

	for _, tc := range []testcase{
		{
			`let reverse_sub = macro(a, b) {quote(unquote(b) - unquote(a))};
			  reverse_sub(1+2, 3+4);
			`,
			`((3+4)-(1+2))`,
		},
		{
			`
			let e = macro() {quote(1+2)};
			e();
`,
			`(1+2)`,
		},
	} {
		env := object.NewEnvironment(nil)
		l := lexer.New(tc.input)
		p := parser.New(l)
		program := p.ParseProgram()

		err := DefineMacros(program, env)
		assert.Nil(t, err)
		res := ExpandMacros(program, env)
		assert.Nil(t, err)

		assert.Equal(t, tc.expected, res.String())
	}
}
