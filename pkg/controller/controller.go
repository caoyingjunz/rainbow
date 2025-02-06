package controller

import (
	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/controller/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/db"
)

type RainbowInterface interface {
	rainbow.AgentGetter
	rainbow.ServerGetter
}

type rain struct {
	factory  db.ShareDaoFactory
	cfg      rainbowconfig.Config
	name     string
	callback string
}

func (p *rain) Agent() rainbow.Interface {
	return rainbow.NewAgent(p.factory, p.cfg, p.name, p.callback)
}

func (p *rain) Server() rainbow.ServerInterface {
	return rainbow.NewServer(p.factory)
}

func New(name string, callback string, cfg rainbowconfig.Config, f db.ShareDaoFactory) RainbowInterface {
	return &rain{
		factory:  f,
		cfg:      cfg,
		name:     name,
		callback: callback,
	}
}
