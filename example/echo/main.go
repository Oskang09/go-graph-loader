package main

import (
	ggl "github.com/Oskang09/go-graph-loader"
	"github.com/graphql-go/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	resolver := New()

	manager := ggl.New()
	err := manager.RegisterSchema(resolver)
	if err != nil {
		panic(err)
	}

	// manager.WriteSchema("schema.json")
	// manager.WriteMagidoc("magidoc.mjs", "schema.json")

	schema := manager.GetSchema()
	graphHandler := handler.New(&handler.Config{
		Schema:     &schema,
		GraphiQL:   false,
		Pretty:     true,
		Playground: true, // Work on GraphQL Playground with "http://127.0.0.1:2024/graphql"
	})

	server := echo.New()
	server.Use(middleware.Logger())
	server.Use(middleware.Recover())

	server.POST("/graphql", echo.WrapHandler(graphHandler))
	server.Logger.Fatal(server.Start(":2024"))

}
