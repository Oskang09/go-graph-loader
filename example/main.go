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
	manager.Schema(resolver)
	result := manager.Do().
		Query(`{ 
			products(cursor: "some nested cursor", type: "HA") {
				list {
					info, name
				}
				cursor
			}
			product(id: 1) {
				info, name,
				price(multiply: 100), 
				priceInteger,
				other { 
					p5, 
					p6 {
						p7
					},
					p8
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
