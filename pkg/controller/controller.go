package controller

import (
	"github.com/caoyingjunz/rainbow/pkg/controller/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/db"
)

type RainbowInterface interface {
	rainbow.AgentGetter
	rainbow.ServerGetter
}

type rain struct {
	factory  db.ShareDaoFactory
	name     string
	callback string
}

func (p *rain) Agent() rainbow.Interface {
	return rainbow.NewAgent(p.factory, p.name, p.callback)
}

func (p *rain) Server() rainbow.ServerInterface {
	return rainbow.NewServer(p.factory)
}

func New(name string, callback string, f db.ShareDaoFactory) RainbowInterface {
	return &rain{
		factory:  f,
		name:     name,
		callback: callback,
	}
}
