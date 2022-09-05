package main

import (
	"context"
)

// Model Resolver Object
type Product struct {
	ID    int64                  `ggl:"id"`
	Name  string                 `ggl:"name"`
	Info  string                 `ggl:"info"`
	Price float64                `ggl:"price"`
	JSON2 map[string]interface{} `ggl:"json2"`
	Other Other                  `ggl:"other" `
}

type Other struct {
	P3 string     `ggl:"p3"`
	P4 string     `ggl:"p4"`
	P5 []string   `ggl:"p5"`
	P6 []NewOther `ggl:"p6"`
	P8 *NewOther  `ggl:"p8"`
}

type NewOther struct {
	P7 string `ggl:"p7"`
}

// NameResolver Overriding
type ProductNameArgs struct {
}

func (product *Product) GGL_Name(ctx context.Context, args *ProductNameArgs) (string, error) {
	return product.Name + "-with-resolver", nil
}

// PriceResolver Overriding
type ProductPriceArgs struct {
	Multiply int `ggl:"multiply"`
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
	ID int `ggl:"id"`
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
	List   []*Product `ggl:"list"`
	Cursor string     `ggl:"cursor"`
}

/* ProductsResolver scalar type */
type ProductListArgs struct {
	Cursor string `ggl:"cursor"`
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
