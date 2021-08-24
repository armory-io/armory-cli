package deploy

import (
	"errors"
	"github.com/armory/armory-cli/internal/deng/protobuff"
	"github.com/golang/protobuf/ptypes/wrappers"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	stepPause = "pause"
)

type stepParser func(str string) (*protobuff.Canary_CanaryStep, error)

func pauseParser(str string) (*protobuff.Canary_CanaryStep, error) {
	if str == stepPause {
		return &protobuff.Canary_CanaryStep{
			Step: &protobuff.Canary_CanaryStep_Pause{
				Pause: &protobuff.Canary_RolloutPause{},
			},
		}, nil
	}
	return nil, nil
}

var waitRe = regexp.MustCompile(`^wait{(\d+)(h|m|s)?}$`)

func waitParser(str string) (*protobuff.Canary_CanaryStep, error) {
	if !strings.HasPrefix(str, "wait{") {
		return nil, nil
	}
	m := waitRe.FindStringSubmatch(str)
	if len(m) == 0 {
		return nil, errors.New("specify a duration for the wait step followed by s(econds), m(inutes), or h(ours)")
	}

	var i protobuff.IntOrString
	if m[2] == "" {
		v, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, err
		}
		if v > math.MaxInt32 || v < math.MinInt32 {
			return nil, errors.New("value for wait stage out of in32 range")
		}
		i = protobuff.IntOrStringFromInt(int32(v))

	} else {
		i = protobuff.IntOrStringFromString(m[1] + m[2])
	}
	return &protobuff.Canary_CanaryStep{
		Step: &protobuff.Canary_CanaryStep_Wait{
			Wait: &protobuff.Canary_RolloutWait{
				Duration: &i,
			},
		},
	}, nil
}

var ratioRe = regexp.MustCompile(`^ratio{(\d+)}$`)

func ratioParser(str string) (*protobuff.Canary_CanaryStep, error) {
	if !strings.HasPrefix(str, "ratio{") {
		return nil, nil
	}
	m := ratioRe.FindStringSubmatch(str)
	if len(m) == 0 {
		return nil, errors.New("specify a ratio")
	}

	v, err := strconv.Atoi(m[1])
	if err != nil {
		return nil, err
	}
	if v < 0 || v > 100 {
		return nil, errors.New("ratio needs to be between 0 and 100")
	}

	w := wrappers.Int32Value{Value: int32(v)}
	return &protobuff.Canary_CanaryStep{
		Step: &protobuff.Canary_CanaryStep_SetWeight{
			SetWeight: &w,
		},
	}, nil
}

var stepParsers = []stepParser{pauseParser, waitParser, ratioParser}

func (p *parser) parseCanarySteps() ([]*protobuff.Canary_CanaryStep, error) {
	defs := p.deploymentConfiguration.StrategySteps
	r := make([]*protobuff.Canary_CanaryStep, 0)
	for _, s := range defs {
		for _, sp := range stepParsers {
			st, err := sp(s)
			if err != nil {
				return nil, err
			}
			if st != nil {
				r = append(r, st)
			}
		}
	}
	return r, nil
}
