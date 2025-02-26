package parser

import (
	"testing"

	"github.com/ChaosNyaruko/monkey/ast"
	"github.com/ChaosNyaruko/monkey/lexer"
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
