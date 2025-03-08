package parser

import (
	"fmt"
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
		value int
	}

	cases := []testcase{
		{"!5;", "!", 5},
		{"-10;", "-", 10},
		// {"-!10", "-", 10},
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
		testIntegerLiteral(t, preStmt.Rhs, x.value)
	}

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
		left  int
		op    string
		right int
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
		infixStmt, ok := stmt.Expression.(*ast.InfixExpression)
		assert.True(t, ok, "should be an infix expression, but got %T, %s",
			stmt.Expression, stmt.Expression.String())
		assert.Equal(t, x.op, infixStmt.Op, "parse op error")
		t.Logf("input: %v, expression: %v", x.input, infixStmt.String())
		testIntegerLiteral(t, infixStmt.Lhs, x.left)
		testIntegerLiteral(t, infixStmt.Rhs, x.right)
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
	}
	for _, x := range cases {
		l := lexer.New(x.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		assert.Equal(t, x.expected, program.String())
	}
}
