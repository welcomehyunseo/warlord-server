package server

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

func NewE0956(err error, username string) error {
	return fmt.Errorf("[E0956] err: %+v, username: %s", err, username)
}

func NewE0809(err error, username string) error {
	return fmt.Errorf("[E0809] err: %+v, username: %s", err, username)
}

func UsernameToPlayerID(username string) (uuid.UUID, error) {
	url := fmt.Sprintf("https://api.mojang.com/users/profiles/minecraft/%s", username)
	resp, err := http.Get(url)
	if err != nil {
		return uuid.Nil, NewE0956(err, username)
	}

	defer resp.Body.Close()

	type JsonForm struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	}

	jsonForm := &JsonForm{}

	if err := json.NewDecoder(resp.Body).Decode(jsonForm); err != nil {
		return uuid.Nil, NewE0809(err, username)
	}

	return uuid.MustParse(jsonForm.Id), nil
}
