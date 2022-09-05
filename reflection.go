package ggl

import "reflect"

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
