package eat

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Beats struct {
	logs chan []*Action
}

func New() *Beats {
	return &Beats{
		logs: make(chan []*Action, 100),
	}
}

func (b *Beats) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case actions := <-b.logs:
			for _, action := range actions {
				fmt.Println(action)
			}
		}
	}
}

type CRUD int

const (
	Index CRUD = iota
	Create
	Delete
	Update
)

type Action struct {
	action    CRUD
	argument  map[string]interface{}
	HasSource bool
	Source    map[string]interface{}
}

func parseAction(line string) (*Action, error) {
	var raw map[string]map[string]interface{}
	err := json.Unmarshal([]byte(line), &raw)
	if err != nil {
		return nil, err
	}
	action := &Action{
		HasSource: true,
	}
	for k, v := range raw {
		switch k {
		case "index":
			action.action = Index
		case "create":
			action.action = Create
		case "delete":
			action.action = Delete
			action.HasSource = false
		case "update":
			action.action = Update
		default:
			return nil, fmt.Errorf("Unknown action %s", k)
		}
		action.argument = v
	}
	return action, nil
}

func (b *Beats) Handle(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method, r.URL)
	fmt.Println(r.Header)
	if r.Body != nil {
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
		actions := make([]*Action, 0)
		for scanner.Scan() {
			action, err := parseAction(scanner.Text())
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(500)
				return
			}
			if action.HasSource {
				scanner.Scan() // assert true
				s := make(map[string]interface{})
				err = json.Unmarshal([]byte(scanner.Text()), &s)
				if err != nil {
					fmt.Println(err)
					w.WriteHeader(500)
					return
				}
				action.Source = s
			}
			actions = append(actions, action)
		}
		b.logs <- actions
	}
	w.Header().Add("Accept-Encoding", "gzip")
	w.Write([]byte("{}"))
}
