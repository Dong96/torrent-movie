package model

import (
	"encoding/json"
	"fmt"
)

type ClientError struct {
	Root     error  `json:"-"`
	Response string `json:"error"`
	Status   int    `json:"-"`
}

func (e ClientError) Error() string {
	return e.Root.Error()
}

func (e ClientError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing response body: %v", err)
	}
	return body, nil
}
