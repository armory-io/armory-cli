package deploy

import (
	"github.com/armory/armory-cli/internal/deng"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"testing"
)

func TestParseLocalKustomization(t *testing.T) {
	p := &parser{
		args: []string{"../../hack/test/kustomize"},
		dep:  &deng.Deployment{},
		fs:   &pflag.FlagSet{},
		log:  logrus.New(),
	}
	p.fs.Bool(ParameterKustomize, false, "")
	p.fs.Bool(ParameterLocal, false, "")
	p.fs.StringSlice(ParameterVersion, nil, "")
	err := p.fs.Parse([]string{"--" + ParameterKustomize})
	if !assert.Nil(t, err) {
		return
	}
	err = p.parseKubernetesArtifacts()
	assert.Nil(t, err)
	// we expect 3 artifacts produced - see hack/test/kustomize
	assert.Equal(t, 3, len(p.dep.Artifacts))
}

func TestParseLocalManifests(t *testing.T) {
	p := &parser{
		args: []string{"../../hack/test"},
		dep:  &deng.Deployment{},
		fs:   &pflag.FlagSet{},
		log:  logrus.New(),
	}
	p.fs.Bool(ParameterKustomize, false, "")
	p.fs.Bool(ParameterLocal, false, "")
	p.fs.StringSlice(ParameterVersion, nil, "")
	err := p.fs.Parse(nil)
	if !assert.Nil(t, err) {
		return
	}
	err = p.parseKubernetesArtifacts()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(p.dep.Artifacts))
}

func TestParseSingleManifests(t *testing.T) {
	p := &parser{
		args: []string{"../../hack/test/deployment.yaml"},
		dep:  &deng.Deployment{},
		fs:   &pflag.FlagSet{},
		log:  logrus.New(),
	}
	p.fs.Bool(ParameterKustomize, false, "")
	p.fs.Bool(ParameterLocal, false, "")
	p.fs.StringSlice(ParameterVersion, nil, "")
	err := p.fs.Parse(nil)
	if !assert.Nil(t, err) {
		return
	}
	err = p.parseKubernetesArtifacts()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(p.dep.Artifacts))
}

func TestLocalKustomizeForceLocal(t *testing.T) {
	p := &parser{
		args: []string{"https://armory.jfrog.io/artifactory/manifests/kubesvc/armory-agent-0.5.7-kustomize.tar.gz"},
		dep:  &deng.Deployment{},
		fs:   &pflag.FlagSet{},
		log:  logrus.New(),
	}
	p.fs.Bool(ParameterKustomize, false, "")
	p.fs.Bool(ParameterLocal, false, "")
	p.fs.StringSlice(ParameterVersion, nil, "")
	err := p.fs.Parse([]string{"--" + ParameterKustomize, "--" + ParameterLocal})
	if !assert.Nil(t, err) {
		return
	}
	err = p.parseKubernetesArtifacts()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(p.dep.Artifacts))
}

func TestLargeSingleManifests(t *testing.T) {
	p := &parser{
		args: []string{"../../hack/test/large/large-cm.yaml"},
		dep:  &deng.Deployment{},
		fs:   &pflag.FlagSet{},
		log:  logrus.New(),
	}
	p.fs.Bool(ParameterKustomize, false, "")
	p.fs.Bool(ParameterLocal, false, "")
	p.fs.StringSlice(ParameterVersion, nil, "")
	err := p.fs.Parse(nil)
	if !assert.Nil(t, err) {
		return
	}
	err = p.parseKubernetesArtifacts()
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(p.dep.Artifacts)) {
		a := p.dep.Artifacts[0]
		an := a.GetKubernetes().GetManifests().GetSource().GetProvided()
		assert.NotNil(t, an)
		p := &deng.ArtifactPayload{}
		assert.Nil(t, an.UnmarshalTo(p))
		cm := v1.ConfigMap{}
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()
		_, _, err := decoder.Decode(p.Payload, nil, &cm)
		assert.Nil(t, err)
		assert.Equal(t, 10, len(cm.Data))
	}
}

func TestVersionOverride(t *testing.T) {
	p := &parser{
		args: []string{"../../hack/test/deployment.yaml"},
		dep:  &deng.Deployment{},
		fs:   &pflag.FlagSet{},
		log:  logrus.New(),
	}
	p.fs.Bool(ParameterKustomize, false, "")
	p.fs.Bool(ParameterLocal, false, "")
	p.fs.StringSlice(ParameterVersion, nil, "")
	err := p.fs.Parse([]string{"--version", "nginx:1.14.2"})
	if !assert.Nil(t, err) {
		return
	}
	err = p.parseKubernetesArtifacts()
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(p.dep.Artifacts)) {
		a := p.dep.Artifacts[0]
		an := a.GetKubernetes().GetManifests().GetSource().GetProvided()
		assert.NotNil(t, an)
		assert.Equal(t, 1, len(a.GetKubernetes().Versions))
		assert.Equal(t, "1.14.2", a.GetKubernetes().Versions["nginx"])
	}
}
