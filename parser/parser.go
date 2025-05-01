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
	INDEX
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixFnMap map[token.TokenType]prefixFn
	infixFnMap  map[token.TokenType]infixFn
}

var precedences = map[token.TokenType]int{
	token.PLUS:     SUM,
	token.LBRACKET: INDEX,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LPAREN:   CALL,
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
	p.prefixFnMap[token.TRUE] = p.parseBoolean
	p.prefixFnMap[token.FALSE] = p.parseBoolean
	p.prefixFnMap[token.NULL] = p.parseNull
	p.prefixFnMap[token.LPAREN] = p.parseGroupingExpression
	p.prefixFnMap[token.IF] = p.parseIfElseExpression
	p.prefixFnMap[token.FUNCTION] = p.parseFunctionLiteral
	p.prefixFnMap[token.MACRO] = p.parseMacroLiteral
	p.prefixFnMap[token.STRING] = p.parseStringLiteral
	p.prefixFnMap[token.LBRACKET] = p.parseArrayLiteral
	p.prefixFnMap[token.LBRACE] = p.parseHashLiteral
	p.infixFnMap[token.PLUS] = p.parseInfixExpression
	p.infixFnMap[token.MINUS] = p.parseInfixExpression
	p.infixFnMap[token.ASTERISK] = p.parseInfixExpression
	p.infixFnMap[token.SLASH] = p.parseInfixExpression
	p.infixFnMap[token.LT] = p.parseInfixExpression
	p.infixFnMap[token.GT] = p.parseInfixExpression
	p.infixFnMap[token.EQ] = p.parseInfixExpression
	p.infixFnMap[token.NOT_EQ] = p.parseInfixExpression
	p.infixFnMap[token.LPAREN] = p.parseInfixExpression
	p.infixFnMap[token.LBRACKET] = p.parseInfixExpression
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// parseFunctionCall is a special case for parseInfixExpression
func (p *Parser) parseFunctionCall(f ast.Expression) ast.Expression {
	fc := &ast.CallExpression{
		Token:     p.curToken, // must be "("
		Arguments: nil,
		F:         f,
	}
	fc.Arguments = p.parseExpressionList(token.RPAREN)
	return fc
}

func (p *Parser) parseIndexExpression(lhs ast.Expression) ast.Expression {
	// myArray[1+2]
	i := &ast.IndexExpression{
		Token: p.curToken,
		Left:  lhs,
		Index: nil,
	}
	p.nextToken()
	i.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		panic("indexing[ is not closed")
	}
	return i
}

func (p *Parser) parseInfixExpression(lhs ast.Expression) ast.Expression {
	// -add.(10 + 2)
	// lhs = add
	if p.curTokenIs(token.LPAREN) { // function calling
		return p.parseFunctionCall(lhs)
	}
	if p.curTokenIs(token.LBRACKET) { // indexing
		return p.parseIndexExpression(lhs)
	}
	res := &ast.InfixExpression{
		Token: p.curToken,
		Lhs:   lhs,
		Op:    p.curToken.Literal,
		Rhs:   nil,
	}
	curPrecedence := p.curPrecedence()
	p.nextToken()
	res.Rhs = p.parseExpression(curPrecedence)
	return res
}

func (p *Parser) parseIfElseExpression() ast.Expression {
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	res := &ast.IfExpression{}
	// parse condition
	res.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	// parse branch1
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	res.If = p.parseBlockStatement()

	// parse else branch
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		// else {}
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		res.Else = p.parseBlockStatement()
	}

	return res
}

// parseBlockStatement will eat the "}" inside.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{
		Token:      p.curToken, // "{ "
		Statements: []ast.Statement{},
	}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseGroupingExpression() ast.Expression {
	//  (expression)

	// read "("
	p.nextToken()

	grouped := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return grouped
}

func (p *Parser) parseNull() ast.Expression {
	res := &ast.NullExpression{
		Token: p.curToken,
	}
	return res
}

func (p *Parser) parseBoolean() ast.Expression {
	res := &ast.BooleanExpression{
		Token: p.curToken,
	}
	res.Value = p.curTokenIs(token.TRUE)
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

func (p *Parser) Error() error {
	if len(p.errors) == 0 {
		return nil
	}

	var s string
	for _, msg := range p.errors {
		s += "\t" + msg + "\n"
	}
	return fmt.Errorf("parser error: %v", s)
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
		msg := fmt.Sprintf("undefined prefix operator: %q", p.curToken.Type)
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
		ReturnValue: nil,
	}
	p.nextToken()

	res.ReturnValue = p.parseExpression(LOWEST)
	if !p.expectPeek(token.SEMICOLON) {
		return nil
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

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)
	if !p.expectPeek(token.SEMICOLON) {
		return nil
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

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	if p.peekTokenIs(token.RPAREN) {
		// no params
		p.nextToken()
		return nil
	}

	p.nextToken() // move to the first identifier
	id := p.parseIdentifier().(*ast.Identifier)
	ids := []*ast.Identifier{id}
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		id := p.parseIdentifier().(*ast.Identifier)
		ids = append(ids, id)
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return ids
}

func (p *Parser) parseHashLiteral() ast.Expression {
	h := &ast.HashLiteral{
		Token: p.curToken,
		Pairs: make(map[ast.Expression]ast.Expression),
	}
	// parse key-value pairs
	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken() // eat "{" or ","
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			p.errors = append(p.errors, "no ':' after a key in hashmap")
			return nil
		}
		p.nextToken() // eat ":"
		if value := p.parseExpression(LOWEST); value != nil {
			h.Pairs[key] = value
		} else {
			p.errors = append(p.errors, fmt.Sprintf("parse value for '%q' error", key.String()))
			return nil
		}

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			p.errors = append(p.errors, fmt.Sprintf("expected ',' or '}', got: %q", p.peekToken.Literal))
			return nil
		}
	}
	// confirm and eat "}"
	if !p.expectPeek(token.RBRACE) {
		p.errors = append(p.errors, fmt.Sprintf("expected '}', got: %q", p.peekToken.Literal))
		return nil
	}
	return h
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	a := &ast.ArrayLiteral{
		Token:    p.curToken,
		Elements: []ast.Expression{},
	}
	a.Elements = p.parseExpressionList(token.RBRACKET)
	return a
}

func (p *Parser) parseExpressionList(close token.TokenType) []ast.Expression {
	if p.peekTokenIs(close) {
		// no params
		p.nextToken()
		return []ast.Expression{}
	}

	p.nextToken() // move to the first Expression
	arg := p.parseExpression(LOWEST)
	args := []ast.Expression{arg}
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		arg := p.parseExpression(LOWEST)
		args = append(args, arg)
	}
	if !p.expectPeek(close) {
		return nil
	}

	return args
}

func (p *Parser) parseStringLiteral() ast.Expression {
	raw := p.curToken.Literal
	// TODO: check the unclosed quotes.
	return &ast.StringLiteral{
		Token: p.curToken,
		Value: raw,
	}
}

func (p *Parser) parseMacroLiteral() ast.Expression {
	f := &ast.MacroLiteral{
		Token: p.curToken,
		Body:  &ast.BlockStatement{},
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	f.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	f.Body = p.parseBlockStatement()

	if !p.curTokenIs(token.RBRACE) {
		panic("the { is not closed for macro")
	}

	return f
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	f := &ast.FunctionLiteral{
		Token: p.curToken,
		Body:  &ast.BlockStatement{},
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	f.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	f.Body = p.parseBlockStatement()

	if !p.curTokenIs(token.RBRACE) {
		panic("the { is not closed")
	}

	return f
}
