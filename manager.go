package ggl

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

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

func (loader *manager) GraphKey(graphKey string) {
	loader.graphKeyTag = graphKey
}

func (loader *manager) RootObjectKey(rootObjectKey string) {
	loader.rootObjectKeyTag = rootObjectKey
}

func (loader *manager) GetSchema() graphql.Schema {
	return loader.schema
}

func (loader *manager) RegisterSchema(resolver interface{}) error {
	schema, err := loader.graphSchema(resolver)
	if err != nil {
		return err
	}
	loader.schema = schema
	return nil
}

func (loader *manager) WriteSchema(file string) error {
	result := graphql.Do(graphql.Params{
		Schema:        loader.schema,
		RequestString: introspectionQuery,
	})

	if result.HasErrors() {
		return result.Errors[0]
	}

	bytes, err := json.Marshal(result.Data)
	if err != nil {
		return err
	}

	schemaFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer schemaFile.Close()

	_, err = schemaFile.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

func (loader *manager) WriteMagidoc(file string, schemaFile string) error {
	magidocFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer magidocFile.Close()

	options := make([]string, 0)
	for name := range loader.baseScalarObject {
		if strings.HasPrefix(name, "goarray_") || strings.HasPrefix(name, "goslice_") {
			options = append(options, fmt.Sprintf("'%v': '[]'", name))
		} else if strings.HasPrefix(name, "gomap_") {
			options = append(options, fmt.Sprintf("'%v': '{}'", name))
		}
	}

	configuration := fmt.Sprintf(`
		export default {
			introspection: {
				type: 'file',
				location: '%s',
			},
			website: {
				template: 'carbon-multi-page',
				options: {
					queryGenerationFactories: {
						'RawString': '',
						'GoStringer': '',
						%v
					}
				}
			},
		}
	`, schemaFile, strings.Join(options, ",\n"))

	_, err = magidocFile.Write([]byte(configuration))
	if err != nil {
		return err
	}
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
