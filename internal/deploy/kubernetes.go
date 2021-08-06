package deploy

import (
	"encoding/json"
	"fmt"
	"github.com/armory/armory-cli/internal/deng"
	"github.com/golang/protobuf/ptypes"
	"github.com/hashicorp/go-getter"
	"io"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func (p *parser) parseKubernetesArtifacts() error {
	// We could have multiple artifacts if we use local files
	// or a single artifact source that may translate into multiple
	// resolved artifacts on the server.
	via, err := p.parseVia()
	if err != nil {
		return err
	}

	ku, err := p.fs.GetBool(ParameterKustomize)
	if err != nil {
		return err
	}

	lcl, err := p.fs.GetBool(ParameterLocal)
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := p.parseVersionOverride(); err != nil {
		return err
	}

	// Check in arguments if there are any Kubernetes manifests
	for _, a := range p.args {
		if err := p.parseArgAsKubernetesManifest(a, ku, wd, via, lcl); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parseVersionOverride() error {
	vs, err := p.fs.GetStringSlice(ParameterVersion)
	if err != nil {
		return err
	}
	m := make(map[string]string)
	for _, v := range vs {
		ps := strings.Split(v, ":")
		if len(ps) != 2 {
			return fmt.Errorf("%s parameter should be in the format image:version (%s)", ParameterVersion, v)
		}
		m[ps[0]] = ps[1]
	}
	p.versions = m
	return nil
}

// parseArgAsKubernetesManifest interprets the parameter as a file, parses the Kubernetes manifest
// and passes it as an artifact. If a directory, all files in that directory are added as artifacts.
func (p *parser) parseArgAsKubernetesManifest(s string, kustomize bool, wd string, via *deng.Via, local bool) error {
	p.log.Debugf("parsing %s as Kubernetes manifest", s)

	src, err := getter.Detect(s, wd, getter.Detectors)
	if err != nil {
		return err
	}

	u, err := url.Parse(src)
	if err != nil {
		return err
	}

	// It's a local file or local was requested
	if u.Scheme == "file" {
		return p.addFromLocal(u.Path, s, kustomize)
	}

	if local {
		return p.resolveLocally(src, s, kustomize, wd)
	}

	if kustomize {
		p.dep.Artifacts = append(p.dep.Artifacts, &deng.Artifact{
			Provider: &deng.Artifact_Kubernetes{
				Kubernetes: &deng.KubernetesArtifact{
					Versions: p.versions,
					Type: &deng.KubernetesArtifact_Kustomize{
						Kustomize: &deng.KustomizeArtifact{
							Source: &deng.ArtifactSource{
								Url: src,
								Via: via,
							},
						},
					},
				},
			},
		})
	} else {
		p.dep.Artifacts = append(p.dep.Artifacts, &deng.Artifact{
			Provider: &deng.Artifact_Kubernetes{
				Kubernetes: &deng.KubernetesArtifact{
					Versions: p.versions,
					Type: &deng.KubernetesArtifact_Manifests{
						Manifests: &deng.ManifestsArtifact{
							Source: &deng.ArtifactSource{
								Url: src,
								Via: via,
							},
						},
					},
				},
			},
		})
	}
	return nil
}

func (p *parser) addFromLocal(src, name string, kustomize bool) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		if kustomize {
			panic("kustomize not supported!")
		} else {
			// Attempt to parse and add each file as a manifest
			files, err := ioutil.ReadDir(src)
			if err != nil {
				return err
			}
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				if err := p.parseFileAsKubernetesManifest(filepath.Join(src, file.Name()), name); err != nil {
					return err
				}
			}
			return nil
		}
	}
	if kustomize {
		panic("kustomize not supported!")
	}
	return p.parseFileAsKubernetesManifest(src, name)
}

func (p *parser) resolveLocally(src, name string, kustomize bool, wd string) error {
	dst, err := os.MkdirTemp("", "deploy")
	if err != nil {
		return err
	}

	// Clean up temp directory
	defer os.RemoveAll(dst)

	// Download the file
	if err := getter.Get(dst, src, func(c *getter.Client) error {
		c.Pwd = wd
		return nil
	}); err != nil {
		return err
	}

	return p.addFromLocal(dst, name, kustomize)
}

// parseFileAsKubernetesManifest parses a file as one or more Kubernetes artifact. Both JSON and YAML
// are supported. For YAML, multiple manifests can be added in the same file.
func (p *parser) parseFileAsKubernetesManifest(path, name string) error {
	p.log.Debugf("parsing file %s as Kubernetes manifest", path)
	var err error
	fr, err := os.Open(path)
	if err != nil {
		return err
	}
	return p.parseReaderAsKubernetesManifest(fr, name)
}

func (p *parser) parseReaderAsKubernetesManifest(reader io.Reader, name string) error {
	buffer, _, isJSON := yaml.GuessJSONStream(reader, 512)
	if isJSON {
		obj := unstructured.Unstructured{}
		if err := json.NewDecoder(buffer).Decode(&obj); err != nil {
			return err
		}
		return p.addKubernetesManifest(name, &obj)
	}

	// We may have multiple yaml representations in the same reader
	// split reader by frame
	decoder := yaml.NewYAMLToJSONDecoder(buffer)
	var err error

	for {
		obj := unstructured.Unstructured{}
		if err = decoder.Decode(&obj); err != nil {
			break
		}
		if err := p.addKubernetesManifest(name, &obj); err != nil {
			return err
		}
	}

	if err == io.EOF {
		err = nil
	}
	return err
}

func (p *parser) addKubernetesManifest(name string, un *unstructured.Unstructured) error {
	b, err := json.Marshal(un.Object)
	if err != nil {
		return err
	}

	any, err := ptypes.MarshalAny(&deng.ArtifactPayload{
		Payload: b,
	})

	if err != nil {
		return err
	}

	p.dep.Artifacts = append(p.dep.Artifacts, &deng.Artifact{
		Name: name,
		Provider: &deng.Artifact_Kubernetes{
			Kubernetes: &deng.KubernetesArtifact{
				Versions: p.versions,
				Type: &deng.KubernetesArtifact_Manifests{
					Manifests: &deng.ManifestsArtifact{
						Source: &deng.ArtifactSource{
							Provided: any,
						},
					},
				},
			},
		},
	})
	return nil
}
