package ggl

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

type manager struct {
	schema             graphql.Schema
	validator          validator
	baseScalarObject   map[string]graphql.Output
	customScalarObject map[string]graphql.Output
	graphKeyTag        string
	rootObjectKeyTag   string
}

type executor struct {
	schema          graphql.Schema
	requestString   string
	rootObject      map[string]interface{}
	variablesValues map[string]interface{}
}

func (loader *manager) RegisterSchema(resolver interface{}) error {
	schema, err := loader.graphSchema(resolver)
	if err != nil {
		return err
	}
	loader.schema = schema
	return nil
}

func (loader *manager) WriteMagidoc(folder string) error {
	// introspection query
	// queryGenerationFactories
	// - loop through baseScalarObject & customScalarObject and set default value for baseScalarObject
	return nil
}

func (loader *manager) RegisterScalar(i interface{}, o graphql.Output) {
	cleanType := cleanPtrType(reflect.TypeOf(i))
	loader.customScalarObject[scalarNameFromType(cleanType)] = o
}

func (loader *manager) RegisterValidator(validator validator) {
	loader.validator = validator
}

func New() *manager {
	loader := new(manager)
	loader.graphKeyTag = "gql"
	loader.rootObjectKeyTag = "root"
	loader.baseScalarObject = make(map[string]graphql.Output)
	loader.customScalarObject = make(map[string]graphql.Output)
	return loader
}
