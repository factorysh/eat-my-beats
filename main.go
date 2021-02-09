package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func beats(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method, r.URL)
	fmt.Println(r.Header)
	b := r.Body
	if b != nil {
		io.Copy(os.Stdout, b)
	}
	w.Header().Add("Accept-Encoding", "gzip")
	w.Write([]byte("{}"))

}

func main() {
	http.HandleFunc("/", beats)
	log.Fatal(http.ListenAndServe(":9200", nil))
}
