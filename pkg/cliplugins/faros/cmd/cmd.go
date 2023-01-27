package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	connectioncmd "github.com/faroshq/faros-ingress/pkg/cliplugins/connection/cmd"
	exposecmd "github.com/faroshq/faros-ingress/pkg/cliplugins/expose/cmd"
	logincmd "github.com/faroshq/faros-ingress/pkg/cliplugins/login/cmd"
)

// New returns a cobra.Command for faros actions.
func New(streams genericclioptions.IOStreams) (*cobra.Command, error) {
	connectionCmd, err := connectioncmd.New(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	exposeCmd, err := exposecmd.New(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	loginCmd, err := logincmd.New(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cmd := &cobra.Command{
		Use:   "faros",
		Short: "Manage faros ingress",
	}

	cmd.AddCommand(connectionCmd)
	cmd.AddCommand(exposeCmd)
	cmd.AddCommand(loginCmd)

	return cmd, nil
}
