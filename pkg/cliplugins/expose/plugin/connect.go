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
	"github.com/faroshq/faros-ingress/pkg/config"
	"github.com/faroshq/faros-ingress/pkg/connector"
	utilstrings "github.com/faroshq/faros-ingress/pkg/util/strings"
)

// ExposeOptions contains options for use quick expose
type ExposeOptions struct {
	*base.Options
	// Name is the name of the Agent to be connected too. If one does not exist, one
	// will be created.
	Name string
	// Secure is the flag to use secure connection with basic auth
	Secure bool
	// DownstreamURL is the URL of the downstream to connect to.
	DownstreamURL string
	// Token is the token to use for the connection.
	Token string
	// ConnectionID is the ID of the connection to use.
	ConnectionID string

	// create is set to true if the connection should be created.
	create bool
}

// NewExposeOptions returns a new ExposeOptions.
func NewExposeOptions(streams genericclioptions.IOStreams) *ExposeOptions {
	return &ExposeOptions{
		Options: base.NewOptions(streams),
	}
}

// BindFlags binds fields CreateOptions as command line flags to cmd's flagset.
func (o *ExposeOptions) BindFlags(cmd *cobra.Command) {
	o.Options.BindFlags(cmd)

	cmd.Flags().BoolVarP(&o.Secure, "secure", "s", false, "Secure with Basic Auth")
	cmd.Flags().StringVarP(&o.Name, "name", "", utilstrings.GetRandomName(), "Name of the connection")
	cmd.Flags().StringVarP(&o.DownstreamURL, "downstream", "d", "http://localhost:8080", "Downstream URL")
	cmd.Flags().StringVarP(&o.Token, "token", "t", "", "Token for the connection")
	cmd.Flags().StringVarP(&o.ConnectionID, "connection-id", "c", "", "Connection ID")
}

// Complete ensures all dynamically populated fields are initialized.
func (o *ExposeOptions) Complete(args []string) error {
	if err := o.Options.Complete(); err != nil {
		return err
	}

	o.DownstreamURL = args[0]

	if o.Name == "" {
		o.Name = utilstrings.GetRandomName()
	}

	o.create = true // default - create new connection

	return nil
}

// Validate validates the expose options.
func (o *ExposeOptions) Validate() error {
	var errs []error

	if err := o.Options.Validate(); err != nil {
		errs = append(errs, err)
	}

	if o.Token != "" && o.ConnectionID != "" { // if both are set, we should reuse
		o.create = false
	}

	return utilerrors.NewAggregate(errs)
}

// Run prepares expose command and runs it.
func (o *ExposeOptions) Run(ctx context.Context) error {
	restConfig, err := o.ClientConfig.ClientConfig()
	if err != nil {
		return err
	}

	u, err := url.Parse(restConfig.Host)
	if err != nil {
		return err
	}

	c := client.NewClient(u, restConfig.BearerToken, nil)

	conns, err := c.ListConnections(ctx)
	if err != nil {
		return err
	}

	var existing *api.Connection
	var found bool
	for _, conn := range conns.Items {
		if conn.Name == o.Name {
			found = true
			existing = &conn
			break
		}
	}

	if !found && o.create {
		fmt.Printf("Creating connection: %s \n", o.Name)
		existing, err = c.CreateConnection(ctx, api.Connection{
			Name:   o.Name,
			Secure: o.Secure,
		})
		if err != nil {
			return err
		}
	}
	if !found && !o.create {
		return fmt.Errorf("connection %s not found", o.Name)
	}

	cfg, err := config.LoadConnector()
	if err != nil {
		return err
	}

	// override from flags
	cfg.ControllerURL = restConfig.Host
	cfg.DownstreamURL = o.DownstreamURL
	cfg.Token = existing.Identity
	cfg.ConnectionID = existing.ID

	client, err := connector.New(cfg)
	if err != nil {
		return err
	}
	go client.Run(ctx)

	fmt.Println("Connecting to connection: " + o.Name)
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println("")
	fmt.Println("URL: " + existing.Hostname + " --> " + o.DownstreamURL)
	fmt.Println("")
	if existing.Secure {
		fmt.Println("Basic auth:")
		fmt.Println("Username: " + existing.Username)
		fmt.Println("Password: " + existing.Password)
	}
	<-ctx.Done()

	return nil
}
