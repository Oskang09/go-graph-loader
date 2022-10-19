# Go Graph Loader ( GGL )

go graph loader is a plugin for load the grahql by resolver and process scalar type with go struct definition instead of schema typing, and come with some simple extension like validator.

# Tag & Method Siganture

In this plugin we're using `gql` as graphql key loader and `root` as root object arguments loader. And about method signature we're following as per below, the `responseType` will be ur model definition so in order to let us to generate graphql response schema as well.

```go
type Resolver struct {

}

func (*Resolver) Product(context.Context) (responseType, error) {

}

type ProductRequest struct {
    Merchant string `root:"merchant"`
    ID string `gql:"id"`
}

func (*Resolver) Product(context.Context, *ProductRequest) (responseType, error) {

}
```

# Model Definition

For model definition by default we're not exposing all the fields only the fields with `gql` tagged will be exposed. Other than that we did support for field resolver or custom resolver which mean we can add extra function on model.


# Supported Primitive Types

```
1. bool
2. int
3. float64
4. string
```

## Model Field Resolver

As per model field resolver, we can overriding the original field resolver which just exposing the value, with this we can customize based on the source of value. For method signature as per [Tag & Method Signature](#tag--method-siganture) mentioned it can be only `context` value or with custom request arguments/

```go
type Product struct {
	ID   int64  `gql:"id"`
	Name string `gql:"name"`
        Price float64 `gql:"price"`
}

type ProductNameArgs struct {
    Extra string `gql:"name"`
}

func (product *Product) GGL_Name(ctx context.Context, args *ProductNameArgs) (string, error) {
	return product.Name + "-" + args.Extra, nil
}

func (product *Product) GGL_Price(ctx context.Context) (int64, error) {
	return int64(product.Price * 100), nil
}
```

```json
{
    product {
        name(extra: "extraname")
        price
    }
}
```

# Custom Field Resolver

For custom field resolver, we're not overriding the original field resolver but we create new resolver for itself with source of value. All the function will be [camelCase](https://en.wikipedia.org/wiki/Camel_case) when define in graphql query.


```go
type Product struct {
	ID   int64  `gql:"id"`
	Name string `gql:"name"`
        Price float64 `gql:"price"`
}

func (product *Product) GGL_NameWithPrice(ctx context.Context) (string, error) {
	return fmt.Sprintf("%v=%v", product.Name, product.Price), nil
}
```

```json
{
    product {
        name,
        price,
        nameWithPrice
    }
}
```

# Code & Execution

```go
package main

import (
	"context"
        "json"
        "log"

	ggl "github.com/Oskang09/go-graph-loader"
)

func main() {
    resolver := // your resolver struct which contains all the root functions

    manager := ggl.New()
    manager.RegisterSchema(resolver)

    // define your custom validator
    // so with this you can validate your incoming 
    // parameters with your own validator
    manager.RegisterValidator(nil)

    // magidoc template generator
    manager.WriteSchema("schema.json")
    manager.WriteMagidoc("magidoc.mjs", "schema.json")

    result := manager.Do().
        Query("{ product { name } }"). // set your query string
        Root(map[string]interface{}{"value":"some root value"}). // (optional) set your root object
        Execute(context.Background()) // your current context, it can be useful for tracking & tracing purpose

    if result.HasErrors() {
        log.Println(result.Errors)
    }
	bytes, _ := json.Marshal(result.Data)
	log.Println(string(bytes))
}
```

# Error & Debugging

We did provide the error footprint while having definition error or schema error, so would help you a lot when debugging the issues.


## PreResolver Siganture Error

```
2022/10/20 00:21:41 ————————————— Go Graph Loader —————————————
2022/10/20 00:21:41 | Package   | main
2022/10/20 00:21:41 | Struct    | Product
2022/10/20 00:21:41 | Type      | PRE_RESOLVER
2022/10/20 00:21:41 | Signature | func(*main.Product) context.Context
2022/10/20 00:21:41 ———————————————————————————————————————————
panic: go-graph-loader: invalid method signature is using for pre resolver function
```

## Resolver Function Signature Error

```
2022/10/20 00:18:03 ————————————— Go Graph Loader —————————————
2022/10/20 00:18:03 | Package   | main
2022/10/20 00:18:03 | Struct    | Product
2022/10/20 00:18:03 | Type      | RESOLVER_METHOD
2022/10/20 00:18:03 | Signature | func(*main.Product, *main.ProductNameArgs) (string, error)
2022/10/20 00:18:03 ———————————————————————————————————————————
panic: go-graph-loader: invalid method signature is using for field resolver function
```

# Documentation Tools

For documentating we will suggest go with [magidoc](https://magidoc.js.org/introduction/welcome) since they will build documentation based on your server's introspection query result. But if you using this plugins you will need to specifiy some custom scalar type which we using to process some array, struct and anoymous types. You can generate magidoc using this cli `magidoc generate`, after you have start the server with `WriteSchema` and `WriteMagidoc` function.

