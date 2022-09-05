package main

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
)

func main() {
	resolver := New()

	schema, err := graphSchema(resolver)
	if err != nil {
		panic(err)
	}

	result := graphql.Do(graphql.Params{
		Schema: schema,
		RequestString: `{ 
			products(cursor: "some nested cursor") {
				list
				cursor
			}
			product(id: 1) {
				info, name,
				price(multiply: 100), 
				priceInteger,
				other { 
					p5, 
					p6 {
						p7
					},
					p8
				}, 
				json2
			} 
		}`,
		Context: context.Background(),
	})

	log.Println(result.Errors)
	bytes, _ := json.Marshal(result.Data)
	log.Println(string(bytes))
}

func graphSchema(resolver interface{}) (graphql.Schema, error) {
	rootQuery := graphql.Fields{}
	val := reflect.ValueOf(resolver)
	valType := reflect.TypeOf(resolver)

	for i := 0; i < val.NumMethod(); i++ {
		method := val.Method(i)
		methodType := method.Type()
		methodDefinition := valType.Method(i)
		methodDefinitionName := strcase.ToLowerCamel(valType.Method(i).Name)
		ptrReturnType := methodType.Out(0)
		returnType := tryElemType(ptrReturnType)

		graphObject := graphql.NewObject(graphql.ObjectConfig{
			Name:   returnType.Name(),
			Fields: graphFieldsByType(ptrReturnType),
		})

		graphField := new(graphql.Field)
		graphField.Name = methodDefinitionName
		graphField.Type = graphObject
		graphField.Args, graphField.Resolve = graphRootResolverByMethod(val, methodDefinition)
		rootQuery[methodDefinitionName] = graphField
	}

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Query",
			Fields: rootQuery,
		}),
	})
}

func tryElem(val reflect.Value) reflect.Value {
	if val.Kind() == reflect.Ptr {
		return val.Elem()
	}
	return val
}

func tryElemType(val reflect.Type) reflect.Type {
	if val.Kind() == reflect.Ptr {
		return val.Elem()
	}
	return val
}

func graphRootResolverByMethod(self reflect.Value, method reflect.Method) (graphql.FieldConfigArgument, func(graphql.ResolveParams) (interface{}, error)) {
	methodType := method.Type
	graphArgs := graphql.FieldConfigArgument{}
	graphLoaderArgs := make(map[string]string)
	if method.Type.NumIn() == 3 {
		requestType := tryElemType(methodType.In(2))
		for i := 0; i < requestType.NumField(); i++ {
			field := requestType.Field(i)
			tag := field.Tag.Get("ggl")
			fn, ac := graphArgumentConfigByStructField(field)
			if ac == nil {
				continue
			}
			graphArgs[fn] = ac
			graphLoaderArgs[tag] = field.Name
		}
	}

	return graphArgs, func(p graphql.ResolveParams) (interface{}, error) {
		rValues := make([]reflect.Value, 0)

		rValues = append(rValues, self)
		rValues = append(rValues, reflect.ValueOf(p.Context))

		if method.Type.NumIn() == 3 {
			requestType := tryElemType(methodType.In(2))
			request := reflect.New(requestType)
			for graphKey, goKey := range graphLoaderArgs {
				if val, ok := p.Args[graphKey]; ok {
					field := tryElem(request).FieldByName(goKey)
					field.Set(reflect.ValueOf(val))
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

func graphResolverByMethod(method reflect.Method) (graphql.FieldConfigArgument, graphql.Output, func(graphql.ResolveParams) (interface{}, error)) {
	methodType := method.Type
	graphArgs := graphql.FieldConfigArgument{}
	graphLoaderArgs := make(map[string]string)
	if method.Type.NumIn() == 3 {
		requestType := tryElemType(methodType.In(2))
		for i := 0; i < requestType.NumField(); i++ {
			field := requestType.Field(i)
			tag := field.Tag.Get("ggl")
			fn, ac := graphArgumentConfigByStructField(field)
			if ac == nil {
				continue
			}
			graphArgs[fn] = ac
			graphLoaderArgs[tag] = field.Name
		}
	}

	responseType := methodType.Out(0)
	graphOutput := graphByTypes(responseType)
	return graphArgs, graphOutput, func(p graphql.ResolveParams) (interface{}, error) {
		rValues := make([]reflect.Value, 0)

		rValues = append(rValues, reflect.ValueOf(p.Source))
		rValues = append(rValues, reflect.ValueOf(p.Context))

		if method.Type.NumIn() == 3 {
			request := reflect.New(tryElemType(methodType.In(2)))
			for graphKey, goKey := range graphLoaderArgs {
				if val, ok := p.Args[graphKey]; ok {
					field := tryElem(request).FieldByName(goKey)
					field.Set(reflect.ValueOf(val))
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

func graphFieldByStructField(field reflect.StructField) (string, *graphql.Field) {
	scalarType := graphByTypes(field.Type)
	graphKey := field.Tag.Get("ggl")
	if graphKey == "" {
		graphKey = strcase.ToLowerCamel(field.Name)
	}
	return graphKey, &graphql.Field{Name: field.Name, Type: scalarType}
}

func graphArgumentConfigByStructField(field reflect.StructField) (string, *graphql.ArgumentConfig) {
	scalarType := graphByTypes(field.Type)
	graphKey := field.Tag.Get("ggl")
	if graphKey == "" {
		graphKey = strcase.ToLowerCamel(field.Name)
	}
	return graphKey, &graphql.ArgumentConfig{Type: scalarType}
}

func graphFieldsByType(ptrType reflect.Type) graphql.Fields {
	outputType := tryElemType(ptrType)
	graphFields := graphql.Fields{}

	reservedFields := make(map[string]*struct{})
	for j := 0; j < outputType.NumField(); j++ {
		field := outputType.Field(j)
		fn, gf := graphFieldByStructField(field)
		methodResolver, hasMethod := ptrType.MethodByName("GGL_" + field.Name)
		if hasMethod {
			reservedFields["GGL_"+field.Name] = nil
			gf.Args, gf.Type, gf.Resolve = graphResolverByMethod(methodResolver)
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
					gf.Args, gf.Type, gf.Resolve = graphResolverByMethod(methodResolver)
				}
				graphFields[strcase.ToLowerCamel(graphName)] = gf
			}
		}
	}

	return graphFields
}

var (
	baseArrayScalar = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "JList",
		Description: "The `JList` scalar type represents []type data.",
		Serialize: func(value interface{}) interface{} {
			bytes, err := json.Marshal(value)
			if err != nil {
				return "[]"
			}
			return json.RawMessage(bytes)
		},
	})
	baseMapScalar = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "JMap",
		Description: "The `JMap` scalar type represents map[type]type data.",
		Serialize: func(value interface{}) interface{} {
			bytes, err := json.Marshal(value)
			if err != nil {
				return "{}"
			}
			return json.RawMessage(bytes)
		},
	})
)

func graphByTypes(field reflect.Type) graphql.Output {
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
		return graphql.NewObject(graphql.ObjectConfig{
			Name:   strings.ReplaceAll(safeField.PkgPath()+"."+safeField.Name(), ".", "_"),
			Fields: graphFieldsByType(field),
		})

	case reflect.Array, reflect.Slice:
		safeField := tryElemType(field).Elem()
		if safeField.Kind() == reflect.Struct {
			return graphql.NewList(graphByTypes(safeField))
		}
		return baseArrayScalar

	case reflect.Map:
		return baseMapScalar
	}

	return graphql.String
}
