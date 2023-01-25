package server

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

func UsernameToPlayerID(username string) (uuid.UUID, error) {
	url := fmt.Sprintf("https://api.mojang.com/users/profiles/minecraft/%s", username)
	resp, err := http.Get(url)
	if err != nil {
		return uuid.Nil, err
	}

	defer resp.Body.Close()

	type JsonForm struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	}

	jsonForm := &JsonForm{}

	if err := json.NewDecoder(resp.Body).Decode(jsonForm); err != nil {
		return uuid.Nil, err
	}

	return uuid.MustParse(jsonForm.Id), nil
}
