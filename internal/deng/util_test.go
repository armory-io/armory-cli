package deng

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadOptionsFromJson(t *testing.T) {
	cases := []struct {
		provider string
		payload  string
		check    func(t2 *testing.T, err error, e *Environment)
	}{
		{
			KubernetesProvider,
			`{"namespace": "default"}`,
			func(t2 *testing.T, err error, e *Environment) {
				assert.Nil(t2, err)
				assert.Equal(t2, KubernetesProvider, e.Provider)
				k := e.GetKubernetes()
				assert.NotNil(t2, k)
				assert.Equal(t2, "default", k.Namespace)
			},
		},
		{
			// should still work and default to the environment provider if not provided
			"",
			`{"namespace": "default"}`,
			func(t2 *testing.T, err error, e *Environment) {
				assert.Nil(t2, err)
				assert.Equal(t2, KubernetesProvider, e.Provider)
				k := e.GetKubernetes()
				assert.NotNil(t2, k)
				assert.Equal(t2, "default", k.Namespace)
			},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			e := Environment{
				Provider: KubernetesProvider,
			}
			err := e.ReadOptionsFromJson(c.provider, []byte(c.payload))
			c.check(t, err, &e)
		})
	}
}

func TestConvertIntStr(t *testing.T) {
	assert.Nil(t, ConvertIntStr(nil))

	i := IntOrStringFromString("100")
	x := ConvertIntStr(&i)
	assert.NotNil(t, x)
	assert.Equal(t, "100", x.StrVal)

	i = IntOrStringFromInt(100)
	x = ConvertIntStr(&i)
	assert.NotNil(t, x)
	assert.Equal(t, int32(100), x.IntVal)

}
