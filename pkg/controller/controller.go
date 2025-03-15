package controller

import (
	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/controller/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/db"
)

type RainbowInterface interface {
	rainbow.ServerGetter
	rainbow.AgentGetter
}

type rain struct {
	factory db.ShareDaoFactory
	cfg     rainbowconfig.Config
}

func (p *rain) Server() rainbow.ServerInterface {
	return rainbow.NewServer(p.factory, p.cfg)
}

func (p *rain) Agent() rainbow.Interface {
	return rainbow.NewAgent(p.factory, p.cfg)
}

func New(cfg rainbowconfig.Config, f db.ShareDaoFactory) RainbowInterface {
	return &rain{
		factory: f,
		cfg:     cfg,
	}
}
