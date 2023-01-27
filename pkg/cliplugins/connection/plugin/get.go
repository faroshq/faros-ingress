package plugin

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/faroshq/faros-ingress/pkg/client"
	"github.com/faroshq/faros-ingress/pkg/cliplugins/base"
	utilprint "github.com/faroshq/faros-ingress/pkg/util/print"
)

// GetOptions contains options for configuring a Agent and its corresponding process.
type GetOptions struct {
	*base.Options

	Name string
}

// NewGetOptions returns a new GetOptions.
func NewGetOptions(streams genericclioptions.IOStreams) *GetOptions {
	return &GetOptions{
		Options: base.NewOptions(streams),
	}
}

// BindFlags binds fields GetOptions as command line flags to cmd's flagset.
func (o *GetOptions) BindFlags(cmd *cobra.Command) {
	o.Options.BindFlags(cmd)
}

// Complete ensures all dynamically populated fields are initialized.
func (o *GetOptions) Complete(args []string) error {
	if err := o.Options.Complete(); err != nil {
		return err
	}

	o.Name = args[0]

	return nil
}

// Validate validates the SyncOptions are complete and usable.
func (o *GetOptions) Validate() error {
	var errs []error

	if err := o.Options.Validate(); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// Run prepares an agent kubeconfig for use with a agent and outputs the
// configuration required to deploy a agent to remote agent
func (o *GetOptions) Run(ctx context.Context) error {
	config, err := o.ClientConfig.ClientConfig()
	if err != nil {
		return err
	}

	u, err := url.Parse(config.Host)
	if err != nil {
		return err
	}

	c := client.NewClient(u, config.BearerToken, nil)

	conns, err := c.ListConnections(ctx)
	if err != nil {
		return err
	}
	for _, conn := range conns.Items {
		if conn.Name == o.Name {
			return utilprint.PrintWithFormat(conn, "yaml")
		}
	}

	return fmt.Errorf("connection %q not found", o.Name)
}
