package eat

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/valyala/fastjson"
)

type Beats struct {
	logs chan []*fastjson.Value
	Mux  *http.ServeMux
	out  io.Writer
}

func New(out io.Writer) *Beats {
	b := &Beats{
		logs: make(chan []*fastjson.Value, 100),
		Mux:  &http.ServeMux{},
		out:  out,
	}
	prefix := "/beats"
	if prefix != "" {
		b.Mux.HandleFunc("/", b.middleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		}))
	}
	b.Mux.HandleFunc(prefix+"/", b.middleware(b.home))
	b.Mux.HandleFunc(prefix+"/_bulk", b.middleware(b.bulk))
	b.Mux.HandleFunc(prefix+"/_template/", b.middleware(b.template))
	b.Mux.HandleFunc(prefix+"/_xpack/", b.middleware(b.xpack))
	b.Mux.HandleFunc(prefix+"/_component_template/", b.middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	return b
}

func (b *Beats) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case logs := <-b.logs:
			for _, log := range logs {
				fmt.Println("type", string(log.GetStringBytes("type")))
				if bytes.Compare(log.GetStringBytes("type"), []byte("redis")) == 0 {
					_, err := b.out.Write(log.MarshalTo(nil))
					if err != nil {
						fmt.Println(err)
						return err
					}
				}
			}
		}
	}
}

func (b *Beats) middleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request", r.Method, r.URL.Path, r.Header)
		w.Header().Add("Accept-Encoding", "gzip")
		h.ServeHTTP(w, r)
	}
}

func (b *Beats) home(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Header)
	if r.URL.Path != "/beats/" {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(`
	{
		"status" : 200,
		"name" : "Eat my beats",
		"cluster_name" : "eat-my-beats",
		"version" : {
		  "number" : "7.10.2",
		  "build_hash" : "",
		  "build_timestamp" : "",
		  "build_snapshot" : false,
		  "lucene_version" : "4.10.4"
		},
		"tagline" : "You Know, for Search"
	  }	
	
	`))
}

func (b *Beats) template(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(404)
		return
	}
	if r.Method == "PUT" {
		w.WriteHeader(200)
		return
	}
	w.Write([]byte("{}"))
}

func (b *Beats) xpack(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
}

func (b *Beats) bulk(w http.ResponseWriter, r *http.Request) {
	chrono := time.Now()
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil {
		w.WriteHeader(400)
		return
	}
	defer r.Body.Close()
	var reader io.Reader
	var err error
	if r.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(r.Body)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	} else {
		reader = r.Body
	}
	scanner := bufio.NewScanner(reader)
	responses, err := b.scan(scanner)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	resp := `{"took":%d, "status": 200, "errors": false, "items":[`
	fmt.Fprintf(w, resp, time.Since(chrono)/time.Second)
	w.Write([]byte(strings.Join(responses, ",")))
	w.Write([]byte("]}"))
}

func (b *Beats) scan(scanner *bufio.Scanner) ([]string, error) {
	var p fastjson.Parser
	logs := make([]*fastjson.Value, 0)
	responses := make([]string, 0)
	for scanner.Scan() {
		v, err := p.Parse(scanner.Text())
		if err != nil {
			return responses, err
		}
		if !v.Exists("delete") {
			ok := scanner.Scan()
			if !ok {
				return responses, errors.New("odd bulk")
			}
			if v.Exists("create") {
				// it's the index name
				//argument := v.Get("create")
				v, err := p.Parse(scanner.Text())
				if err != nil {
					return responses, err
				}
				logs = append(logs, v)
			}
			responses = append(responses, `{"index":{"status": 201}}`)
		}
	}
	b.logs <- logs
	return responses, nil
}
