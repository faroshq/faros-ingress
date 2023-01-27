package main

import (
	"context"
	"flag"
	"os"

	"k8s.io/klog/v2"

	"github.com/mjudeikis/portal/pkg/config"
	"github.com/mjudeikis/portal/pkg/servers/gateway"
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
	c, err := config.LoadGateway()
	if err != nil {
		return err
	}

	server, err := gateway.New(ctx, c)
	if err != nil {
		return err
	}
	return server.Run(ctx)
}