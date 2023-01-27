package plugin

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/mjudeikis/portal/pkg/api"
	"github.com/mjudeikis/portal/pkg/client"
	"github.com/mjudeikis/portal/pkg/cliplugins/base"
	"github.com/mjudeikis/portal/pkg/config"
	"github.com/mjudeikis/portal/pkg/connector"
	utilstrings "github.com/mjudeikis/portal/pkg/util/strings"
)

// ConnectOptions contains options for configuring a Agent and its corresponding process.
type ConnectOptions struct {
	*base.Options
	// Name is the name of the Agent to be connected too. If one does not exist, one
	// will be created.
	Name string
	// Secure is the flag to use secure connection with basic auth
	Secure bool
	// APIEndpoint is the URL of the API endpoint to connect to.
	APIEndpoint string
	// DownstreamURL is the URL of the downstream to connect to.
	DownstreamURL string
	// Token is the token to use for the connection.
	Token string
	// ConnectionID is the ID of the connection to use.
	ConnectionID string

	// create is set to true if the connection should be created.
	create bool
}

// NewConnectOptions returns a new ConnectOptions.
func NewConnectOptions(streams genericclioptions.IOStreams) *ConnectOptions {
	return &ConnectOptions{
		Options: base.NewOptions(streams),
	}
}

// BindFlags binds fields CreateOptions as command line flags to cmd's flagset.
func (o *ConnectOptions) BindFlags(cmd *cobra.Command) {
	o.Options.BindFlags(cmd)

	cmd.Flags().BoolVarP(&o.Secure, "secure", "s", false, "Secure with Basic Auth")
	cmd.Flags().StringVarP(&o.Name, "name", "", utilstrings.GetRandomName(), "Name of the connection")
	cmd.Flags().StringVarP(&o.DownstreamURL, "downstream", "d", "http://localhost:8080", "Downstream URL")
	cmd.Flags().StringVarP(&o.Token, "token", "t", "", "Token for the connection")
	cmd.Flags().StringVarP(&o.ConnectionID, "connection-id", "c", "", "Connection ID")
}

// Complete ensures all dynamically populated fields are initialized.
func (o *ConnectOptions) Complete(args []string) error {
	if err := o.Options.Complete(); err != nil {
		return err
	}

	if len(args) != 1 {
		o.Name = utilstrings.GetRandomName()
	} else {
		o.Name = args[0]
	}

	o.create = true // default - create new connection

	return nil
}

// Validate validates the SyncOptions are complete and usable.
func (o *ConnectOptions) Validate() error {
	var errs []error

	if err := o.Options.Validate(); err != nil {
		errs = append(errs, err)
	}

	if o.Token != "" && o.ConnectionID != "" { // if both are set, we should reuse
		o.create = false
	}

	return utilerrors.NewAggregate(errs)
}

// Run prepares an agent kubeconfig for use with a agent and outputs the
// configuration required to deploy a agent to remote agent
func (o *ConnectOptions) Run(ctx context.Context) error {
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
		fmt.Println("Basic Auth:")
		fmt.Println("Username: " + existing.Username)
		fmt.Println("Password: " + existing.Password)
	}
	<-ctx.Done()

	return nil
}
