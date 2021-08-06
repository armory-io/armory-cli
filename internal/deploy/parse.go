package deploy

import (
	"fmt"
	"github.com/armory/armory-cli/internal/deng"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

const (
	ParameterEnvironmentName      = "account"
	ParameterEnvironmentType      = "account-type"
	ParameterEnvironmentNamespace = "namespace"
	ParameterKustomize            = "kustomize"
	ParameterLocal                = "local"
	ParameterViaAccount           = "via-account"
	ParameterViaProvider          = "via-provider"
	ParameterApplication          = "app"
	ParameterWait                 = "wait"
	ParameterVersion              = "version"

	// Strategy flags
	ParameterStrategy      = "strategy"
	ParameterStrategySteps = "canary-step"
)

func newParser(fs *pflag.FlagSet, args []string, log *logrus.Logger) *parser {
	return &parser{fs: fs, args: args, log: log, dep: &deng.Deployment{}}
}

type parser struct {
	fs       *pflag.FlagSet
	args     []string
	dep      *deng.Deployment
	log      *logrus.Logger
	versions map[string]string
}

func (p *parser) parse() (*deng.Deployment, error) {
	a, err := p.fs.GetString(ParameterApplication)
	if err != nil {
		return nil, err
	}
	p.dep.Application = a

	// Parse environment
	if err := p.parseEnvironment(); err != nil {
		return nil, err
	}

	switch p.dep.Environment.Provider {
	case deng.KubernetesProvider:
		// Parse artifacts now that we know the provider
		if err := p.parseKubernetesArtifacts(); err != nil {
			return nil, err
		}
	}

	if err := p.parseStrategy(); err != nil {
		return nil, err
	}

	return p.dep, nil
}

func (p *parser) parseEnvironment() error {
	t, err := p.fs.GetString(ParameterEnvironmentType)
	if err != nil {
		return err
	}
	switch t {
	case deng.KubernetesProvider:
		break
	default:
		return fmt.Errorf("unknown environment provider %s", t)
	}

	n, err := p.fs.GetString(ParameterEnvironmentName)
	if err != nil {
		return err
	}
	p.dep.Environment = &deng.Environment{
		Provider: t,
		Account:  n,
	}
	if t == deng.KubernetesProvider {
		ns, err := p.fs.GetString(ParameterEnvironmentNamespace)
		if err != nil {
			return err
		}
		p.dep.Environment.Qualifier = &deng.Environment_Kubernetes{
			Kubernetes: &deng.KubernetesQualifier{
				Namespace: ns,
			},
		}
	}
	return nil
}

func (p *parser) parseVia() (*deng.Via, error) {
	return nil, nil
}
