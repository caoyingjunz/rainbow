package controller

import (
	"github.com/go-redis/redis/v8"

	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/controller/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/controller/rainbowd"
	"github.com/caoyingjunz/rainbow/pkg/db"
)

type RainbowInterface interface {
	rainbow.ServerGetter
	rainbow.AgentGetter
	rainbowd.RainbowdGetter
}

type rain struct {
	factory     db.ShareDaoFactory
	cfg         rainbowconfig.Config
	redisClient *redis.Client
}

func (p *rain) Server() rainbow.ServerInterface {
	return rainbow.NewServer(p.factory, p.cfg, p.redisClient)
}

func (p *rain) Agent() rainbow.Interface {
	return rainbow.NewAgent(p.factory, p.cfg, p.redisClient)
}

func (p *rain) Rainbowd() rainbowd.Interface {
	return rainbowd.New(p.factory, p.cfg)
}

func New(cfg rainbowconfig.Config, f db.ShareDaoFactory, redisClient *redis.Client) RainbowInterface {
	return &rain{
		factory:     f,
		cfg:         cfg,
		redisClient: redisClient,
	}
}
