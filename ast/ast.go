// Package ast provides ...
package ast

import (
	"bytes"
	"fmt"

	"github.com/ChaosNyaruko/monkey/token"
)

var _ Statement = &LetStatement{}
var _ Statement = &BlockStatement{}
var _ Expression = &Identifier{}
var _ Expression = &IntegerLiteral{}
var _ Expression = &PrefixExpression{}
var _ Expression = &InfixExpression{}
var _ Expression = &BooleanExpression{}
var _ Expression = &IfExpression{}

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()

	}
	return ""
}

type Identifier struct {
	Token token.Token // IDENT
	Value string      // the "Name" of the Identifer, x/y/z
}

func (i *Identifier) String() string {
	return i.Value
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

type IntegerLiteral struct {
	Token token.Token // {INT, "5"}
	Value int         // "5" -> 5
}

func (i *IntegerLiteral) String() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *IntegerLiteral) expressionNode() {}
func (i *IntegerLiteral) TokenLiteral() string {
	return i.Token.Literal
}

type BooleanExpression struct {
	Token token.Token
	Value bool
}

func (i *BooleanExpression) String() string {
	return i.Token.Literal
}

func (i *BooleanExpression) expressionNode() {}
func (i *BooleanExpression) TokenLiteral() string {
	return i.Token.Literal
}

type InfixExpression struct {
	Token token.Token // +-/*<> == !=
	Lhs   Expression
	Op    string
	Rhs   Expression
}

func (i *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(i.Lhs.String())
	out.WriteString(i.Op)
	out.WriteString(i.Rhs.String())
	out.WriteString(")")

	return out.String()
}

func (i *InfixExpression) expressionNode() {}
func (i *InfixExpression) TokenLiteral() string {
	return i.Token.Literal
}

type PrefixExpression struct {
	Token token.Token // {BANG/MINUS, "!"/"-"}
	Op    string
	Rhs   Expression
}

func (i *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(i.Op)
	out.WriteString(i.Rhs.String())
	out.WriteString(")")

	return out.String()
}

func (i *PrefixExpression) expressionNode() {}
func (i *PrefixExpression) TokenLiteral() string {
	return i.Token.Literal
}

type LetStatement struct {
	Token token.Token // LET
	// let name = value
	// let x = y;
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) String() string {
	// "let Name = Value;"
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")

	return out.String()
}

func (ls *LetStatement) statementNode() {}

func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) String() string {
	// return <ReturnValue>;
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")
	out.WriteString(rs.ReturnValue.String())
	out.WriteString(";")
	return out.String()
}

func (rs *ReturnStatement) statementNode() {}

func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

type ExpressionStatement struct {
	Token      token.Token // The first token of the expression.
	Expression Expression
}

func (es *ExpressionStatement) String() string {
	return es.Expression.String()
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}

type IfExpression struct {
	Token     token.Token // "if"
	Condition Expression
	If        *BlockStatement
	Else      *BlockStatement
}

func (i *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(i.Condition.String())
	out.WriteString(" ")
	out.WriteString(i.If.String())

	if i.Else != nil {
		out.WriteString("else ")
		out.WriteString(i.Else.String())
	}

	return out.String()
}

func (i *IfExpression) expressionNode() {}
func (i *IfExpression) TokenLiteral() string {
	return i.Token.Literal
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (b *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range b.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

func (i *BlockStatement) statementNode() {}
func (i *BlockStatement) TokenLiteral() string {
	return i.Token.Literal
}
