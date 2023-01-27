package main

import (
	"context"
	"flag"

	"github.com/mjudeikis/portal/pkg/config"
	"github.com/mjudeikis/portal/pkg/connector"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	opts := zap.Options{
		Development: true,
	}

	opts.BindFlags(flag.CommandLine)
	klog.InitFlags(flag.CommandLine)

	flag.Parse()
	flag.Lookup("v").Value.Set("6")

	ctx := context.Background()

	err := run(ctx)
	if err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	config, err := config.LoadConnector()
	if err != nil {
		return err
	}
	return runClient(ctx, config)
}

func runClient(ctx context.Context, config *config.ConnectorConfig) error {
	client, err := connector.New(config)
	if err != nil {
		return err
	}
	go client.Run(ctx)
	<-ctx.Done()
	return nil
}
