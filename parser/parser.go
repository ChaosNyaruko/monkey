package parser

import (
	"fmt"
	"strconv"

	"github.com/ChaosNyaruko/monkey/ast"
	"github.com/ChaosNyaruko/monkey/lexer"
	"github.com/ChaosNyaruko/monkey/token"
)

const (
	_ = iota
	LOWEST
	EQUALS // ==
	// false == (2 < 3)
	LESSGREATER // > <
	// 1 + (2 * 3)
	SUM
	PRODUCT
	// (-X) * Y
	PREFIX // !X -X
	CALL   // -x + (foo(1,2))
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixFnMap map[token.TokenType]prefixFn
	infixFnMap  map[token.TokenType]infixFn
}

var precedneces = map[token.TokenType]int{
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:         l,
		curToken:  token.Token{},
		peekToken: token.Token{},
	}
	p.prefixFnMap = make(map[token.TokenType]prefixFn)
	p.infixFnMap = make(map[token.TokenType]infixFn)
	// Identifer
	p.prefixFnMap[token.IDENT] = p.parseIdentifier
	p.prefixFnMap[token.INT] = p.parseIntegerLiteral
	p.prefixFnMap[token.BANG] = p.parsePrefixExpression
	p.prefixFnMap[token.MINUS] = p.parsePrefixExpression
	p.infixFnMap[token.PLUS] = p.parseInfixExpression
	p.infixFnMap[token.MINUS] = p.parseInfixExpression
	p.infixFnMap[token.ASTERISK] = p.parseInfixExpression
	p.infixFnMap[token.SLASH] = p.parseInfixExpression
	p.infixFnMap[token.LT] = p.parseInfixExpression
	p.infixFnMap[token.GT] = p.parseInfixExpression
	p.infixFnMap[token.EQ] = p.parseInfixExpression
	p.infixFnMap[token.NOT_EQ] = p.parseInfixExpression
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedneces[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedneces[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseInfixExpression(lhs ast.Expression) ast.Expression {
	// 10 + 2
	res := &ast.InfixExpression{
		Token: p.curToken,
		Lhs:   lhs,
		Op:    p.curToken.Literal,
		Rhs:   nil,
	}
	curPrecedence := p.curPrecedence() // +
	p.nextToken()
	// 10 + 2
	res.Rhs = p.parseExpression(curPrecedence)
	return res
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	res := &ast.PrefixExpression{
		Token: p.curToken,
		Op:    p.curToken.Literal,
		Rhs:   nil,
	}
	// !xx
	p.nextToken()
	// !func(y)
	res.Rhs = p.parseExpression(PREFIX)
	return res
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	num := &ast.IntegerLiteral{
		Token: p.curToken,
	}

	if v, err := strconv.ParseInt(p.curToken.Literal, 10, 64); err != nil {
		msg := fmt.Sprintf("cannot parse %q as int", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	} else {
		num.Value = int(v)
	}
	return num
}

func (p *Parser) parseIdentifier() ast.Expression {
	id := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	return id
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
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
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

func (p *Parser) parseExpression(precedence int) ast.Expression {
	fn, ok := p.prefixFnMap[p.curToken.Type]
	if !ok {
		msg := fmt.Sprintf("undefined prefix operator: %q", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lhs := fn()

	// 1 + 2; let foo = bar;
	// 1 + 2 let foo = bar;
	// 1 * 2 + 3;
	// !foo + bar - zoo
	// curToken foo
	// lhs: (!foo)
	// precedence = LOWEST
	// peek + (SUM)
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		iFn, ok := p.infixFnMap[p.peekToken.Type]
		if !ok {
			return lhs
		}
		p.nextToken()
		lhs = iFn(lhs)
	}

	return lhs
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{
		Token:      p.curToken,
		Expression: p.parseExpression(LOWEST),
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	res := &ast.ReturnStatement{
		Token:       p.curToken,
		ReturnValue: nil, // TODO
	}
	p.nextToken()

	// TODO: we don't have Expression parsing yet, so just read until a semicolon.
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return res
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{
		Token: p.curToken, // "LET"
		Name:  &ast.Identifier{},
		Value: nil,
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{
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

func (p *Parser) peekTokenIs(tok token.TokenType) bool {
	return p.peekToken.Type == tok
}

type (
	prefixFn func() ast.Expression
	infixFn  func(lhs ast.Expression) ast.Expression // res = lhs + rhs
)
