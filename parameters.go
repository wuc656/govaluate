package govaluate

import (
	"errors"
)

/*
Parameters is a collection of named parameters that can be used by an EvaluableExpression to retrieve parameters
when an expression tries to use them.
*/
type Parameters interface {

	/*
		Get gets the parameter of the given name, or an error if the parameter is unavailable.
		Failure to find the given parameter should be indicated by returning an error.
	*/
	Get(name string) (any, error)
}

type MapParameters map[string]any

func (p MapParameters) Get(name string) (any, error) {

	value, found := p[name]

	if !found {
		errorMessage := "No parameter '" + name + "' found."
		return nil, errors.New(errorMessage)
	}

	return value, nil
}
