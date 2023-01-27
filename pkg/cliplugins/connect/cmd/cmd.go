package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/mjudeikis/portal/pkg/cliplugins/connection/plugin"
)

var (
	connectExample = `
	# Connect to faros ingress.
	# set config file location
	KUBECONFIG=<config-file>

	# Connect using token and connection id
	%[1]s --token <token> --connection-id <connection-id>

	# Quick connect using name of existing or new connection
	%[1]s <connection-name>

	# Connect to specific localhost address
	%[1]s <connection-name> --downstream https://localhost:8443

	# Connect to specific localhost address and use basic auth
	%[1]s <connection-name> --secure
`
)

// New provides a cobra command for connect
func New(streams genericclioptions.IOStreams) (*cobra.Command, error) {
	// List command
	connectOptions := plugin.NewConnectOptions(streams)
	connectCmd := &cobra.Command{
		Use:          "connect",
		Short:        "connect to a connection",
		Example:      fmt.Sprintf(connectExample, "kubectl faros connect"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {

			if err := connectOptions.Complete(args); err != nil {
				return err
			}

			if err := connectOptions.Validate(); err != nil {
				return err
			}

			return connectOptions.Run(c.Context())
		},
	}
	connectOptions.BindFlags(connectCmd)

	return connectCmd, nil
}
