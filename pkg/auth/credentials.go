package auth

import (
	"encoding/json"
	"io/ioutil"
)

type Credentials struct {
	Audience  string `json:"audience"`
	Source    string `json:"source"`
	ClientId  string `json:"clientId"`
	ExpiresAt string `json:"expiresAt"`
	Token     string `json:"token"`
}

func NewCredentials(audience, source, clientId, expiresAt, token string) *Credentials {
	return &Credentials{
		Audience:  audience,
		Source:    source,
		ClientId:  clientId,
		ExpiresAt: expiresAt,
		Token:     token,
	}
}

func (c *Credentials) WriteCredentials(fileLocation string) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileLocation, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadCredentials(fileLocation string) (Credentials, error) {
	data, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return Credentials{}, err
	}
	credentials := Credentials{}
	err = json.Unmarshal(data, &credentials)
	if err != nil {
		return Credentials{}, err
	}
	return credentials, nil
}
