package testing

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"time"
)

const aLongLongTimeAgo = 233431200
const armoryClaims = "https://cloud.armory.io/principal"

func CreateFakeJwt() (string, error) {
	armoryCustomClaims := map[string]interface{}{
		"envId": "12345",
		"orgId": "xyz",
	}
	t := jwt.New()
	t.Set(jwt.SubjectKey, `armory-cli`)
	t.Set(jwt.AudienceKey, `http://localhost`)
	t.Set(jwt.IssuedAtKey, time.Unix(aLongLongTimeAgo, 0))
	t.Set(armoryClaims, armoryCustomClaims)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", fmt.Errorf("failed to generate private key: %s", err)
	}
	jwkKey, err := jwk.New(key)
	if err != nil {
		return "", fmt.Errorf("failed to create JWK key: %s", err)
	}
	signed, err := jwt.Sign(t, jwa.RS256, jwkKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %s", err)
	}
	return string(signed), nil
}
