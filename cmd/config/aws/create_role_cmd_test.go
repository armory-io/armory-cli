package aws

import (
	"bytes"
	"github.com/armory/armory-cli/pkg/config"
	clitesting "github.com/armory/armory-cli/pkg/testing"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getStdTestConfig(token string, outFmt string) *config.Configuration {
	return config.New(&config.Input{
		AccessToken:  &token,
		ApiAddr:      lo.ToPtr("https://localhost"),
		ClientId:     lo.ToPtr(""),
		ClientSecret: lo.ToPtr(""),
		OutFormat:    &outFmt,
		IsTest:       lo.ToPtr(true),
	})
}

func TestCreateRoleNoResponse(t *testing.T) {
	token, err := clitesting.CreateFakeJwt()
	assert.NoError(t, err)
	//r, w, err := os.Pipe()
	r := ClosingBuffer{
		bytes.NewBufferString("N\n"),
	}

	if err != nil {
		t.Errorf("Cannot create stdin pipe: %v", err)
	}
	//if _, err := w.WriteString(fmt.Sprintf("%v\n", "N")); err != nil {
	//	t.Errorf("Cannot write output to prompt: %v", err)
	//}
	cmd := NewCreateRoleCmd(getStdTestConfig(token, "text"), r)
	err = cmd.Execute()
	assert.NoError(t, err, "", "The prompt should exit without error")
}

type ClosingBuffer struct {
	*bytes.Buffer
}

func (cb ClosingBuffer) Close() error {
	return nil
}
