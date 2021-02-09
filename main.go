package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/factorysh/eat-my-beats/eat"
)

func main() {
	f, err := os.OpenFile("./out.log", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	b := eat.New(f)
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go b.Start(ctx)
	http.Handle("/beats/", b.Mux)
	log.Fatal(http.ListenAndServe(":9200", nil))
}
