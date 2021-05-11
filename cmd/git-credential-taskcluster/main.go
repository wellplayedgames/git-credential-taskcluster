package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"
	tcclient "github.com/taskcluster/taskcluster/v43/clients/client-go"
	"github.com/wellplayedgames/hive-runtime/pkg/logging"

	"github.com/wellplayedgames/git-credential-taskcluster/internal/credential"
	"github.com/wellplayedgames/git-credential-taskcluster/internal/credential/taskcluster"
)

var CLI struct {
	Verbose bool   `kong:"help='Increase logging verbosity'"`
	Command string `kong:"arg,help='The git-credential command to execute'"`

	SecretName string `kong:"help='The Taskcluster secret to read credentials from',env='TASKCLUSTER_GIT_SECRET',default='shared/git'"`

	TaskclusterProxyURL    string `kong:"help='A URL to a Taskcluster proxy to use',env='TASKCLUSTER_PROXY_URL'"`
	TaskclusterRootURL     string `kong:"help='The Taskcluster instance root URL',env='TASKCLUSTER_ROOT_URL'"`
	TaskclusterClientID    string `kong:"help='The Taskcluster client ID',env='TASKCLUSTER_CLIENT_ID'"`
	TaskclusterAccessToken string `kong:"help='The Taskcluster access token',env='TASKCLUSTER_ACCESS_TOKEN'"`
	TaskclusterCertificate string `kong:"help='The Taskcluster certificate',env='TASKCLUSTER_CERTIFICATE'"`
}

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	logger := logging.MustMakeLogger(
		logging.WithVerbosity(CLI.Verbose),
		logging.WithHumanReadableLogs(true),
	)

	kong.Parse(&CLI)

	helper := taskcluster.Helper{
		RootURL: CLI.TaskclusterRootURL,
		Credentials: tcclient.Credentials{
			ClientID:    CLI.TaskclusterClientID,
			AccessToken: CLI.TaskclusterAccessToken,
			Certificate: CLI.TaskclusterCertificate,
		},
		SecretName: CLI.SecretName,
		Logger:     logger,
	}

	if CLI.TaskclusterProxyURL != "" {
		helper.RootURL = CLI.TaskclusterProxyURL
		helper.Credentials = tcclient.Credentials{}
	}

	if helper.RootURL == "" {
		logger.Info("Taskcluster root URL is missing")
		os.Exit(1)
	}

	if err := credential.RunHelper(ctx, &helper, CLI.Command, os.Stdin, os.Stdout); err != nil {
		logger.Error(err, "error running helper command")
		os.Exit(1)
	}
}
