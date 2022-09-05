package ggl

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
)

func (loader *manager) graphSchema(resolver interface{}) (graphql.Schema, error) {
	rootQuery := graphql.Fields{}
	val := reflect.ValueOf(resolver)
	valType := reflect.TypeOf(resolver)

	for i := 0; i < val.NumMethod(); i++ {
		methodDefinition := valType.Method(i)
		methodDefinitionName := strcase.ToLowerCamel(valType.Method(i).Name)

		graphField := new(graphql.Field)
		graphField.Name = methodDefinitionName
		graphField.Args, graphField.Type, graphField.Resolve = loader.graphResolverByMethod(&val, methodDefinition)
		rootQuery[methodDefinitionName] = graphField
	}

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Query",
			Fields: rootQuery,
		}),
	})
}

func (loader *manager) graphResolverByMethod(root *reflect.Value, method reflect.Method) (graphql.FieldConfigArgument, graphql.Output, func(graphql.ResolveParams) (interface{}, error)) {
	methodType := method.Type
	graphArgs := graphql.FieldConfigArgument{}
	graphLoaderArgs := make(map[string]string)
	rootLoaderArgs := make(map[string]string)
	if method.Type.NumIn() == 3 {
		requestType := tryElemType(methodType.In(2))
		for i := 0; i < requestType.NumField(); i++ {
			field := requestType.Field(i)

			root := field.Tag.Get("root")
			if root != "" {
				rootLoaderArgs[root] = field.Name
			}

			gql := field.Tag.Get("gql")
			if gql != "" {
				fn, ac := loader.graphArgumentConfigByStructField(field)
				if ac == nil {
					continue
				}
				graphArgs[fn] = ac
				graphLoaderArgs[gql] = field.Name
			}
		}
	}

	responseType := methodType.Out(0)
	var graphOutput graphql.Output
	if root != nil {
		graphOutput = graphql.NewObject(graphql.ObjectConfig{
			Name:   tryElemType(responseType).Name(),
			Fields: loader.graphFieldsByType(responseType),
		})
	} else {
		graphOutput = loader.graphByTypes(responseType)
	}

	return graphArgs, graphOutput, func(p graphql.ResolveParams) (interface{}, error) {
		rValues := make([]reflect.Value, 0)

		if root != nil {
			rValues = append(rValues, *root)
		} else {
			rValues = append(rValues, reflect.ValueOf(p.Source))
		}
		rValues = append(rValues, reflect.ValueOf(p.Context))

		if method.Type.NumIn() == 3 {
			request := reflect.New(tryElemType(methodType.In(2)))
			for graphKey, goKey := range graphLoaderArgs {
				if val, ok := p.Args[graphKey]; ok {
					field := tryElem(request).FieldByName(goKey)
					field.Set(reflect.ValueOf(val))
				}
			}

			if len(rootLoaderArgs) > 0 {
				rootObject := p.Info.RootValue.(map[string]interface{})
				for graphKey, goKey := range rootLoaderArgs {
					if val, ok := rootObject[graphKey]; ok {
						field := tryElem(request).FieldByName(goKey)
						field.Set(reflect.ValueOf(val))
					}
				}
			}

			if loader.validator != nil {
				err := loader.validator.Validate(request.Interface())
				if err != nil {
					return nil, err
				}
			}

			rValues = append(rValues, request)
		}

		rsp := method.Func.Call(rValues)
		if rsp[1].Interface() != nil {
			return rsp[0].Interface(), rsp[1].Interface().(error)
		}
		return rsp[0].Interface(), nil
	}
}

func (loader *manager) graphFieldByStructField(field reflect.StructField) (string, *graphql.Field) {
	scalarType := loader.graphByTypes(field.Type)
	graphKey := field.Tag.Get("gql")
	if graphKey == "" {
		return "", nil
	}
	return graphKey, &graphql.Field{Name: field.Name, Type: scalarType}
}

func (loader *manager) graphArgumentConfigByStructField(field reflect.StructField) (string, *graphql.ArgumentConfig) {
	scalarType := loader.graphByTypes(field.Type)
	graphKey := field.Tag.Get("gql")
	if graphKey == "" {
		return "", nil
	}
	return graphKey, &graphql.ArgumentConfig{Type: scalarType}
}

func (loader *manager) graphFieldsByType(ptrType reflect.Type) graphql.Fields {
	outputType := tryElemType(ptrType)
	graphFields := graphql.Fields{}

	reservedFields := make(map[string]*struct{})
	for j := 0; j < outputType.NumField(); j++ {
		field := outputType.Field(j)
		fn, gf := loader.graphFieldByStructField(field)
		if gf == nil {
			continue
		}

		methodResolver, hasMethod := ptrType.MethodByName("GGL_" + field.Name)
		if hasMethod {
			reservedFields["GGL_"+field.Name] = nil
			gf.Args, gf.Type, gf.Resolve = loader.graphResolverByMethod(nil, methodResolver)
		}
		graphFields[fn] = gf
	}

	for j := 0; j < ptrType.NumMethod(); j++ {
		field := ptrType.Method(j)
		if strings.HasPrefix(field.Name, "GGL_") {
			if _, ok := reservedFields[field.Name]; !ok {
				graphName := strcase.ToLowerCamel(strings.TrimPrefix(field.Name, "GGL_"))
				gf := new(graphql.Field)
				gf.Name = field.Name
				methodResolver, hasMethod := ptrType.MethodByName(field.Name)
				if hasMethod {
					gf.Args, gf.Type, gf.Resolve = loader.graphResolverByMethod(nil, methodResolver)
				}
				graphFields[strcase.ToLowerCamel(graphName)] = gf
			}
		}
	}

	return graphFields
}

func (loader *manager) graphByTypes(field reflect.Type) graphql.Output {
	kind := field.Kind()
	if kind == reflect.Ptr {
		kind = field.Elem().Kind()
	}

	switch field.Kind() {
	case reflect.Bool:
		return graphql.Boolean

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return graphql.Int

	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return graphql.Float

	case reflect.String:
		return graphql.String

	case reflect.Struct:
		safeField := tryElemType(field)

		scalarName := safeField.PkgPath() + "." + safeField.Name()
		scalarName = strings.ReplaceAll(scalarName, ".", "_")
		scalarName = strings.ReplaceAll(scalarName, "/", "_")
		scalarName = strings.ReplaceAll(scalarName, "-", "_")

		if _, ok := loader.scalarObject[scalarName]; !ok {
			loader.scalarObject[scalarName] = graphql.NewObject(graphql.ObjectConfig{
				Name:   scalarName,
				Fields: loader.graphFieldsByType(field),
			})
		}
		return loader.scalarObject[scalarName]

	case reflect.Array, reflect.Slice:
		safeField := tryElemType(field).Elem()
		if safeField.Kind() == reflect.Ptr {
			safeField = safeField.Elem()
		}

		if safeField.Kind() == reflect.Struct {
			return graphql.NewList(loader.graphByTypes(safeField))
		}
		return loader.scalarObject["_array"]

	case reflect.Map:
		return loader.scalarObject["_map"]
	}

	if field.Implements(reflect.TypeOf(new(fmt.GoStringer)).Elem()) {
		return loader.scalarObject["_string"]
	}
	return loader.scalarObject["_raw"]
}
