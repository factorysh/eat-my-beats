package main

import (
	"context"
	"log"
	"net/http"

	"github.com/factorysh/eat-my-beats/eat"
)

func main() {
	b := eat.New()
	ctx := context.TODO()
	go b.Start(ctx)
	http.HandleFunc("/", b.Handle)
	log.Fatal(http.ListenAndServe(":9200", nil))
}
