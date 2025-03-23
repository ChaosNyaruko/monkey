package object

import (
	"fmt"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
)

var _ Object = &Integer{}
var _ Object = &Boolean{}
var _ Object = &Null{}
var _ Object = &ReturnValue{}

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
