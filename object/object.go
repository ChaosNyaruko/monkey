package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/ChaosNyaruko/monkey/ast"
)

type ObjectType string

type Hashable interface {
	Object
	HashKey() HashKey
	// Equals(Object) bool TODO: deal with hash conflicts.
}

const (
	INTEGER_OBJ      = "INTEGER"
	STRING_OBJ       = "STRING"
	HASH_OBJ         = "HASH"
	BOOLEAN_OBJ      = "BOOLEAN"
	ARRAY_OBJ        = "ARRAY"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	FUNCTION_OBJ     = "FUNCTION"
	BUILTIN_OBJ      = "BUILTIN"
	QUOTE_OBJ        = "QUOTE"
)

var _ Hashable = &Integer{}
var _ Hashable = &Boolean{}
var _ Hashable = &Integer{}

var _ Object = &Integer{}
var _ Object = &Boolean{}
var _ Object = &Null{}
var _ Object = &ReturnValue{}
var _ Object = &Function{}
var _ Object = &String{}
var _ Object = &Array{}
var _ Object = &Quote{}

type Object interface {
	Inspect() string
	Type() ObjectType
}

type Integer struct {
	Value int
}

func (i *Integer) HashKey() HashKey {
	return HashKey{
		Type: i.Type(),
		Key:  uint64(i.Value),
	}
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

func (b *Boolean) HashKey() HashKey {
	var key uint64 = 0
	if b.Value {
		key = 1
	}
	return HashKey{
		Type: b.Type(),
		Key:  key,
	}
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

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	_, err := h.Write([]byte(s.Value))
	if err != nil {
		panic(err)
	}
	return HashKey{
		Type: s.Type(),
		Key:  h.Sum64(),
	}
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

type HashKey struct {
	Type ObjectType
	Key  uint64
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	// no map[Object]Object -> {"key": "xx", "key": yy}
	// no map[string]Object -> {2: "a", "2": "b"}
	Pairs map[HashKey]HashPair
}

func (h *Hash) Inspect() string {
	var out bytes.Buffer
	kvs := []string{}
	for _, v := range h.Pairs {
		kvs = append(kvs, fmt.Sprintf("%s:%s", v.Key.Inspect(), v.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(kvs, ","))
	out.WriteString("}")
	return out.String()
}

func (h *Hash) Type() ObjectType {
	return HASH_OBJ
}

type Quote struct {
	Node ast.Node
}

func (q *Quote) Inspect() string {
	return "QUOTE(" + q.Node.String() + ")"
}

func (a *Quote) Type() ObjectType {
	return QUOTE_OBJ
}

type Macro struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (m *Macro) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range m.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("macro")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ","))
	out.WriteString(")")
	out.WriteString("{")
	out.WriteString(m.Body.String())
	out.WriteString("}\n")

	return out.String()
}

func (m *Macro) Type() ObjectType {
	return FUNCTION_OBJ
}
