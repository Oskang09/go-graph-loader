package main

import (
	"context"
	"encoding/json"
	"log"

	ggl "github.com/Oskang09/go-graph-loader"
)

func main() {
	resolver := New()

	manager := ggl.New()
	err := manager.RegisterSchema(resolver)
	if err != nil {
		panic(err)
	}

	result := manager.Do().
		Query(`{ 
			products(cursor: "some nested cursor", type: "HA") {
				list {
					info, name
				}
				cursor
			}
			product(id: 1) {
				info, 
				name(
					test: "[\"1\"]",
					test2: "{\"a\":\"a\"}"
				),
				price(multiply: 100), 
				priceInteger,
				other { 
					p5, 
					p6 {
						p7
					},
					p8 {
						p7
					}
				}, 
				json2
			} 
		}`).
		Root(map[string]interface{}{"_data": true}).
		Execute(context.Background())

	if result.HasErrors() {
		log.Println(result.Errors)
	}
	bytes, _ := json.Marshal(result.Data)
	log.Println(string(bytes))
}
