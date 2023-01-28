package plugin

import (
	"context"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/faroshq/faros-ingress/pkg/client"
	"github.com/faroshq/faros-ingress/pkg/cliplugins/base"
	utilprint "github.com/faroshq/faros-ingress/pkg/util/print"
	utiltime "github.com/faroshq/faros-ingress/pkg/util/time"
)

// ListOptions contains options for configuring a Agent and its corresponding process.
type ListOptions struct {
	*base.Options
}

// NewListOptions returns a new ListOptions.
func NewListOptions(streams genericclioptions.IOStreams) *ListOptions {
	return &ListOptions{
		Options: base.NewOptions(streams),
	}
}

// BindFlags binds fields ListOptions as command line flags to cmd's flagset.
func (o *ListOptions) BindFlags(cmd *cobra.Command) {
	o.Options.BindFlags(cmd)
}

// Complete ensures all dynamically populated fields are initialized.
func (o *ListOptions) Complete(args []string) error {
	if err := o.Options.Complete(); err != nil {
		return err
	}

	return nil
}

// Validate validates the SyncOptions are complete and usable.
func (o *ListOptions) Validate() error {
	var errs []error

	if err := o.Options.Validate(); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// Run prepares an agent kubeconfig for use with a agent and outputs the
// configuration required to deploy a agent to remote agent
func (o *ListOptions) Run(ctx context.Context) error {
	config, err := o.ClientConfig.ClientConfig()
	if err != nil {
		return err
	}

	u, err := url.Parse(config.Host)
	if err != nil {
		return err
	}

	c := client.NewClient(u, config.BearerToken, nil)

	list, err := c.ListConnections(ctx)
	if err != nil {
		return err
	}

	if o.Output == utilprint.FormatTable {
		table := utilprint.DefaultTable()
		table.SetHeader([]string{"NAME", "HOSTNAME", "LAST USED", "TTL", "SECURE"})
		for _, conn := range list.Items {
			{
				table.Append([]string{
					conn.Name,
					conn.Hostname,
					utiltime.Since(conn.LastUsed).String() + " ago",
					conn.TTL.String(),
					strconv.FormatBool(conn.Secure),
				})
			}
		}
		table.Render()
		return nil
	}

	return utilprint.PrintWithFormat(list, o.Output)
}
