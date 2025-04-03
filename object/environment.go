package object

import "fmt"

type Environment struct {
	vars map[string]Object
}

func NewEnvironment() *Environment {
	return &Environment{
		vars: map[string]Object{},
	}
}

func (e *Environment) Get(id string) (Object, error) {
	obj, ok := e.vars[id]
	if !ok {
		return nil, fmt.Errorf("undefined identifier: %s", id)
	}
	return obj, nil
}

func (e *Environment) Set(id string, obj Object) (Object, error) {
	// TODO: do we allow repeated definition?
	e.vars[id] = obj
	return obj, nil
}
