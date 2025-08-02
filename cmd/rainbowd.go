package main

import (
	"flag"

	"github.com/caoyingjunz/pixiulib/config"
	"k8s.io/klog/v2"

	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
)

var (
	rainbowdFile = flag.String("configFile", "./config.yaml", "config file")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	c := config.New()
	c.SetConfigFile(*rainbowdFile)
	c.SetConfigType("yaml")

	var cfg rainbowconfig.Config
	if err := c.Binding(&cfg); err != nil {
		klog.Fatal(err)
	}
	cfg.Rainbowd.SetDefault()
}
