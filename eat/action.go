package eat

import (
	"encoding/json"
	"fmt"
)

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
