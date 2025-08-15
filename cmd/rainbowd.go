package main

import (
	"context"
	"flag"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

var (
	rainbowdFile = flag.String("configFile", "./config.yaml", "config file")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	opts, err := options.NewOptions(*rainbowdFile)
	if err != nil {
		klog.Fatal(err)
	}
	if err = opts.Complete(); err != nil {
		klog.Fatal(err)
	}

	for _, runner := range []func(context.Context, int) error{opts.Controller.Rainbowd().Run} {
		if err = runner(context.TODO(), 5); err != nil {
			klog.Fatal("failed to rainbowd: ", err)
		}
	}

	klog.Infof("rainbowd 已启动")
	select {}
}
