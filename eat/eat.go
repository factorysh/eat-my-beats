package eat

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
)

func Beats(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method, r.URL)
	fmt.Println(r.Header)
	b := r.Body
	if b != nil {
		if r.Header.Get("Content-Encoding") == "gzip" {
			g, err := gzip.NewReader(b)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			io.Copy(os.Stdout, g)
		} else {
			io.Copy(os.Stdout, b)
		}
	}
	w.Header().Add("Accept-Encoding", "gzip")
	w.Write([]byte("{}"))
}
