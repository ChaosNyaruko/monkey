package parser

import (
	"fmt"
	"log"
	"os"

	"github.com/ChaosNyaruko/monkey/ast"
	"github.com/ChaosNyaruko/monkey/lexer"
	"github.com/ChaosNyaruko/monkey/token"
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:         l,
		curToken:  token.Token{},
		peekToken: token.Token{},
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	default:
		log.Printf("unsupported type: %v", p.curToken.Type)
		return nil
	}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, but got %s", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	fmt.Fprintf(os.Stderr, "parse 'let' stmt")
	stmt := &ast.LetStatement{
		Token: p.curToken, // "LET"
		Name:  &ast.Identifer{},
		Value: nil,
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifer{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: skip for now, so just read it until a semicolon encountered.
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) curTokenIs(tok token.TokenType) bool {
	return p.curToken.Type == tok
}
