package object

import "fmt"

type Environment struct {
	vars   map[string]Object
	parent *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		vars:   map[string]Object{},
		parent: parent,
	}
}

func (e *Environment) Get(id string) (Object, error) {
	if e == nil {
		return nil, fmt.Errorf("undefined identifier: %s\n", id)
	}
	if obj, ok := e.vars[id]; ok {
		return obj, nil
	}
	return e.parent.Get(id)
}

func (e *Environment) Set(id string, obj Object) (Object, error) {
	// TODO: do we allow repeated definition?
	e.vars[id] = obj
	return obj, nil
}
