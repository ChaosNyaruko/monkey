package eval

import (
	"fmt"

	"github.com/ChaosNyaruko/monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Name: "len",
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("wrong number of arguments, expected %d, but got %d\n", 1, len(args))
			}
			a := args[0]
			s, ok := a.(*object.String)
			if !ok {
				return nil, fmt.Errorf("not supported on %v\n", a.Type())
			}
			return &object.Integer{
				Value: len(s.Value),
			}, nil
		},
	},
}
