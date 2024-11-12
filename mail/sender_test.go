package mail

import (
	"github.com/mariobasic/simplebank/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGmailSender_Send(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	config, err := util.LoadConfig("../")
	require.NoError(t, err)

	sender := NewGmailSender(config.Email.Sender.Name, config.Email.Sender.Address, config.Email.Sender.Password)

	subject := "A test email"
	content := `
	<h1>Hello world</h1>
	<p>This is test message</p>
`
	to := []string{"simplebank980@gmail.com"}
	attachFiles := []string{"../README.md"}

	err = sender.Send(subject, content, to, nil, nil, attachFiles)
	assert.NoError(t, err)
}
