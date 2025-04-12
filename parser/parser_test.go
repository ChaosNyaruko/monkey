package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ChaosNyaruko/monkey/ast"
	"github.com/ChaosNyaruko/monkey/lexer"
	"github.com/ChaosNyaruko/monkey/token"
	"github.com/stretchr/testify/assert"
)

func TestHashLiteral(t *testing.T) {
	tests := []struct {
		input string
		empty bool // expected
	}{
		{`{"foo": "bar", 2: true, "2": false, true: "TRUE"}`, false},
		{`{}`, true},
	}

	for _, tc := range tests {
		l := lexer.New(tc.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p, tc.input)
		stmt := program.Statements[0].(*ast.ExpressionStatement)
		s, ok := stmt.Expression.(*ast.HashLiteral)
		assert.True(t, ok, "should be a hash literal expression, but got %T for input: %v", s, tc.input)
		if tc.empty {
			assert.Equal(t, 0, len(s.Pairs))
			return
		}
		for k, v := range s.Pairs {
			switch kv := k.(type) {
			case *ast.StringLiteral:
				if kv.Value == "foo" {
					assert.Equal(t, "bar", v.(*ast.StringLiteral).Value)
				} else if kv.Value == "2" {
					assert.Equal(t, false, v.(*ast.BooleanExpression).Value)
				}
			case *ast.IntegerLiteral:
				if kv.Value == 2 {
					assert.Equal(t, true, v.(*ast.BooleanExpression).Value)
				}
			case *ast.BooleanExpression:
				assert.True(t, kv.Value)
				assert.Equal(t, "TRUE", v.(*ast.StringLiteral).Value)
			}
		}
	}
}

func TestIndexExpression(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"myArray[1+2]"},
	}

	for _, tc := range tests {
		l := lexer.New(tc.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p, tc.input)
		stmt := program.Statements[0].(*ast.ExpressionStatement)
		s, ok := stmt.Expression.(*ast.IndexExpression)
		assert.True(t, ok, "should be an index expression, but got %T for input: %v", s, tc.input)
		testIdentifier(t, s.Left, "myArray")
		testInfixExpression(t, s.Index, 1, "+", 2)
	}
}

func TestArrayLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"[]", []int{}},
		{"[1,6]", []int{1, 6}},
	}

	for _, tc := range tests {
		l := lexer.New(tc.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p, tc.input)
		stmt := program.Statements[0].(*ast.ExpressionStatement)
		s, ok := stmt.Expression.(*ast.ArrayLiteral)
		assert.True(t, ok, "should be an array literal, but got %T for input: %v", stmt, tc.input)
		assert.Equal(t, len(tc.expected), len(s.Elements))
		for i, e := range s.Elements {
			testExpression(t, e, tc.expected[i])
		}
		// testIntegerLiteral(t, s.Elements[0], 1)
		// testInfixExpression(t, s.Elements[1], 2, "*", 3)
	}
}

func TestStringLiteral(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue string
	}{
		{`"hello world"`, `hello world`},
	}

	for _, tc := range tests {
		l := lexer.New(tc.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p, tc.input)
		stmt := program.Statements[0].(*ast.ExpressionStatement)
		s, ok := stmt.Expression.(*ast.StringLiteral)
		assert.True(t, ok, "should be a string literal, but got %T for input: %v", stmt, tc.input)
		assert.Equal(t, tc.expectedValue, s.Value)
	}
}

func TestLetStatments(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedIdentifier string
		expectedValue      any
	}{
		{"1", "let x = 5;", "x", 5},
		{"1", "let y = 10;", "y", 10},
		{"1", "let foobar = 383838;", "foobar", 383838},
		{"malformed", "let foo = 1 + 2", "foo", nil},
		{"1", "let foo = 1 + 2;", "foo", &ast.InfixExpression{
			Token: token.Token{
				Type:    token.PLUS,
				Literal: "1",
			},
			Lhs: &ast.IntegerLiteral{
				Token: token.Token{
					Type:    token.INT,
					Literal: "1",
				},
				Value: 1,
			},
			Op: "+",
			Rhs: &ast.IntegerLiteral{
				Token: token.Token{
					Type:    token.INT,
					Literal: "2",
				},
				Value: 2,
			},
		}},
	}

	for _, tc := range tests {
		l := lexer.New(tc.input)
		p := New(l)

		program := p.ParseProgram()
		if tc.name == "malformed" {
			assert.NotZero(t, p.Errors())
			continue
		} else {
			checkParserErrors(t, p, tc.input)
		}
		if fmt.Sprintf("%d", len(program.Statements)) != tc.name {
			t.Fatalf("program.Statements doen't contain %s statements. got=%d", tc.name, len(program.Statements))
		}
		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tc.expectedIdentifier, tc.expectedValue) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string, value any) bool {
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

	testExpression(t, letStmt.Value, value)

	return true

}

