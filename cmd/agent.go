package main

import (
	"context"
	"flag"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

var (
	filePath = flag.String("configFile", "./config.yaml", "config file")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	opts, err := options.NewOptions(*filePath)
	if err != nil {
		klog.Fatal(err)
	}
	if err = opts.Complete(); err != nil {
		klog.Fatal(err)
	}

	runers := []func(context.Context, int) error{opts.Controller.Agent().Run}
	for _, runner := range runers {
		if err = runner(context.TODO(), 5); err != nil {
			klog.Fatal("failed to rainbow agent: ", err)
		}
	}

	select {}
}
