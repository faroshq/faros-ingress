package plugin

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/mjudeikis/portal/pkg/cliplugins/base"
	"github.com/mjudeikis/portal/pkg/models"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var kubeConfigAuthKey = "faros"

// LoginSetupOptions contains options for login via faros API
type LoginSetupOptions struct {
	*base.Options

	// ConfigFile of CLI config
	ConfigFile string

	// for testing
	modifyConfig func(configAccess clientcmd.ConfigAccess, newConfig *clientcmdapi.Config) error
}

// NewGenerateOptions returns a new GenerateOptions.
func NewLoginSetupOptions(streams genericclioptions.IOStreams) *LoginSetupOptions {
	return &LoginSetupOptions{
		Options: base.NewOptions(streams),
		modifyConfig: func(configAccess clientcmd.ConfigAccess, newConfig *clientcmdapi.Config) error {
			return clientcmd.ModifyConfig(configAccess, *newConfig, true)
		},
	}
}

// BindFlags binds fields GenerateOptions as command line flags to cmd's flagset.
func (o *LoginSetupOptions) BindFlags(cmd *cobra.Command) {
	o.Options.BindFlags(cmd)
}

// Complete ensures all dynamically populated fields are initialized.
func (o *LoginSetupOptions) Complete(args []string) error {
	if err := o.Options.Complete(); err != nil {
		return err
	}

	return nil
}

// Validate validates the inputs
func (o *LoginSetupOptions) Validate() error {
	var errs []error

	if err := o.Options.Validate(); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// Run prepares initiated login flow via IDP
func (o *LoginSetupOptions) Run(ctx context.Context) error {
	fmt.Println("Logging into Faros Hub...")

	doneCh := make(chan struct{})
	errCh := make(chan error)
	response := &models.LoginResponse{}

	// local server to catch the response
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		result, err := handleLoginCallback(req, w)
		if err != nil {
			errCh <- err
			return
		}
		response = result
		doneCh <- struct{}{}
	})

	l, err := getLocalListener()
	if err != nil {
		return err
	}

	// start serving locally and wait for the response
	go func() {
		if err := http.Serve(l, http.DefaultServeMux); err != nil {
			errCh <- fmt.Errorf("trying to start local http server: %s", err)
		}
	}()

	url := fmt.Sprintf("%s/api/v1alpha1/oidc/login?redirect_uri=http://localhost:%d", o.APIEndpoint, l.Addr().(*net.TCPAddr).Port)

	if err := open.Run(url); err != nil {
		return fmt.Errorf("trying to open web browser, error: %s", err)
	}

	// wait for the response
	select {
	case <-doneCh:
		return o.configureKubeConfig(ctx, *response)
	case err := <-errCh:
		return fmt.Errorf("trying to authorize the client: %s", err)

	case <-time.After(60 * time.Second):
		return errors.New("authorization request timed out waiting to complete")
	}

}

func (o *LoginSetupOptions) configureKubeConfig(ctx context.Context, response models.LoginResponse) error {
	config, err := o.ClientConfig.RawConfig()
	if err != nil {
		return err
	}

	// setup user
	user, exists := config.AuthInfos[kubeConfigAuthKey]
	if !exists {
		user = clientcmdapi.NewAuthInfo()
	}
	user.Token = response.RawIDToken
	config.AuthInfos[kubeConfigAuthKey] = user

	ca, err := base64.StdEncoding.DecodeString(response.CertificateAuthorityData)
	if err != nil {
		return err
	}

	// setup cluster
	config.Clusters[kubeConfigAuthKey] = &clientcmdapi.Cluster{
		Server: response.ServerBaseURL,
	}
	if response.CertificateAuthorityData != "" {
		config.Clusters[kubeConfigAuthKey].CertificateAuthorityData = ca
	} else {
		config.Clusters[kubeConfigAuthKey].InsecureSkipTLSVerify = true
	}
	config.Contexts[kubeConfigAuthKey] = &clientcmdapi.Context{
		Cluster:  kubeConfigAuthKey,
		AuthInfo: kubeConfigAuthKey,
	}
	config.CurrentContext = kubeConfigAuthKey

	fmt.Printf("Saving configuration. \n")

	return o.modifyConfig(o.ClientConfig.ConfigAccess(), &config)
}
