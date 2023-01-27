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

// UpdateOptions contains options for configuring a Agent and its corresponding process.
type UpdateOptions struct {
	*base.Options
	// Name is the name of the Agent to be Updated.
	Name string
	// Username is the username of the Agent to be Updated.
	Username string
	// Password is the password of the Agent to be Updated.
	Password string
	// Hostname is the hostname of the Agent to be Updated.
	Hostname string
}

// NewUpdateOptions returns a new UpdateOptions.
func NewUpdateOptions(streams genericclioptions.IOStreams) *UpdateOptions {
	return &UpdateOptions{
		Options: base.NewOptions(streams),
	}
}

// BindFlags binds fields UpdateOptions as command line flags to cmd's flagset.
func (o *UpdateOptions) BindFlags(cmd *cobra.Command) {
	o.Options.BindFlags(cmd)

	cmd.Flags().StringVarP(&o.Username, "username", "u", "", "Username for agent")
	cmd.Flags().StringVarP(&o.Password, "password", "p", "", "Password for agent")
	cmd.Flags().StringVarP(&o.Hostname, "hostname", "n", "", "Hostname for agent")
}

// Complete ensures all dynamically populated fields are initialized.
func (o *UpdateOptions) Complete(args []string) error {
	if err := o.Options.Complete(); err != nil {
		return err
	}

	o.Name = args[0]

	return nil
}

// Validate validates the SyncOptions are complete and usable.
func (o *UpdateOptions) Validate() error {
	var errs []error

	if err := o.Options.Validate(); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// Run prepares an agent kubeconfig for use with a agent and outputs the
// configuration required to deploy a agent to remote agent
func (o *UpdateOptions) Run(ctx context.Context) error {
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

	for _, conn := range list.Items {
		if conn.Name == o.Name {
			conn.Username = o.Username
			conn.Password = o.Password
			conn.Hostname = o.Hostname

			_, err := c.UpdateConnection(ctx, conn)
			if err != nil {
				return err
			}

			fmt.Printf("Connection %s Updated", conn.Name)
			fmt.Printf("\n")
			if o.Hostname != "" {
				fmt.Printf("Hostname: '%s'", conn.Hostname)
				fmt.Printf("\n")
			}
			if o.Username != "" {
				fmt.Printf("Username: '%s'", conn.Username)
			}
			if o.Password != "" {
				fmt.Printf("Password: '%s'", conn.Password)
			}
			fmt.Printf("\n")
			fmt.Printf("Token and username with password will be shown only once. Please save it now\n\n")
		}
	}

	return err
}
