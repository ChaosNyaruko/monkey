// Package ast provides ...
package ast

import (
	"github.com/ChaosNyaruko/monkey/token"
)

var _ Statement = &LetStatement{}
var _ Expression = &Identifer{}

type Node interface {
	TokenLiteral() string
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

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()

	}
	return ""
}

type Identifer struct {
	Token token.Token // IDENT
	Value string
}

func (i *Identifer) expressionNode() {}
func (i *Identifer) TokenLiteral() string {
	return i.Token.Literal
}

type LetStatement struct {
	Token token.Token // LET
	// let name = value
	// let x = y;
	Name  *Identifer
	Value Expression
}

func (ls *LetStatement) statementNode() {}

func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}