func checkParserErrors(t *testing.T, p *Parser, input string) {
	errs := p.Errors()
	if len(errs) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errs))

	for _, msg := range errs {
		t.Errorf("parser error: %q, input: %v", msg, input)
	}
	t.FailNow()
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValue any
	}{
		{"1", "return 5;", 5},
		{"1", "return 10;", 10},
		{"2", "return 993;322", 993},
		{"malformed", "return 993 322", nil},
		{"1", "return 1 + 2;", &ast.InfixExpression{
			Token: token.Token{
				Type:    token.PLUS,
				Literal: "1",
			},
			Lhs: &ast.IntegerLiteral{
				Token: token.Token{
					Type:    token.INT,
					Literal: "1",
				},
				Value: 1,
			},
			Op: "+",
			Rhs: &ast.IntegerLiteral{
				Token: token.Token{
					Type:    token.INT,
					Literal: "2",
				},
				Value: 2,
			},
		}},
	}

	for _, tc := range tests {
		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		if tc.name == "malformed" {
			assert.NotZero(t, p.Errors())
			continue
		} else {
			checkParserErrors(t, p, tc.input)
		}

		if fmt.Sprintf("%d", len(program.Statements)) != tc.name {
			t.Fatalf("program.Statements doen't contain %s statements. got=%d", tc.name, len(program.Statements))
		}

		stmt := program.Statements[0]
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not a return statement. got=%T", stmt)
			continue
		}

		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt TokenLiteral not 'return'. got=%q", returnStmt.TokenLiteral())
		}

		testExpression(t, returnStmt.ReturnValue, tc.expectedValue)
	}

}

func TestIdentifierExpression(t *testing.T) {
	input := `
	foobar;
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p, input)

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
	checkParserErrors(t, p, input)

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
		checkParserErrors(t, p, x.input)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement,  but got %T", program.Statements[0])
		preStmt, ok := stmt.Expression.(*ast.PrefixExpression)
		assert.True(t, ok, "should be an prefix expression, but got %T", stmt.Expression)
		assert.Equal(t, x.op, preStmt.Op, "parse op error")
		t.Logf("input: %v, expression: %v", x.input, preStmt.String())
		testExpression(t, preStmt.Rhs, x.value)
	}

}

func testExpression(t *testing.T, exp ast.Expression, expected any) {
	switch v := expected.(type) {
	case int:
		testIntegerLiteral(t, exp, v)
	case int64:
		testIntegerLiteral(t, exp, int(v))
	case string:
		testIdentifier(t, exp, v)
	case bool:
		testBoolean(t, exp, v)
	case *ast.InfixExpression:
		testInfixExpression(t, exp, v.Lhs, v.Op, v.Rhs)
	case *ast.IntegerLiteral:
		testIntegerLiteral(t, exp, v.Value)
	default:
		t.Errorf("testLiteralExpression: unsupported type %T", expected)
	}
}

func testInfixExpression(t *testing.T, exp ast.Expression, left any, op string, right any) {
	t.Logf("testing infix expression: %v, expected: %v %v %v", exp, left, op, right)
	e, ok := exp.(*ast.InfixExpression)
	assert.True(t, ok, "should be an infix expression")
	testExpression(t, e.Lhs, left)
	assert.Equal(t, op, e.Op)
	testExpression(t, e.Rhs, right)
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
		checkParserErrors(t, p, x.input)

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
		{"a*[1,2,3,4][b*c] * d", "((a*([1,2,3,4][(b*c)]))*d)"},
		{"add(a*b[2], b[1], 2*[1,2][0])", "add((a*(b[2])),(b[1]),(2*([1,2][0])))"},
	}
	for _, x := range cases {
		l := lexer.New(x.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, x.input)
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
		checkParserErrors(t, p, tc.input)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement")
		testExpression(t, stmt.Expression, tc.expected)
	}
}

func TestIfExpression(t *testing.T) {
	type testcase struct {
		input    string
		expected any
	}

	for _, tc := range []testcase{
		{"if (x < y) {x}", ""},
	} {

		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tc.input)

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
		{"if (x < y) {x} else {y}", ""},
	} {

		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tc.input)

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
		{"fn(x,y){x+y;}", ""},
	} {

		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tc.input)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement")
		exp, ok := stmt.Expression.(*ast.FunctionLiteral)
		assert.True(t, ok, "should be a function literal expression statement")

		testExpression(t, exp.Parameters[0], "x")
		testExpression(t, exp.Parameters[1], "y")

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
		checkParserErrors(t, p, tc.input)

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
		checkParserErrors(t, p, tc.input)

		assert.Equal(t, 1, len(program.Statements),
			"program.Statements doesn't contain proper statements, %s", program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "should be an expression statement")
		exp, ok := stmt.Expression.(*ast.CallExpression)
		assert.True(t, ok, "should be a function literal expression statement")

		assert.Equal(t, tc.expected, exp.String())
	}
}
