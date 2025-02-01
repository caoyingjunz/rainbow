package controller

import (
	"github.com/caoyingjunz/rainbow/pkg/controller/image"
	"github.com/caoyingjunz/rainbow/pkg/controller/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/db"
)

type RainbowInterface interface {
	rainbow.RainbowAgentGetter
}

type rain struct {
	factory db.ShareDaoFactory
}

func (p *rain) RainbowAgent() rainbow.Interface {
	return rainbow.NewRainbowAgent(p.factory)
}

func New(cfg image.Config, f db.ShareDaoFactory) RainbowInterface {
	return &rain{
		factory: f,
	}
}
