package taskcluster

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	tcclient "github.com/taskcluster/taskcluster/v43/clients/client-go"
	"github.com/taskcluster/taskcluster/v43/clients/client-go/tcsecrets"

	"github.com/wellplayedgames/git-credential-taskcluster/internal/credential"
)

type hostSecret struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type secretContents struct {
	Hosts map[string]*hostSecret `json:"hosts,omitempty"`
}

// Helper implements a git-credential helper for use in Taskcluster tasks
type Helper struct {
	credential.NullHelper
	tcclient.Credentials
	RootURL    string
	SecretName string
	Logger     logr.Logger
}

var _ credential.Helper = (*Helper)(nil)

func (h *Helper) Retrieve(ctx context.Context, input credential.HelperMessage) (credential.HelperMessage, error) {
	var ret credential.HelperMessage

	secrets := tcsecrets.New(&h.Credentials, h.RootURL)
	secret, err := secrets.Get(h.SecretName)
	if err != nil {
		return ret, err
	}

	var contents secretContents
	if err := json.Unmarshal(secret.Secret, &contents); err != nil {
		return ret, err
	}

	h.Logger.Info("Fetching credentials", "host", input.Host)

	host := contents.Hosts[input.Host]
	if host == nil {
		return ret, fmt.Errorf("unknown host: %s", input.Host)
	}

	ret.Username = host.Username
	ret.Password = host.Password
	return ret, nil
}
