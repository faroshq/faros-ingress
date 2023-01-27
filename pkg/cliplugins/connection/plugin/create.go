package plugin

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/faroshq/faros-ingress/pkg/api"
	"github.com/faroshq/faros-ingress/pkg/client"
	"github.com/faroshq/faros-ingress/pkg/cliplugins/base"
)

// CreateOptions contains options for configuring a connection
type CreateOptions struct {
	*base.Options
	// Name is the name of the Agent to be Created.
	Name string

	// Hostname is the hostname of the agent
	Hostname string

	// Secure is the flag to use secure connection with basic auth
	Secure bool
}

// NewCreateOptions returns a new CreateOptions.
func NewCreateOptions(streams genericclioptions.IOStreams) *CreateOptions {
	return &CreateOptions{
		Options: base.NewOptions(streams),
	}
}

// BindFlags binds fields CreateOptions as command line flags to cmd's flagset.
func (o *CreateOptions) BindFlags(cmd *cobra.Command) {
	o.Options.BindFlags(cmd)

	cmd.Flags().BoolVarP(&o.Secure, "secure", "s", false, "Secure with basic auth")
	cmd.Flags().StringVarP(&o.Hostname, "hostname", "", "", "Hostname of the agent")
}

// Complete ensures all dynamically populated fields are initialized.
func (o *CreateOptions) Complete(args []string) error {
	if err := o.Options.Complete(); err != nil {
		return err
	}

	o.Name = args[0]

	return nil
}

// Validate validates the SyncOptions are complete and usable.
func (o *CreateOptions) Validate() error {
	var errs []error

	if err := o.Options.Validate(); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// Run prepares an agent kubeconfig for use with a agent and outputs the
// configuration required to deploy a agent to remote agent
func (o *CreateOptions) Run(ctx context.Context) error {
	config, err := o.ClientConfig.ClientConfig()
	if err != nil {
		return err
	}

	u, err := url.Parse(config.Host)
	if err != nil {
		return err
	}

	c := client.NewClient(u, config.BearerToken, nil)

	conn, err := c.CreateConnection(ctx, api.Connection{
		Name:     o.Name,
		Secure:   o.Secure,
		Hostname: o.Hostname,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Connection %s created", conn.Name)
	fmt.Printf("\n")
	fmt.Printf("ID: '%s'", conn.ID)
	fmt.Printf("\n")
	fmt.Printf("Token: '%s'", conn.Identity)
	fmt.Printf("\n")
	fmt.Printf("Hostname: '%s'", conn.Hostname)
	fmt.Printf("\n")
	if conn.Secure {
		fmt.Printf("Username: '%s'\n", conn.Username)
		fmt.Printf("Password: '%s'", conn.Password)
		fmt.Printf("\n")
		fmt.Printf("Token and username with password will be shown only once. Please save it now\n\n")
	}

	return err
}
