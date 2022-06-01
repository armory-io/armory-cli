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

const (
	expLeewaySec int64 = 300
)

type Auth struct {
	clientId             string
	secret               string
	tokenIssuerUrl       string
	audience             string
	source               string
	token                string
	memCachedCredentials *Credentials
}

func NewAuth(clientId, clientSecret, source, tokenIssuerUrl, audience, token string) *Auth {
	return &Auth{
		clientId:       clientId,
		secret:         clientSecret,
		source:         source,
		tokenIssuerUrl: tokenIssuerUrl,
		audience:       audience,
		token:          token,
	}
}

func (a *Auth) GetToken() (string, error) {
	if a.token != "" {
		return a.token, nil
	}

	if os.Getenv("CI") == "true" {
		creds, err := a.getTokenForCI()
		if err != nil {
			return "", err
		}
		return creds.Token, nil
	}

	return a.getTokenForSystemUser()
}

func (a *Auth) getTokenForCI() (*Credentials, error) {
	if a.memCachedCredentials != nil {
		return a.memCachedCredentials, nil
	}

	if a.clientId == "" || a.secret == "" {
		return nil, errors.New("no credentials set or expired, run armory login command or add clientId and clientSecret flags on the command")
	}

	token, expires, err := a.authentication(nil)
	if err != nil {
		return nil, err
	}
	a.memCachedCredentials = NewCredentials(a.audience, a.source, a.clientId, expires.Format(time.RFC3339), token, "")
	return a.memCachedCredentials, nil
}

func (a *Auth) getTokenForSystemUser() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(dirname + "/.armory"); os.IsNotExist(err) {
		err := os.Mkdir(dirname+"/.armory", os.ModePerm)
		if err != nil {
			return "", err
		}
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

		if time.Now().Add(time.Duration(expLeewaySec)*time.Second).Before(expiresAt) &&
			(a.clientId == "" || a.clientId == currentCreds.ClientId) {
			return currentCreds.Token, nil
		}
	}

	if a.clientId == "" || a.secret == "" {
		return "", errors.New("no credentials set; please login or add clientId and clientSecret flags to the command")
	}

	token, expires, err := a.authentication(nil)
	if err != nil {
		return "", err
	}

	credentials := NewCredentials(a.audience, a.source, a.clientId, expires.Format(time.RFC3339), token, "")
	err = credentials.WriteCredentials(dirname + "/.armory/credentials")
	if err != nil {
		return "", err
	}

	return credentials.Token, nil
}

func (a *Auth) GetEnvironmentId() (string, error) {
	if a.token != "" {
		return NewCredentials("", "", "", "", a.token, "").GetEnvironmentId()
	}

	if os.Getenv("CI") == "true" {
		creds, err := a.getTokenForCI()
		if err != nil {
			return "", err
		}
		return creds.Token, nil
	}

	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	currentCreds, err := LoadCredentials(dirname + "/.armory/credentials")

	if err != nil {
		return "", err
	}
	return currentCreds.GetEnvironmentId()
}

func (a *Auth) authentication(ctx context.Context) (string, *time.Time, error) {
	if a.token != "" {
		return "", nil, errors.New("do not try to execute remote authentication when a Token has been provided to the command")
	}
	data := url.Values{}
	data.Set("grant_type", a.source)
	data.Set("client_id", a.clientId)
	data.Set("client_secret", a.secret)
	data.Set("audience", a.audience)
	req, err := http.NewRequest(http.MethodPost, a.tokenIssuerUrl+"/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("accept", "application/json")
	c := &http.Client{
		Timeout: time.Second * 10,
	}
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

	parsedJwt, err := jwt.Parse([]byte(rt.AccessToken))
	if err != nil {
		return "", nil, err
	}
	exp := parsedJwt.Expiration()
	return rt.AccessToken, &exp, nil
}

type remoteToken struct {
	AccessToken string `json:"access_token"`
}
