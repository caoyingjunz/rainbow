package main

import (
	"flag"

	"github.com/caoyingjunz/pixiulib/config"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/controller/image"
)

var (
	filePath = flag.String("configFile", "./config.yaml", "config file")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	c := config.New()
	c.SetConfigFile(*filePath)
	c.SetConfigType("yaml")

	var cfg image.Config
	if err := c.Binding(&cfg); err != nil {
		klog.Fatal(err)
	}
	img := image.Image{
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
