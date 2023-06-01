package ggl

import "reflect"

func convertToOriginalPointer(originalType reflect.Type, originalValue reflect.Value) reflect.Value {
	val := originalValue
	for i := 0; i < numOfPointer(originalType); i++ {
		tmp := reflect.New(val.Type())
		tmp.Elem().Set(val)
		val = tmp
	}
	return val
}

func numOfPointer(val reflect.Type) int {
	count := 0
	for {
		if val.Kind() != reflect.Ptr {
			return count
		}
		val = val.Elem()
		count += 1
	}
}

func cleanPtrValue(val reflect.Value) reflect.Value {
	for {
		if val.Kind() != reflect.Ptr {
			return val
		}
		val = reflect.Indirect(val)
	}
}

func cleanPtrType(val reflect.Type) reflect.Type {
	for {
		if val.Kind() != reflect.Ptr {
			return val
		}
		val = val.Elem()
	}
}

func checkIsContext(val reflect.Type) bool {
	return val.PkgPath() == "context" && val.Name() == "Context"
}
