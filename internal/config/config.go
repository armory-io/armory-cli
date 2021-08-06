package config

import (
	"fmt"
	"github.com/armory-io/go-cloud-service/pkg/client"
	"github.com/armory-io/go-cloud-service/pkg/token"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
)

const (
	deployCliEnvVarPrefix = "ARMORY"
	deployCliConfigDir    = ".armory"
	deployCliConfigFile   = "config.yaml"
)

type Config struct {
	Contexts       []Context `yaml:"contexts"`
	CurrentContext string    `yaml:"current-context"`
}

func (c *Config) getContext(name string) *Context {
	for _, ctx := range c.Contexts {
		if ctx.Name == name {
			return &ctx
		}
	}
	return nil
}

func (c *Config) checkCurrentContext() {
	if c.getContext(c.CurrentContext) == nil {
		if len(c.Contexts) > 0 {
			c.CurrentContext = c.Contexts[0].Name
		}
	}
}

type Context struct {
	Identity   token.Identity `yaml:"identity"`
	Name       string         `yaml:"name"`
	Connection client.Service `yaml:"connection"`
}

func (c *Context) NewConnection(log *logrus.Logger) client.Connection {
	return client.New(c.Connection, &c.Identity, log)
}

func getHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	home = path.Join(home, deployCliConfigDir)
	return home
}

func getConfigFile() string {
	if cfg := os.Getenv(fmt.Sprintf("%s_CONFIG", deployCliEnvVarPrefix)); cfg != "" {
		return cfg
	}
	return path.Join(getHome(), deployCliConfigFile)
}

// loadConfig loads configuration from the default or overridden location
// If withDefaults is specified, it applies default settings to each loaded
// context.
func loadConfig(withDefaults bool) (*Config, error) {
	f := getConfigFile()
	fl, err := os.Open(f)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	buf, err := ioutil.ReadAll(fl)
	if err != nil {
		_ = os.Remove(f)
		return nil, err
	}

	c := &Config{}
	if withDefaults {
		// We deserialize to a map
		m := make(map[string]interface{})
		if err := yaml.Unmarshal(buf, m); err != nil {
			return nil, err
		}

		// First decode non context attributes
		if err := mapstructure.Decode(m, c); err != nil {
			return nil, err
		}

		// Reset - we're rebuilding the slice with defaults
		c.Contexts = nil
		ctxs, ok := m["contexts"].([]interface{})
		if ok {
			for i := range ctxs {
				ctx := Context{
					Connection: defaultService(),
					Identity:   token.DefaultIdentity(),
				}
				if err := mapstructure.Decode(ctxs[i], &ctx); err != nil {
					return nil, err
				}
				c.Contexts = append(c.Contexts, ctx)
			}
		}

	} else {
		if err := yaml.Unmarshal(buf, c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// saveConfig saves the current configuration. It also checks the current context
// refers to an existing context.
func saveConfig(c *Config) error {
	f := getConfigFile()
	c.checkCurrentContext()
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f, b, 0600)
}

// defaultService returns default connection settings
func defaultService() client.Service {
	return client.Service{
		Grpc:                      "deploy.cloud.armory.io:443",
		KeepAliveHeartbeatSeconds: 30,
		KeepAliveTimeOutSeconds:   10,
	}
}
