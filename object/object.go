package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ChaosNyaruko/monkey/ast"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	FUNCTION_OBJ     = "FUNCTION"
)

var _ Object = &Integer{}
var _ Object = &Boolean{}
var _ Object = &Null{}
var _ Object = &ReturnValue{}
var _ Object = &Function{}

type Object interface {
	Inspect() string
	Type() ObjectType
}

type Integer struct {
	Value int
}

func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}

type ReturnValue struct {
	Value Object
}

func (i *ReturnValue) Inspect() string {
	return i.Value.Inspect()
}

func (i *ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%v", b.Value)
}

func (b *Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}

type Null struct {
}

func (n *Null) Inspect() string {
	return "null"
}

func (b *Null) Type() ObjectType {
	return NULL_OBJ
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ","))
	out.WriteString(")")
	out.WriteString("{")
	out.WriteString(f.Body.String())
	out.WriteString("}\n")

	return out.String()
}

func (f *Function) Type() ObjectType {
	return FUNCTION_OBJ
}
