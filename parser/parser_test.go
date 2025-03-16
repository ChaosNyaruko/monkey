package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ChaosNyaruko/monkey/ast"
	"github.com/ChaosNyaruko/monkey/lexer"
	"github.com/stretchr/testify/assert"
)

func TestLetStatments(t *testing.T) {
	input := `
	let x = 5;
	let y = 10;
	let foobar = 383838;
`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 3 {
		t.Fatalf("expected 3 statements, but got %v", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("should be a let, but got: %v", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s should be a LetStatement, but got %T", s)
		return false
	}

	// check left side
	if letStmt.Name.Value != name {
		t.Errorf("left name.Value should be %q, but got %q", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("left name.TokenLiteral should be %q, but got %q", name, letStmt.Name.TokenLiteral())
		return false
	}

	// TODO: the right side is an expression, we haven't implement expression parsing yet.

	return true

}

func checkParserErrors(t *testing.T, p *Parser) {
	errs := p.Errors()
	if len(errs) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errs))

	for _, msg := range errs {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestReturnStatements(t *testing.T) {
	input := `
	return 5;
	return 10;
	return 993 322;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements doen't contain 3 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not a return statement. got=%T", stmt)
			continue
		}

		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt TokenLiteral not 'return'. got=%q", returnStmt.TokenLiteral())
		}

	}

}

func TestIdentifierExpression(t *testing.T) {
	input := `
	foobar;
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(t, 1, len(program.Statements),
		"program.Statements doesn't contain proper statements, %s", program)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok, "should be an expression statement")
	assert.Equal(t, "foobar", stmt.String())
}

func TestIntegerLiteral(t *testing.T) {
	input := `
	5;
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(t, 1, len(program.Statements),
		"program.Statements doesn't contain proper statements, %s", program)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok, "should be an expression statement")
	assert.Equal(t, "5", stmt.String())

	assert.Equal(t, 5, stmt.Expression.(*ast.IntegerLiteral).Value)
}

func TestPrefixExpressions(t *testing.T) {
	type testcase struct {
		input string
		op    string
		value any
	}

	cases := []testcase{
		{"!5;", "!", 5},
		{"-10;", "-", 10},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}
	for _, x := range cases {
		l := lexer.New(x.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement,  but got %T", program.Statements[0])
		preStmt, ok := stmt.Expression.(*ast.PrefixExpression)
		assert.True(t, ok, "should be an prefix expression, but got %T", stmt.Expression)
		assert.Equal(t, x.op, preStmt.Op, "parse op error")
		t.Logf("input: %v, expression: %v", x.input, preStmt.String())
		testLiteralExpression(t, preStmt.Rhs, x.value)
	}

}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected any) {
	switch v := expected.(type) {
	case int:
		testIntegerLiteral(t, exp, v)
	case int64:
		testIntegerLiteral(t, exp, int(v))
	case string:
		testIdentifier(t, exp, v)
	case bool:
		testBoolean(t, exp, v)
	default:
		t.Errorf("testLiteralExpression: unsupported type %T", expected)
	}
}

func testInfixExpression(t *testing.T, exp ast.Expression, left any, op string, right any) {
	t.Logf("testing infix expression: %v, expected: %v %v %v", exp, left, op, right)
	e, ok := exp.(*ast.InfixExpression)
	assert.True(t, ok, "should be an infix expression")
	testLiteralExpression(t, e.Lhs, left)
	assert.Equal(t, op, e.Op)
	testLiteralExpression(t, e.Rhs, right)
}

func testBoolean(t *testing.T, ep ast.Expression, value bool) {
	ident, ok := ep.(*ast.BooleanExpression)
	assert.True(t, ok, "should be a boolean, but got %T", ep)
	assert.Equal(t, value, ident.Value)
}

func testIdentifier(t *testing.T, ep ast.Expression, value string) {
	ident, ok := ep.(*ast.Identifier)
	assert.True(t, ok, "should be an identifier, but got %T", ep)
	assert.Equal(t, value, ident.Value)
	assert.Equal(t, value, ident.TokenLiteral())
}

func testIntegerLiteral(t *testing.T, ep ast.Expression, value int) {
	in, ok := ep.(*ast.IntegerLiteral)
	assert.True(t, ok, "should be an integer literal, but got %T", ep)
	assert.Equal(t, fmt.Sprintf("%d", value), in.String())
	assert.Equal(t, value, in.Value)
}

func TestInfixExpressions(t *testing.T) {
	type testcase struct {
		input string
		left  any
		op    string
		right any
	}

	cases := []testcase{
		{"10+2;", 10, "+", 2},
		{"1-10;", 1, "-", 10},
		{"5*10;", 5, "*", 10},
		{"10/5;", 10, "/", 5},
		{"10>5;", 10, ">", 5},
		{"10<5;", 10, "<", 5},
		{"10==5;", 10, "==", 5},
		{"10!=5;", 10, "!=", 5},
		{"alice*bob;", "alice", "*", "bob"},
	}
	for _, x := range cases {
		l := lexer.New(x.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement,  but got %T, %s",
			program.Statements[0], program.Statements[0].String())
		testInfixExpression(t, stmt.Expression, x.left, x.op, x.right)
	}
}

func TestOpPrecedence(t *testing.T) {
	type testcase struct {
		input    string
		expected string
	}

	cases := []testcase{
		{"a + b * c", "(a+(b*c))"},
		{"a + b / c", "(a+(b/c))"},
		{"a / b + c", "((a/b)+c)"},
		{"a * b + c", "((a*b)+c)"},
		{"-a * b", "((-a)*b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a+b)+c)"},
		{"a + b - c", "((a+b)-c)"},
		{"a * b * c", "((a*b)*c)"},
		{"a * b / c", "((a*b)/c)"},
		{"a + b + c * d /f - !e * g", "(((a+b)+((c*d)/f))-((!e)*g))"},
		{"a < b > c == d != e", "((((a<b)>c)==d)!=e)"},
		{"a + b; b / c", "(a+b)(b/c)"},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 == false", "((3>5)==false)"},
		{"3 < 5 == true", "((3<5)==true)"},
		{"true == 3 < 5", "(true==(3<5))"},
		{"1 + (2 + 3) + 4", "((1+(2+3))+4)"},
		{"2/(5+5)", "(2/(5+5))"},
		{"(5+5)*2", "((5+5)*2)"},
		{"-(5+5)", "(-(5+5))"},
		{"!(true==true)", "(!(true==true))"},
	}
	for _, x := range cases {
		l := lexer.New(x.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		assert.Equal(t, x.expected, program.String())
	}
}

func TestBooleanExpression(t *testing.T) {
	type testcase struct {
		input    string
		expected bool
	}

	for _, tc := range []testcase{
		{"true;", true},
		{"false;", false},
	} {

		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement")
		testLiteralExpression(t, stmt.Expression, tc.expected)
	}
}

func TestIfExpression(t *testing.T) {
	type testcase struct {
		input    string
		expected any
	}

	for _, tc := range []testcase{
		{"if (x < y) {x}", "TODO"},
	} {

		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement")
		exp, ok := stmt.Expression.(*ast.IfExpression)
		assert.True(t, ok, "should be an if expression statement")
		testInfixExpression(t, exp.Condition, "x", "<", "y")

		b1 := exp.If.Statements
		assert.Equal(t, 1, len(b1), "bad if branch")
		s1, ok := b1[0].(*ast.ExpressionStatement) // x
		assert.True(t, ok, "should be an expression statement statement")
		testIdentifier(t, s1.Expression, "x")
	}
}

func TestIfElseExpression(t *testing.T) {
	type testcase struct {
		input    string
		expected any
	}

	for _, tc := range []testcase{
		{"if (x < y) {x} else {y}", "TODO"},
	} {

		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement")
		exp, ok := stmt.Expression.(*ast.IfExpression)
		assert.True(t, ok, "should be an if expression statement")
		testInfixExpression(t, exp.Condition, "x", "<", "y")

		b1 := exp.If.Statements
		assert.Equal(t, 1, len(b1), "bad if branch")
		s1, ok := b1[0].(*ast.ExpressionStatement) // x
		assert.True(t, ok, "should be an expression statement statement")
		testIdentifier(t, s1.Expression, "x")

		assert.NotNil(t, exp.Else)
		b2 := exp.Else.Statements
		assert.Equal(t, 1, len(b2), "bad else branch")
		s2, ok := b2[0].(*ast.ExpressionStatement) // y
		assert.True(t, ok, "should be an expression statement statement")
		testIdentifier(t, s2.Expression, "y")
	}
}

func TestFunctionLiteral(t *testing.T) {
	type testcase struct {
		input    string
		expected any
	}

	for _, tc := range []testcase{
		{"fn(x,y){x+y;}", "TODO"},
	} {

		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement")
		exp, ok := stmt.Expression.(*ast.FunctionLiteral)
		assert.True(t, ok, "should be a function literal expression statement")

		testLiteralExpression(t, exp.Parameters[0], "x")
		testLiteralExpression(t, exp.Parameters[1], "y")

		body := exp.Body
		assert.Equal(t, 1, len(body.Statements))
		testInfixExpression(t, body.Statements[0].(*ast.ExpressionStatement).Expression, "x", "+", "y")
	}
}

func TestFunctionParameters(t *testing.T) {
	type testcase struct {
		input    string
		expected []string
	}

	for _, tc := range []testcase{
		{"fn(x,y){x+y;}", []string{"x", "y"}},
		{"fn(){}", nil},
		{"fn(x){2 * x}", []string{"x"}},
		{"fn(x,y,z){x+y*z;}", []string{"x", "y", "z"}},
	} {

		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement")
		exp, ok := stmt.Expression.(*ast.FunctionLiteral)
		assert.True(t, ok, "should be a function literal expression statement")

		var ps []string
		for _, p := range exp.Parameters {
			ps = append(ps, p.String())
		}
		t.Logf("ps is %v", ps)
		assert.Equal(t, strings.Join(tc.expected, ","), strings.Join(ps, ","))
	}
}

func TestCallFunction(t *testing.T) {
	type testcase struct {
		input    string
		expected string
	}
	for _, tc := range []testcase{
		{"add(1, 2+3, 4 + 5*6, 7*8+10)", "add(1,(2+3),(4+(5*6)),((7*8)+10))"},
		{"non()", "non()"},
		{"negate(1)", "negate(1)"},
	} {

		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement")
		exp, ok := stmt.Expression.(*ast.CallExpression)
		assert.True(t, ok, "should be a function literal expression statement")

		assert.Equal(t, tc.expected, exp.String())
	}
}
