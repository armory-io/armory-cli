package deploy

import (
	"fmt"
	"github.com/armory/armory-cli/internal/deng"
	"github.com/armory/armory-cli/internal/deng/protobuff"
)

func NewParser(deploymentConfiguration *deng.DeploymentConfiguration) *parser {
	return &parser{deploymentConfiguration: deploymentConfiguration, dep: &protobuff.Deployment{}}
}

type parser struct {
	deploymentConfiguration *deng.DeploymentConfiguration
	dep      *protobuff.Deployment
	versions map[string]string
}

func (p *parser) Parse() (*protobuff.Deployment, error) {
	application := p.deploymentConfiguration.Application
	p.dep.Application = application

	// Parse environment
	if err := p.parseEnvironment(); err != nil {
		return nil, err
	}

	switch p.dep.Environment.Provider {
	case protobuff.KubernetesProvider:
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
	environmentType := p.deploymentConfiguration.EnvironmentType
	switch environmentType {
	case protobuff.KubernetesProvider:
		break
	default:
		return fmt.Errorf("unknown environment provider %s", environmentType)
	}

	environmentName := p.deploymentConfiguration.EnvironmentName
	p.dep.Environment = &protobuff.Environment{
		Provider: environmentType,
		Account:  environmentName,
	}
	if environmentType == protobuff.KubernetesProvider {
		ns := p.deploymentConfiguration.EnvironmentNamespace
		p.dep.Environment.Qualifier = &protobuff.Environment_Kubernetes{
			Kubernetes: &protobuff.KubernetesQualifier{
				Namespace: ns,
			},
		}
	}
	return nil
}

func (p *parser) parseVia() (*protobuff.Via, error) {
	return nil, nil
}
