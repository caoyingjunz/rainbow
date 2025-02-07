package main

import (
	"flag"

	"github.com/caoyingjunz/pixiulib/config"
	"k8s.io/klog/v2"

	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/controller/plugin"
)

var (
	pluginFile = flag.String("configFile", "./config.yaml", "config file")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	c := config.New()
	c.SetConfigFile(*pluginFile)
	c.SetConfigType("yaml")

	var cfg rainbowconfig.Config
	if err := c.Binding(&cfg); err != nil {
		klog.Fatal(err)
	}
	pc := plugin.NewPluginController(cfg)
	if err := pc.Complete(); err != nil {
		klog.Fatal(err)
	}
	defer pc.Close()

	if err := pc.Validate(); err != nil {
		klog.Fatal(err)
	}

	if err := pc.Run(); err != nil {
		klog.Fatal(err)
	}
}
