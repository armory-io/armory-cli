package deploy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParsePause(t *testing.T) {
	p, err := pauseParser("pause")
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, p.GetPause())
}

func TestParseWait(t *testing.T) {
	p, err := waitParser("wait")
	assert.Nil(t, err)
	assert.Nil(t, p)

	p, err = waitParser("wait{}")
	assert.NotNil(t, err)
	assert.Nil(t, p)

	// invalid unit
	p, err = waitParser("wait{10x}")
	assert.NotNil(t, err)
	assert.Nil(t, p)

	p, err = waitParser("wait{10m}")
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, p.GetWait())
	assert.Equal(t, "10m", p.GetWait().GetDuration().GetSValue())

	// this is also valid - should default to 10s
	p, err = waitParser("wait{10}")
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, p.GetWait())
	assert.Equal(t, int32(10), p.GetWait().GetDuration().GetIValue())
}

func TestParseRatio(t *testing.T) {
	p, err := ratioParser("not")
	assert.Nil(t, err)
	assert.Nil(t, p)

	p, err = ratioParser("ratio{}")
	assert.NotNil(t, err)
	assert.Nil(t, p)

	p, err = ratioParser("ratio{abc}")
	assert.NotNil(t, err)
	assert.Nil(t, p)

	p, err = ratioParser("ratio{105}")
	assert.NotNil(t, err)
	assert.Nil(t, p)

	p, err = ratioParser("ratio{10}")
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, int32(10), p.GetSetWeight().Value)
}
