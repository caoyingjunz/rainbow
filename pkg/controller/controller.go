package controller

import (
	"github.com/caoyingjunz/rainbow/pkg/controller/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/db"
)

type RainbowInterface interface {
	rainbow.AgentGetter
}

type rain struct {
	factory db.ShareDaoFactory
	name    string
}

func (p *rain) Agent() rainbow.Interface {
	return rainbow.NewAgent(p.factory, p.name)
}

func New(name string, f db.ShareDaoFactory) RainbowInterface {
	return &rain{
		factory: f,
		name:    name,
	}
}
