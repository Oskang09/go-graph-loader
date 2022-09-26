package ggl

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/iancoleman/strcase"
)

var (
	rawScalarObjectFunc = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "RawString",
		Description: "The `RawString` scalar type represents any type.",
		Serialize: func(value interface{}) interface{} {
			return fmt.Sprintf("%v", value)
		},
	})

	goStringerScalarObjectFunc = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "GoStringer",
		Description: "The `GoStringer` scalar type represents Stringer type.",
		Serialize: func(value interface{}) interface{} {
			stringer := value.(fmt.GoStringer)
			return stringer.GoString()
		},
	})
)

func scalarNameFromType(t reflect.Type) string {
	scalarName := t.PkgPath() + "." + t.Name()
	if t.Name() == "" {
		scalarName = t.PkgPath() + ".interface"
	}
	scalarName = strings.ReplaceAll(scalarName, ".", "_")
	scalarName = strings.ReplaceAll(scalarName, "/", "_")
	scalarName = strings.ReplaceAll(scalarName, "-", "_")
	return scalarName
}

func graphNameFromTag(tag reflect.StructTag, key string) string {
	return strings.Split(tag.Get(key), ",")[0]
}

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
		requestType := cleanPtrType(methodType.In(2))
		for i := 0; i < requestType.NumField(); i++ {
			field := requestType.Field(i)

			root := graphNameFromTag(field.Tag, loader.rootObjectKeyTag)
			if root != "" {
				rootLoaderArgs[root] = field.Name
			}

			gql := graphNameFromTag(field.Tag, loader.graphKeyTag)
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
			Name:   cleanPtrType(responseType).Name(),
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
			request := reflect.New(cleanPtrType(methodType.In(2)))
			for graphKey, goKey := range graphLoaderArgs {
				if val, ok := p.Args[graphKey]; ok {
					field := cleanPtrValue(request).FieldByName(goKey)
					field.Set(reflect.ValueOf(val))
				}
			}

			if len(rootLoaderArgs) > 0 {
				rootObject := p.Info.RootValue.(map[string]interface{})
				for graphKey, goKey := range rootLoaderArgs {
					if val, ok := rootObject[graphKey]; ok {
						field := cleanPtrValue(request).FieldByName(goKey)
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
	graphKey := graphNameFromTag(field.Tag, loader.graphKeyTag)
	if graphKey == "" {
		return "", nil
	}
	return graphKey, &graphql.Field{Name: field.Name, Type: scalarType}
}

func (loader *manager) graphArgumentConfigByStructField(field reflect.StructField) (string, *graphql.ArgumentConfig) {
	scalarType := loader.graphByTypes(field.Type)
	graphKey := graphNameFromTag(field.Tag, loader.graphKeyTag)
	if graphKey == "" {
		return "", nil
	}
	return graphKey, &graphql.ArgumentConfig{Type: scalarType}
}

func (loader *manager) graphFieldsByType(ptrType reflect.Type) graphql.Fields {
	outputType := cleanPtrType(ptrType)
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
	cleanField := cleanPtrType(field)
	kind := cleanField.Kind()

	if scalar, ok := loader.customScalarObject[scalarNameFromType(cleanField)]; ok {
		return scalar
	}

	switch kind {

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
		safeField := cleanPtrType(field)
		scalarName := scalarNameFromType(safeField)
		if _, ok := loader.baseScalarObject[scalarName]; !ok {
			loader.baseScalarObject[scalarName] = graphql.NewObject(graphql.ObjectConfig{
				Name:   scalarName,
				Fields: loader.graphFieldsByType(field),
			})
		}
		return loader.baseScalarObject[scalarName]

	case reflect.Slice:
		parentType := cleanPtrType(field)
		childType := parentType.Elem()
		if childType.Kind() == reflect.Ptr {
			childType = childType.Elem()
		}

		if childType.Kind() == reflect.Struct {
			return graphql.NewList(loader.graphByTypes(childType))
		}
		return loader.sliceScalarObject(field, childType)

	case reflect.Array:
		parentType := cleanPtrType(field)
		childType := parentType.Elem()
		if childType.Kind() == reflect.Ptr {
			childType = childType.Elem()
		}

		if childType.Kind() == reflect.Struct {
			return graphql.NewList(loader.graphByTypes(childType))
		}
		return loader.arrayScalarObject(field, childType)

	case reflect.Map:
		return loader.mapScalarObject(field)
	}

	if field.Implements(reflect.TypeOf(new(fmt.GoStringer)).Elem()) {
		return goStringerScalarObjectFunc
	}

	return rawScalarObjectFunc
}

func (loader *manager) mapScalarObject(t reflect.Type) graphql.Output {
	keyType := cleanPtrType(t).Key()
	valueType := cleanPtrType(t).Elem()
	valueName := valueType.Name()
	keyName := keyType.Name()
	if keyName == "" {
		keyName = "interface"
	}
	if valueName == "" {
		valueName = "interface"
	}

	scalarName := "gomap_" + keyName + "_" + valueName
	if _, ok := loader.baseScalarObject[scalarName]; !ok {
		loader.baseScalarObject[scalarName] = graphql.NewScalar(graphql.ScalarConfig{
			Name:        scalarName,
			Description: "The `gomap_" + keyName + "_" + valueName + "` scalar type represents map[" + keyName + "]" + valueName + " data.",
			ParseValue: func(value interface{}) interface{} {
				mapValue := reflect.MakeMap(cleanPtrType(t))
				if value == nil {
					return convertToOriginalPointer(t, mapValue).Interface()
				}

				jsonString := fmt.Sprintf("%v", value)
				jsonMap := make(map[string]interface{})
				if err := json.Unmarshal([]byte(jsonString), &jsonMap); err != nil {
					return convertToOriginalPointer(t, mapValue).Interface()
				}

				for key, value := range jsonMap {
					val := reflect.ValueOf(value)
					if val.CanConvert(valueType) {
						mapValue.SetMapIndex(reflect.ValueOf(key), val.Convert(valueType))
					}
				}
				return convertToOriginalPointer(t, mapValue).Interface()
			},
			ParseLiteral: func(valueAST ast.Value) interface{} {
				mapValue := reflect.MakeMap(cleanPtrType(t))
				jsonString := fmt.Sprintf("%v", valueAST.GetValue())
				jsonMap := make(map[string]interface{})
				if err := json.Unmarshal([]byte(jsonString), &jsonMap); err != nil {
					return convertToOriginalPointer(t, mapValue).Interface()
				}

				for key, value := range jsonMap {
					val := reflect.ValueOf(value)
					if val.CanConvert(valueType) {
						mapValue.SetMapIndex(reflect.ValueOf(key), val.Convert(valueType))
					}
				}
				return convertToOriginalPointer(t, mapValue).Interface()
			},
			Serialize: func(value interface{}) interface{} {
				bytes, err := json.Marshal(value)
				if err != nil {
					return "{}"
				}
				return json.RawMessage(bytes)
			},
		})
	}
	return loader.baseScalarObject[scalarName]
}

func (loader *manager) arrayScalarObject(t reflect.Type, childType reflect.Type) graphql.Output {
	childName := childType.Name()
	if childName == "" {
		childName = "interface"
	}
	scalarName := "goarray_" + childName
	if _, ok := loader.baseScalarObject[scalarName]; !ok {
		loader.baseScalarObject[scalarName] = graphql.NewScalar(graphql.ScalarConfig{
			Name:        scalarName,
			Description: "The `goarray_" + childName + "` scalar type represents [n]" + childName + " data.",
			ParseValue: func(value interface{}) interface{} {
				arrayValue := reflect.New(reflect.ArrayOf(cleanPtrType(t).Len(), cleanPtrType(childType))).Elem()
				jsonString := fmt.Sprintf("%v", value)
				jsonArray := make([]interface{}, 0)
				if err := json.Unmarshal([]byte(jsonString), &jsonArray); err != nil {
					return convertToOriginalPointer(t, arrayValue).Interface()
				}

				for index, value := range jsonArray {
					val := reflect.ValueOf(value)
					if val.CanConvert(childType) {
						arrayValue.Index(index).Set(val.Convert(childType))
					}
				}
				return convertToOriginalPointer(t, arrayValue).Interface()
			},
			ParseLiteral: func(valueAST ast.Value) interface{} {
				arrayValue := reflect.New(reflect.ArrayOf(cleanPtrType(t).Len(), cleanPtrType(childType))).Elem()
				jsonString := fmt.Sprintf("%v", valueAST.GetValue())
				jsonArray := make([]interface{}, 0)
				if err := json.Unmarshal([]byte(jsonString), &jsonArray); err != nil {
					return convertToOriginalPointer(t, arrayValue).Interface()
				}

				for index, value := range jsonArray {
					val := reflect.ValueOf(value)
					if val.CanConvert(childType) {
						arrayValue.Index(index).Set(val.Convert(childType))
					}
				}
				return convertToOriginalPointer(t, arrayValue).Interface()
			},
			Serialize: func(value interface{}) interface{} {
				bytes, err := json.Marshal(value)
				if err != nil {
					return "[]"
				}
				return json.RawMessage(bytes)
			},
		})
	}
	return loader.baseScalarObject[scalarName]
}

func (loader *manager) sliceScalarObject(t reflect.Type, childType reflect.Type) graphql.Output {
	childName := childType.Name()
	if childName == "" {
		childName = "interface"
	}
	scalarName := "goslice_" + childName
	if _, ok := loader.baseScalarObject[scalarName]; !ok {
		loader.baseScalarObject[scalarName] = graphql.NewScalar(graphql.ScalarConfig{
			Name:        scalarName,
			Description: "The `goslice_" + childName + "` scalar type represents []" + childName + " data.",
			ParseValue: func(value interface{}) interface{} {
				sliceValue := reflect.New(reflect.SliceOf(childType)).Elem()
				jsonString := fmt.Sprintf("%v", value)
				jsonArray := make([]interface{}, 0)
				if err := json.Unmarshal([]byte(jsonString), &jsonArray); err != nil {
					return convertToOriginalPointer(t, sliceValue).Interface()
				}

				for _, value := range jsonArray {
					val := reflect.ValueOf(value)
					if val.CanConvert(childType) {
						sliceValue.Set(reflect.Append(sliceValue, val.Convert(childType)))
					}
				}
				return convertToOriginalPointer(t, sliceValue).Interface()
			},
			ParseLiteral: func(valueAST ast.Value) interface{} {
				sliceValue := reflect.New(reflect.SliceOf(childType)).Elem()
				jsonString := fmt.Sprintf("%v", valueAST.GetValue())
				jsonArray := make([]interface{}, 0)
				if err := json.Unmarshal([]byte(jsonString), &jsonArray); err != nil {
					return convertToOriginalPointer(t, sliceValue).Interface()
				}

				for _, value := range jsonArray {
					val := reflect.ValueOf(value)
					if val.CanConvert(childType) {
						sliceValue.Set(reflect.Append(sliceValue, val.Convert(childType)))
					}
				}
				return convertToOriginalPointer(t, sliceValue).Interface()
			},
			Serialize: func(value interface{}) interface{} {
				bytes, err := json.Marshal(value)
				if err != nil {
					return "[]"
				}
				return json.RawMessage(bytes)
			},
		})
	}
	return loader.baseScalarObject[scalarName]
}
