package main

import (
	"context"
	"log"
)

// Model Resolver Object
type Product struct {
	ID    int64                  `gql:"id"`
	Name  string                 `gql:"name"`
	Info  string                 `gql:"info"`
	Price float64                `gql:"price"`
	JSON2 map[string]interface{} `gql:"json2"`
	Other Other                  `gql:"other" `
}

type Other struct {
	P3 string     `gql:"p3"`
	P4 string     `gql:"p4"`
	P5 []string   `gql:"p5"`
	P6 []NewOther `gql:"p6"`
	P8 *NewOther  `gql:"p8"`
}

type NewOther struct {
	P7 string `gql:"p7"`
}

// NameResolver Overriding
type ProductNameArgs struct {
	Test  [5]string         `gql:"test"`
	Test2 map[string]string `gql:"test2"`
}

func (product *Product) PreResolver(ctx context.Context) context.Context {
	log.Println("invoke preResolver")
	return ctx
}

func (product *Product) GGL_Name(ctx context.Context, args *ProductNameArgs) (string, error) {
	return product.Name + "-with-resolver", nil
}

// PriceResolver Overriding
type ProductPriceArgs struct {
	Multiply int `gql:"multiply"`
}

func (product *Product) GGL_Price(ctx context.Context, args *ProductPriceArgs) (float64, error) {
	return product.Price * float64(args.Multiply), nil
}

// PriceInteger Resolver
func (product *Product) GGL_PriceInteger(ctx context.Context) (int64, error) {
	return 100, nil
}

/* RootResolver Object */
type Resolver struct {
}

func New() *Resolver {
	return new(Resolver)
}

/* ProductResolver scalar type */
type ProductArgs struct {
	Test bool `root:"_data"`
	ID   int  `gql:"id"`
}

func (resolver *Resolver) Product(ctx context.Context, args *ProductArgs) (*Product, error) {
	// do the data fetching via service / db call
	product := new(Product)
	product.ID = 1
	product.Price = 3.2
	product.Info = "asdokaskodsa"
	product.Name = "someproductname"

	product.Other.P3 = "sadasd"
	product.Other.P4 = "pomiomo"

	product.Other.P5 = append(product.Other.P5, "test")
	product.Other.P5 = append(product.Other.P5, "oska")

	product.Other.P6 = append(product.Other.P6, NewOther{"a"})
	product.Other.P6 = append(product.Other.P6, NewOther{"c"})

	product.JSON2 = make(map[string]interface{})
	product.JSON2["a"] = true
	product.JSON2["b"] = 3
	return product, nil
}

type Pagination struct {
	List   []*Product `gql:"list"`
	Cursor string     `gql:"cursor"`
}

/* ProductsResolver scalar type */
type ProductListArgs struct {
	Cursor string `gql:"cursor"`
	Type   string `gql:"type"`
}

func (resolver *Resolver) Products(ctx context.Context, args *ProductListArgs) (*Pagination, error) {
	products := make([]*Product, 0)

	product := new(Product)
	product.ID = 1
	product.Price = 3.2
	product.Info = "asdokaskodsa"
	product.Name = "someproductname"

	product.Other.P3 = "sadasd"
	product.Other.P4 = "pomiomo"

	product.Other.P5 = append(product.Other.P5, "test")
	product.Other.P5 = append(product.Other.P5, "oska")

	product.Other.P6 = append(product.Other.P6, NewOther{"a"})
	product.Other.P6 = append(product.Other.P6, NewOther{"c"})

	product.JSON2 = make(map[string]interface{})
	product.JSON2["a"] = true
	product.JSON2["b"] = 3

	products = append(products, product)
	return &Pagination{products, args.Cursor}, nil
}
