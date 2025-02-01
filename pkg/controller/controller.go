package controller

import (
	"github.com/caoyingjunz/rainbow/pkg/controller/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/db"
)

type RainbowInterface interface {
	rainbow.RainbowAgentGetter
}

type rain struct {
	factory db.ShareDaoFactory
	name    string
}

func (p *rain) RainbowAgent() rainbow.Interface {
	return rainbow.NewRainbowAgent(p.factory, p.name)
}

func New(name string, f db.ShareDaoFactory) RainbowInterface {
	return &rain{
		factory: f,
		name:    name,
	}
}
