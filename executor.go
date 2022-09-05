package ggl

import (
	"context"

	"github.com/graphql-go/graphql"
)

func (loader *manager) Do() executor {
	return executor{schema: loader.schema}
}

func (exe executor) Query(query string) executor {
	exe.requestString = query
	return exe
}

func (exe executor) Root(rootObject map[string]interface{}) executor {
	exe.rootObject = rootObject
	return exe
}

func (exe executor) Variables(variablesValues map[string]interface{}) executor {
	exe.variablesValues = variablesValues
	return exe
}

func (exe executor) Execute(ctx context.Context) *graphql.Result {
	return graphql.Do(graphql.Params{
		Context:        ctx,
		Schema:         exe.schema,
		RequestString:  exe.requestString,
		RootObject:     exe.rootObject,
		VariableValues: exe.variablesValues,
	})
}
