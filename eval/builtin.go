package eval

import (
	"fmt"

	"github.com/ChaosNyaruko/monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len":   {Name: "len", Fn: Len},
	"first": {Name: "first", Fn: First},
	"last":  {Name: "last", Fn: Last},
	"rest":  {Name: "rest", Fn: Rest},
	"push":  {Name: "push", Fn: Push},
}

func Len(args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments, expected %d, but got %d\n", 1, len(args))
	}
	a := args[0]
	switch s := a.(type) {
	case *object.String:
		return &object.Integer{
			Value: len(s.Value),
		}, nil
	case *object.Array:
		return &object.Integer{
			Value: len(s.Elements),
		}, nil
	default:
		return nil, fmt.Errorf("not supported on %v\n", a.Type())
	}
}

func First(args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments, expected %d, but got %d\n", 1, len(args))
	}
	a := args[0]
	switch s := a.(type) {
	case *object.Array:
		if len(s.Elements) == 0 {
			return NULL, nil
		}
		return s.Elements[0], nil
	default:
		return nil, fmt.Errorf("not supported on %v\n", a.Type())
	}
}

func Last(args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments, expected %d, but got %d\n", 1, len(args))
	}
	a := args[0]
	switch s := a.(type) {
	case *object.Array:
		if len(s.Elements) == 0 {
			return NULL, nil
		}
		return s.Elements[len(s.Elements)-1], nil
	default:
		return nil, fmt.Errorf("not supported on %v\n", a.Type())
	}
}

func Rest(args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments, expected %d, but got %d\n", 1, len(args))
	}
	a := args[0]
	switch s := a.(type) {
	case *object.Array:
		if len(s.Elements) == 0 {
			return NULL, nil
		}
		return &object.Array{
			Elements: s.Elements[1:],
		}, nil
	default:
		return nil, fmt.Errorf("not supported on %v\n", a.Type())
	}
}

func Push(args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments, expected %d, but got %d\n", 2, len(args))
	}
	a := args[0]
	o := args[1]
	switch s := a.(type) {
	case *object.Array:
		newArray := &object.Array{}
		l := len(s.Elements)
		newArray.Elements = make([]object.Object, l+1)
		copy(newArray.Elements, s.Elements)
		newArray.Elements[l] = o
		return newArray, nil
	default:
		return nil, fmt.Errorf("not supported on %v\n", a.Type())
	}
}
