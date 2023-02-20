package server

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

func UsernameToUUID(
	username string,
) (
	UID,
	error,
) {
	url := fmt.Sprintf(
		"https://api.mojang.com/users/profiles/minecraft/%s",
		username,
	)
	resp, err := http.Get(url)
	if err != nil {
		return NilUID, err
	}
	defer resp.Body.Close()

	type JsonForm struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	}

	jsonForm := &JsonForm{}

	if err := json.NewDecoder(resp.Body).Decode(jsonForm); err != nil {
		return NilUID, err
	}

	return UID(uuid.MustParse(jsonForm.Id)), nil
}

func UUIDToTextureString(
	uid UID,
) (
	string,
	string,
	error,
) {
	uidString := uuid.UUID(uid).String()
	uidStringWithoutHyphens := strings.Replace(uidString, "-", "", -1)
	url := fmt.Sprintf(
		"https://sessionserver.mojang.com/session/minecraft/profile/%s?unsigned=false",
		uidStringWithoutHyphens,
	)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	type Property struct {
		Name      string `json:"name"`
		Value     string `json:"value"`
		Signature string `json:"signature"`
	}
	type JsonForm struct {
		Id         string      `json:"id"`
		Name       string      `json:"name"`
		Properties []*Property `json:"properties"`
	}

	jsonForm := &JsonForm{}

	if err := json.NewDecoder(resp.Body).Decode(jsonForm); err != nil {
		return "", "", err
	}

	property := jsonForm.Properties[0]
	textureString := property.Value
	signature := property.Signature

	return textureString, signature, nil
}
