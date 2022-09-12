package config

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestConfigurationTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigurationTestSuite))
}

type ConfigurationTestSuite struct {
	suite.Suite
}

func (suite *ConfigurationTestSuite) TestParse() {
	cases := []struct {
		in     string
		err    string
		scheme string
		host   string
	}{
		{
			"http://127.0.0.1:8080",
			"",
			"http",
			"127.0.0.1:8080",
		},
		{
			"https://127.0.0.1:8080",
			"",
			"https",
			"127.0.0.1:8080",
		},
		{
			"https://localhost:8080",
			"",
			"https",
			"localhost:8080",
		},
		{
			"https://api.cloud.armory.io",
			"",
			"https",
			"api.cloud.armory.io",
		},
		{
			"https://api.cloud.armory.io/",
			"",
			"https",
			"api.cloud.armory.io",
		},
		{
			"http://127.0.0.1:8080?asdfasdf",
			"",
			"http",
			"127.0.0.1:8080",
		},
		{
			"api.cloud.armory.io",
			"expected url to contain scheme http or https, provided addr: api.cloud.armory.io",
			"",
			"",
		},
		{
			"https://",
			"expected url to contain a host, provided addr: https://",
			"",
			"",
		},
		{
			"https://api.cloud.armory.io/someBaseUrl",
			"expected url to not contain a path, provided addr: https://api.cloud.armory.io/someBaseUrl",
			"",
			"",
		},
		{
			"dssdf://asdfasdf:asdf",
			"failed to parse supplied Armory Cloud address, provided addr: dssdf://asdfasdf:asdf, thrown error: parse \"dssdf://asdfasdf:asdf\": invalid port \":asdf\" after host",
			"",
			"",
		},
	}

	for _, c := range cases {
		conf := New(&Input{ApiAddr: &c.in})
		parsedUrl, err := conf.getArmoryCloudAdder()

		if c.err != "" {
			suite.EqualErrorf(err, c.err, "Error messages do not match. Want: '%s', got: '%s'", c.err, err)
		} else {
			suite.Nil(err)
			suite.Equal(c.scheme, parsedUrl.Scheme)
			suite.Equal(c.host, parsedUrl.Host)
		}
	}
}
