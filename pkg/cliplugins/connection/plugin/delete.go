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
)

// DeleteOptions contains options for configuring a Agent and its corresponding process.
type DeleteOptions struct {
	*base.Options

	Names []string
}

// NewDeleteOptions returns a new DeleteOptions.
func NewDeleteOptions(streams genericclioptions.IOStreams) *DeleteOptions {
	return &DeleteOptions{
		Options: base.NewOptions(streams),
	}
}

// BindFlags binds fields DeleteOptions as command line flags to cmd's flagset.
func (o *DeleteOptions) BindFlags(cmd *cobra.Command) {
	o.Options.BindFlags(cmd)
}

// Complete ensures all dynamically populated fields are initialized.
func (o *DeleteOptions) Complete(args []string) error {
	if err := o.Options.Complete(); err != nil {
		return err
	}

	o.Names = args

	return nil
}

// Validate validates the SyncOptions are complete and usable.
func (o *DeleteOptions) Validate() error {
	var errs []error

	if err := o.Options.Validate(); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// Run runs the DeleteOptions.
func (o *DeleteOptions) Run(ctx context.Context) error {
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
		for _, name := range o.Names {
			if conn.Name == name {
				err = c.DeleteConnection(ctx, conn)
				if err != nil {
					return err
				}
				fmt.Printf("Connection '%s' deleted \n", name)
			}
		}
	}

	return nil
}
