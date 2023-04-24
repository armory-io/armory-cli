package auth

import (
	"encoding/json"
	"os"

	"gopkg.in/square/go-jose.v2/jwt"
)

const (
	armoryClaims string = "https://cloud.armory.io/principal"
)

type Credentials struct {
	Audience     string `json:"audience"`
	Source       string `json:"source"`
	ClientId     string `json:"clientId"`
	ExpiresAt    string `json:"expiresAt"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

func NewCredentials(audience, source, clientId, expiresAt, token string, refreshToken string) *Credentials {
	return &Credentials{
		Audience:     audience,
		Source:       source,
		ClientId:     clientId,
		ExpiresAt:    expiresAt,
		Token:        token,
		RefreshToken: refreshToken,
	}
}

func (c *Credentials) WriteCredentials(fileLocation string) error {
	data, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(fileLocation, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadCredentials(fileLocation string) (Credentials, error) {
	data, err := os.ReadFile(fileLocation)
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

func (c *Credentials) GetEnvironmentId() (string, error) {
	tok, err := jwt.ParseSigned(c.Token)
	if err != nil {
		return "", err
	}
	var claims map[string]interface{}
	err = tok.UnsafeClaimsWithoutVerification(&claims) //we've already obtained what we know to be a valid token from Auth0
	if err != nil {
		return "", err
	}
	return claims[armoryClaims].(map[string]interface{})["envId"].(string), nil
}

func (c *Credentials) GetOrganizationId() (string, error) {
	tok, _ := jwt.ParseSigned(c.Token)
	var claims map[string]interface{}
	err := tok.UnsafeClaimsWithoutVerification(&claims) //we've already obtained what we know to be a valid token from Auth0
	if err != nil {
		return "", err
	}
	return claims[armoryClaims].(map[string]interface{})["orgId"].(string), nil
}
