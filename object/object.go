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
	STRING_OBJ       = "STRING"
	BOOLEAN_OBJ      = "BOOLEAN"
	ARRAY_OBJ        = "ARRAY"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	FUNCTION_OBJ     = "FUNCTION"
	BUILTIN_OBJ      = "BUILTIN"
)

var _ Object = &Integer{}
var _ Object = &Boolean{}
var _ Object = &Null{}
var _ Object = &ReturnValue{}
var _ Object = &Function{}
var _ Object = &String{}
var _ Object = &Array{}

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

type String struct {
	Value string
}

func (s *String) Inspect() string {
	return s.Value
}

func (s *String) Type() ObjectType {
	return STRING_OBJ
}

type BuiltinFunction func(args ...Object) (Object, error)

type Builtin struct {
	Name string
	Fn   BuiltinFunction
}

func (b *Builtin) Inspect() string {
	return fmt.Sprintf("%v is a builtin\n", b.Name)
}

func (b *Builtin) Type() ObjectType {
	return BUILTIN_OBJ
}

type Array struct {
	Elements []Object
}

func (a *Array) Inspect() string {
	var out bytes.Buffer
	es := []string{}
	for _, e := range a.Elements {
		es = append(es, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(es, ","))
	out.WriteString("]")
	return out.String()
}

func (a *Array) Type() ObjectType {
	return ARRAY_OBJ
}
