package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/lestrrat-go/jwx/jwt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Auth struct {
	clientId       string `yaml:"clientId,omitempty" json:"clientId,omitempty"`
	secret         string `yaml:"secret,omitempty" json:"secret,omitempty"`
	tokenIssuerUrl string `yaml:"tokenIssuerUrl,omitempty" json:"tokenIssuerUrl,omitempty"`
	audience       string `yaml:"audience,omitempty" json:"audience,omitempty"`
	verify         bool   `yaml:"verify" json:"verify"`
	source         string `yaml:"source" json:"source"`
}

func NewAuth(clientId, clientSecret, source string) *Auth {
	return &Auth{
		clientId:       clientId,
		secret:         clientSecret,
		source:         source,
		tokenIssuerUrl: "__tokenIssuerUrl__",
		audience:       "__audience__",
		verify:         true,
	}
}

func (a *Auth) GetToken() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	exists, err := util.FileExists(dirname + "/.armory/credentials")
	if err != nil {
		return "", err
	}
	if exists {
		currentCreds, err := LoadCredentials(dirname + "/.armory/credentials")
		if err != nil {
			return "", err
		}

		expiresAt, err := time.Parse(time.RFC3339, currentCreds.ExpiresAt)
		if err != nil {
			return "", err
		}

		if time.Now().Before(expiresAt) && (a.clientId == "" || a.clientId == currentCreds.ClientId){
			return currentCreds.Token, nil
		}
	}

	if a.clientId == "" || a.secret == "" {
		return "", errors.New("no credentials set or expired, run armory login command or add clientId and clientSecret flags on the command")
	}

	token, expires, err := a.authentication(nil)
	if err != nil {
		return "", err
	}

	credentials := NewCredentials(a.audience, a.source, a.clientId, expires.Format(time.RFC3339), token)
	err = credentials.WriteCredentials(dirname + "/.armory/credentials")
	if err != nil {
		return "", err
	}

	return credentials.Token, nil
}

func (a *Auth) authentication(ctx context.Context) (string, *time.Time, error) {
	data := url.Values{}
	data.Set("grant_type", a.source)
	data.Set("client_id", a.clientId)
	data.Set("client_secret", a.secret)
	data.Set("audience", a.audience)
	req, err := http.NewRequest(http.MethodPost, a.tokenIssuerUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return "", nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("accept", "application/json")
	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return "", nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", nil, fmt.Errorf("unexpected status code while getting token %d", res.StatusCode)
	}
	defer res.Body.Close()
	tk, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", nil, err
	}

	rt := &remoteToken{}
	if err := json.Unmarshal(tk, rt); err != nil {
		return "", nil, fmt.Errorf("unable to parse response from %s: %w", a.tokenIssuerUrl, err)
	}
	if rt.AccessToken == "" {
		return "", nil, fmt.Errorf("no access_token returned from %s", a.tokenIssuerUrl)
	}

	t, err := jwt.Parse([]byte(rt.AccessToken), a.parseOptions()...)
	if err != nil {
		return "", nil, err
	}
	exp := t.Expiration()
	return rt.AccessToken, &exp, nil
}

func (a *Auth) parseOptions() []jwt.ParseOption {
	var opts []jwt.ParseOption
	if a.verify {
		opts = append(opts, jwt.WithValidate(true))
	}
	return opts
}

type remoteToken struct {
	AccessToken string `json:"access_token"`
}




