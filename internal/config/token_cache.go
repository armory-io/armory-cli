package config

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
)

const cacheFile = "tokens.json"

func getTokenFile() string {
	return path.Join(getHome(), cacheFile)
}

func readTokens(log *logrus.Logger) map[string]string {
	f := getTokenFile()
	fl, err := os.Open(f)
	if err != nil {
		if os.IsNotExist(err) {
			// Ignore, we're good
			return nil
		}
		log.WithError(err).Warn("unable to open token cache, ignoring")
		return nil
	}
	buf, err := ioutil.ReadAll(fl)
	if err != nil {
		log.WithError(err).Warn("unable to read token cache, deleting token cache")
		_ = os.Remove(f)
		return nil
	}

	m := make(map[string]string)
	if err := json.Unmarshal(buf, &m); err != nil {
		// file is corrupted, delete it
		log.WithError(err).Warn("unable to decode token cache, deleting token cache")
		_ = os.Remove(f)
	}
	return m
}

func getCurrentToken(log *logrus.Logger, account string) string {
	m := readTokens(log)
	if m != nil {
		return m[account]
	}
	return ""
}

func storeToken(log *logrus.Logger, account, token string) error {
	m := readTokens(log)
	m[account] = token
	b, err := json.Marshal(m)
	if err != nil {
		log.WithError(err).Warnf("unable to cache token for account %s", account)
		return err
	}
	return ioutil.WriteFile(getTokenFile(), b, 0600)
}
