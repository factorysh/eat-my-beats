package main

import (
	"log"
	"net/http"

	"github.com/factorysh/eat-my-beats/eat"
)

func main() {
	http.HandleFunc("/", eat.Beats)
	log.Fatal(http.ListenAndServe(":9200", nil))
}
