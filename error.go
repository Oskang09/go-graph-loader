package ggl

import (
	"log"
	"reflect"
)

const (
	// definitions
	definitionTypePreResolver    = "PRE_RESOLVER"
	definitionTypeResolverMethod = "RESOLVER_METHOD"

	// errors
	errInvalidMethodSignatureForPreResolverFunction   = "invalid method signature is using for pre resolver function"
	errInvalidMethodSignatureForFieldResolverFunction = "invalid method signature is using for field resolver function"
)

func panicWithFootprint(
	definitionType string,
	object reflect.Type,
	signatureType reflect.Type,
	debug string,
) {
	cleanObject := cleanPtrType(object)
	log.Println("————————————— Go Graph Loader —————————————")
	log.Println("| Package   | " + cleanObject.PkgPath())
	log.Println("| Struct    | " + cleanObject.Name())
	log.Println("| Type      | " + definitionType)
	if signatureType != nil {
		log.Println("| Signature | " + signatureType.String())
	}
	log.Println("———————————————————————————————————————————")
	panic("go-graph-loader: " + debug)
}
