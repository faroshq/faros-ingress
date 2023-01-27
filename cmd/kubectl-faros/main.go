package main

import (
	goflags "flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/mjudeikis/portal/pkg/cliplugins/faros/cmd"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-faros", pflag.ExitOnError)
	pflag.CommandLine = flags

	farosCmd, err := cmd.New(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// setup klog
	fs := goflags.NewFlagSet("klog", goflags.PanicOnError)
	klog.InitFlags(fs)
	farosCmd.PersistentFlags().AddGoFlagSet(fs)

	if err := farosCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
