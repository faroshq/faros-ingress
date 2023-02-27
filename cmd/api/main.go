package main

import (
	"context"
	"flag"
	"os"

	"k8s.io/klog/v2"

	"github.com/faroshq/faros-ingress/pkg/config"
	"github.com/faroshq/faros-ingress/pkg/servers/api"
)

func main() {
	klog.InitFlags(flag.CommandLine)

	flag.Parse()
	flag.Lookup("v").Value.Set("2")

	ctx := context.Background()
	ctx = klog.NewContext(ctx, klog.NewKlogr())

	err := run(ctx)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	c, err := config.LoadConfig()
	if err != nil {
		return err
	}

	server, err := api.New(ctx, c)
	if err != nil {
		return err
	}
	return server.Run(ctx)
}
