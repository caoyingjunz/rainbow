package main

import (
	"flag"

	"github.com/caoyingjunz/pixiulib/config"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/cmd/app/options"
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
	var cfg options.Config
	if err := c.Binding(&cfg); err != nil {
		klog.Fatal(err)
	}
	img := plugin.Image{
		Cfg: cfg,
	}
	if err := img.Complete(); err != nil {
		klog.Fatal(err)
	}
	defer img.Close()
	if err := img.Validate(); err != nil {
		klog.Fatal(err)
	}
	if err := img.PushImages(); err != nil {
		klog.Fatal(err)
	}
}
