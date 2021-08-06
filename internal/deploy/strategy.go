package deploy

import (
	"fmt"
	"github.com/armory/armory-cli/internal/deng/protobuff"
	"strings"
)

const (
	strategyUpdate    = "update"
	strategyBlueGreen = "bluegreen"
	strategyCanary    = "canary"
)

func (p *parser) parseStrategy() error {
	s, err := p.fs.GetString(ParameterStrategy)
	if err != nil {
		return err
	}
	// TODO fill in more parameters
	switch strings.ToLower(s) {
	case strategyUpdate:
		p.dep.Strategy = map[string]*protobuff.Strategy{
			"default": {
				Type: &protobuff.Strategy_Update{
					Update: true,
				},
			},
		}
		return nil
	case strategyBlueGreen:
		p.dep.Strategy = map[string]*protobuff.Strategy{
			"default": {
				Type: &protobuff.Strategy_BlueGreen{
					BlueGreen: &protobuff.BlueGreen{
						ActiveService:  "active",
						PreviewService: "preview",
					},
				},
			},
		}
		return nil
	case strategyCanary:
		steps, err := p.parseCanarySteps()
		if err != nil {
			return err
		}
		p.dep.Strategy = map[string]*protobuff.Strategy{
			"default": {
				Type: &protobuff.Strategy_Canary{
					Canary: &protobuff.Canary{
						Steps: steps,
					},
				},
			},
		}
		return nil
	}
	return fmt.Errorf("unknown strategy %s", s)
}
