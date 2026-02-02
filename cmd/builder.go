package main

import (
	"flag"

	"github.com/caoyingjunz/pixiulib/config"
	"k8s.io/klog/v2"

	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/controller/builder"
)

var (
	builderFile = flag.String("configFile", "./config.yaml", "config file")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	c := config.New()
	c.SetConfigFile(*builderFile)
	c.SetConfigType("yaml")

	var cfg rainbowconfig.Config
	if err := c.Binding(&cfg); err != nil {
		klog.Fatal(err)
	}

	bc := builder.NewBuilderController(cfg)
	if err := bc.Complete(); err != nil {
		klog.Fatal(err)
	}
	defer bc.Close()

	if err := bc.Run(); err != nil {
		klog.Fatal(err)
	}

}
