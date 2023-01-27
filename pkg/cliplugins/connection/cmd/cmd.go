package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/faroshq/faros-ingress/pkg/cliplugins/connection/plugin"
)

var (
	connectionExample = `
	# Ensure a agent is running on the specified connection target.
	%[1]s TBC
	KUBECONFIG=<pcluster-config> <agent_image>
`
)

// New provides a cobra command for workload operations.
func New(streams genericclioptions.IOStreams) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Aliases:          []string{"connection", "conn"},
		Use:              "connections",
		Short:            "Manages connections",
		SilenceUsage:     true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// List command
	listOptions := plugin.NewListOptions(streams)
	listCmd := &cobra.Command{
		Use:          "list",
		Short:        "List connections",
		Example:      fmt.Sprintf(connectionExample, "kubectl faros connections list"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := listOptions.Complete(args); err != nil {
				return err
			}

			if err := listOptions.Validate(); err != nil {
				return err
			}

			return listOptions.Run(c.Context())
		},
	}

	listOptions.BindFlags(listCmd)
	cmd.AddCommand(listCmd)

	// List command
	createOptions := plugin.NewCreateOptions(streams)
	createCmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a connection",
		Example:      fmt.Sprintf(connectionExample, "kubectl faros connection create"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) != 1 {
				return c.Help()
			}

			if err := createOptions.Complete(args); err != nil {
				return err
			}

			if err := createOptions.Validate(); err != nil {
				return err
			}

			return createOptions.Run(c.Context())
		},
	}

	createOptions.BindFlags(createCmd)
	cmd.AddCommand(createCmd)

	// Get command
	getOptions := plugin.NewGetOptions(streams)
	getCmd := &cobra.Command{
		Use:          "get",
		Short:        "Get a connection",
		Example:      fmt.Sprintf(connectionExample, "kubectl faros connection get"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) != 1 {
				return c.Help()
			}

			if err := getOptions.Complete(args); err != nil {
				return err
			}

			if err := getOptions.Validate(); err != nil {
				return err
			}

			return getOptions.Run(c.Context())
		},
	}

	getOptions.BindFlags(getCmd)
	cmd.AddCommand(getCmd)

	// Delete command
	deleteOptions := plugin.NewDeleteOptions(streams)
	deleteCmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete a agent",
		Example:      fmt.Sprintf(connectionExample, "kubectl faros agent connection"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) != 1 {
				return c.Help()
			}

			if err := deleteOptions.Complete(args); err != nil {
				return err
			}

			if err := deleteOptions.Validate(); err != nil {
				return err
			}

			return deleteOptions.Run(c.Context())
		},
	}

	deleteOptions.BindFlags(deleteCmd)
	cmd.AddCommand(deleteCmd)

	// Connect command
	connectOptions := plugin.NewConnectOptions(streams)
	connectCmd := &cobra.Command{
		Use:          "connect",
		Short:        "Connect using connector",
		Example:      fmt.Sprintf(connectionExample, "kubectl faros agent connection connect"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) != 1 {
				return c.Help()
			}

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
	cmd.AddCommand(connectCmd)

	return cmd, nil
}
