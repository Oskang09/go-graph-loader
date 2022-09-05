package ggl

import (
	"encoding/json"
	"fmt"

	"github.com/graphql-go/graphql"
)

type manager struct {
	schema       graphql.Schema
	validator    validator
	scalarObject map[string]graphql.Output
}

type executor struct {
	schema          graphql.Schema
	requestString   string
	rootObject      map[string]interface{}
	variablesValues map[string]interface{}
}

func (loader *manager) LoadSchema(resolver interface{}) error {
	schema, err := loader.graphSchema(resolver)
	if err != nil {
		return err
	}
	loader.schema = schema
	return nil
}

func New() *manager {
	loader := new(manager)
	loader.scalarObject = make(map[string]graphql.Output)
	loader.scalarObject["_raw"] = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "RawString",
		Description: "The `RawString` scalar type represents any type.",
		Serialize: func(value interface{}) interface{} {
			return fmt.Sprintf("%v", value)
		},
	})
	loader.scalarObject["_string"] = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "GoStringer",
		Description: "The `GoStringer` scalar type represents Stringer type.",
		Serialize: func(value interface{}) interface{} {
			stringer := value.(fmt.GoStringer)
			return stringer.GoString()
		},
	})
	loader.scalarObject["_array"] = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "GoArray",
		Description: "The `GoArray` scalar type represents []type data.",
		Serialize: func(value interface{}) interface{} {
			bytes, err := json.Marshal(value)
			if err != nil {
				return "[]"
			}
			return json.RawMessage(bytes)
		},
	})
	loader.scalarObject["_map"] = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "GoMap",
		Description: "The `GoMap` scalar type represents map[type]type data.",
		Serialize: func(value interface{}) interface{} {
			bytes, err := json.Marshal(value)
			if err != nil {
				return "{}"
			}
			return json.RawMessage(bytes)
		},
	})

	return loader
}
