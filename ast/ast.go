// Package ast provides ...
package ast

import (
	"bytes"
	"fmt"
	"strings"

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
var _ Expression = &FunctionLiteral{}
var _ Expression = &MacroLiteral{}
var _ Expression = &CallExpression{}

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

type NullExpression struct {
	Token token.Token
	Value any
}

func (i *NullExpression) String() string {
	return i.Token.Literal
}

func (i *NullExpression) expressionNode() {}
func (i *NullExpression) TokenLiteral() string {
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

type FunctionLiteral struct {
	Token      token.Token   // "fn"
	Parameters []*Identifier // (x, y)
	Body       *BlockStatement
}

func (i *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range i.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ","))
	out.WriteString(")")
	out.WriteString(i.Body.String())
	return out.String()
}

func (i *FunctionLiteral) expressionNode() {}
func (i *FunctionLiteral) TokenLiteral() string {
	return i.Token.Literal
}

type CallExpression struct {
	Token     token.Token // "("
	Arguments []Expression
	F         Expression // the called function, add(1+2) or function literal `fn(x,y){x+y;}(1,2)`
}

func (i *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, p := range i.Arguments {
		args = append(args, p.String())
	}
	out.WriteString(i.F.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ","))
	out.WriteString(")")
	return out.String()
}

func (i *CallExpression) expressionNode() {}
func (i *CallExpression) TokenLiteral() string {
	return i.Token.Literal
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (s *StringLiteral) String() string {
	return s.Token.Literal
}

func (s *StringLiteral) expressionNode() {}
func (s *StringLiteral) TokenLiteral() string {
	return s.Token.Literal
}

type ArrayLiteral struct {
	Token    token.Token // '['
	Elements []Expression
}

func (a *ArrayLiteral) String() string {
	var out bytes.Buffer

	es := []string{}
	for _, e := range a.Elements {
		es = append(es, e.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(es, ","))
	out.WriteString("]")

	return out.String()
}

func (a *ArrayLiteral) expressionNode() {}
func (a *ArrayLiteral) TokenLiteral() string {
	return a.Token.Literal
}

type IndexExpression struct {
	Token token.Token // "["
	Left  Expression  // the expression to be indexed
	Index Expression  // the index
}

func (i *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(i.Left.String())
	out.WriteString("[")
	out.WriteString(i.Index.String())
	out.WriteString("]")
	out.WriteString(")")
	return out.String()
}

func (i *IndexExpression) expressionNode() {}
func (i *IndexExpression) TokenLiteral() string {
	return i.Token.Literal
}

type HashLiteral struct {
	Token token.Token // '{'
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}

	for k, v := range hl.Pairs {
		pairs = append(pairs, k.String()+":"+v.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ","))
	out.WriteString("}")

	return out.String()
}

func (hl *HashLiteral) expressionNode() {}
func (hl *HashLiteral) TokenLiteral() string {
	return hl.Token.Literal
}

type MacroLiteral struct {
	Token      token.Token   // "macro"
	Parameters []*Identifier // (x, y)
	Body       *BlockStatement
}

func (m *MacroLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range m.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("macro")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ","))
	out.WriteString(")")
	out.WriteString(m.Body.String())
	return out.String()
}

func (m *MacroLiteral) expressionNode() {}
func (m *MacroLiteral) TokenLiteral() string {
	return m.Token.Literal
}
