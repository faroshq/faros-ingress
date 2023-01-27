package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/faroshq/faros-ingress/pkg/cliplugins/expose/plugin"
)

var (
	connectExample = `
	# Connect to faros ingress.
	%[1]s http://localhost:8080

	# Quick connect using name of existing or new connection
	%[1]s http://localhost:8080

	# Connect using token and connection id
	%[1]s http://localhost:8080 --token <token> --connection-id <connection-id>

	# Connect to specific localhost address and use basic auth
	%[1]s --secure
`
)

// New provides a cobra command for expose
func New(streams genericclioptions.IOStreams) (*cobra.Command, error) {
	// List command
	exposeOptions := plugin.NewExposeOptions(streams)
	exposeCmd := &cobra.Command{
		Use:          "expose",
		Short:        "expose local services with faros ingress",
		Example:      fmt.Sprintf(connectExample, "kubectl faros expose"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return c.Help()
			}
			if err := exposeOptions.Complete(args); err != nil {
				return err
			}

			if err := exposeOptions.Validate(); err != nil {
				return err
			}

			return exposeOptions.Run(c.Context())
		},
	}
	exposeOptions.BindFlags(exposeCmd)

	return exposeCmd, nil
}
