package deploy

import (
	"fmt"
	"github.com/armory/armory-cli/internal/deng"
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
		p.dep.Strategy = map[string]*deng.Strategy{
			"default": {
				Type: &deng.Strategy_Update{
					Update: true,
				},
			},
		}
		return nil
	case strategyBlueGreen:
		p.dep.Strategy = map[string]*deng.Strategy{
			"default": {
				Type: &deng.Strategy_BlueGreen{
					BlueGreen: &deng.BlueGreen{
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
		p.dep.Strategy = map[string]*deng.Strategy{
			"default": {
				Type: &deng.Strategy_Canary{
					Canary: &deng.Canary{
						Steps: steps,
					},
				},
			},
		}
		return nil
	}
	return fmt.Errorf("unknown strategy %s", s)
}
