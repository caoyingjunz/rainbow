package main

import (
	"flag"

	"github.com/caoyingjunz/pixiulib/config"
	"k8s.io/klog/v2"

	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
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

}
