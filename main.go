package main

import (
	"log"

	"github.com/vitalvas/cassandra-redis-proxy/app"
)

func main() {
	application := app.New()

	if err := application.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
