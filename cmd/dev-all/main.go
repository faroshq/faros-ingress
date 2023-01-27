package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"k8s.io/klog/v2"

	"github.com/faroshq/faros-ingress/pkg/config"
	devproxyclient "github.com/faroshq/faros-ingress/pkg/dev/client"
	"github.com/faroshq/faros-ingress/pkg/servers/api"
	"github.com/faroshq/faros-ingress/pkg/servers/gateway"
)

func main() {
	klog.InitFlags(flag.CommandLine)

	flag.Parse()
	flag.Lookup("v").Value.Set("6")

	ctx := context.Background()
	ctx = klog.NewContext(ctx, klog.NewKlogr())

	err := run(ctx)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {

	devVars := []string{
		"FAROS_API_CLUSTER_KUBECONFIG=faros.kubeconfig",
		"FAROS_API_TLS_KEY_FILE=dev/server.pem",
		"FAROS_API_TLS_CERT_FILE=dev/server.pem",
		"FAROS_GATEWAY_TLS_CERT_FILE=dev/server.pem",
		"FAROS_GATEWAY_TLS_KEY_FILE=dev/server.pem",
		"FAROS_API_HOSTNAME_SUFFIX=apps.dev.faros.sh",
		"FAROS_API_DEFAULT_GATEWAY=https://localhost:8444",
		"FAROS_OIDC_ISSUER_URL=https://dex.dev.faros.sh",
		"FAROS_GATEWAY_INTERNAL_GATEWAY_URL=https://localhost:8444",
	}

	for _, v := range devVars {
		parts := strings.Split(v, "=")
		err := os.Setenv(parts[0], parts[1])
		if err != nil {
			return err
		}
	}

	configAPI, err := config.LoadAPI()
	if err != nil {
		return err
	}

	configGateway, err := config.LoadGateway()
	if err != nil {
		return err
	}

	gateway, err := gateway.New(ctx, configGateway)
	if err != nil {
		return err
	}

	api, err := api.New(ctx, configAPI)
	if err != nil {
		return err
	}

	clientAPI, err := devproxyclient.New("https://localhost:30443", "https://localhost:8443", "dev/proxy-client.crt", "dev/proxy-client.key", "faros-dev")
	if err != nil {
		return err
	}

	clientGateway, err := devproxyclient.New("https://localhost:30444", "https://localhost:8444", "dev/proxy-client.crt", "dev/proxy-client.key", "faros-dev")
	if err != nil {
		return err
	}

	go gateway.Run(ctx)
	go api.Run(ctx)

	go clientAPI.Run(ctx)
	go clientGateway.Run(ctx)

	<-ctx.Done()

	return nil
}
