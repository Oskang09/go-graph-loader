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
    manager.Schema(resolver)

    // define your custom validator
    // so with this you can validate your incoming 
    // parameters with your own validator
    manager.Validator(nil)

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

# Documentation Tools

For documentating we will suggest go with [magidoc](https://magidoc.js.org/introduction/welcome) since they will build documentation based on your server's introspection query result. But if you using this plugins you will need to specifiy some custom scalar type which we using to process some array, struct and anoymous types, you can found it at [magicdoc.mjs](magidoc.mjs). You can generate using this cli `magidoc generate -f schema/magidoc.mjs`

```js
export default {
    introspection: {
        type: 'file',
        location: 'schema/schema.json',
    },
    website: {
        template: 'carbon-multi-page',
        options: {
            queryGenerationFactories: {
                'GoMap': '{}',
                'GoArray': '[]',
                'RawString': '',
            }
        }
    },
}
```
