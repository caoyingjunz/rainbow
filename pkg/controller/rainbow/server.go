package rainbow

import (
	"github.com/caoyingjunz/rainbow/pkg/db"
)

type ServerGetter interface {
	Server() ServerInterface
}

type ServerInterface interface {
}

type ServerController struct {
	factory db.ShareDaoFactory
}

func NewServer(f db.ShareDaoFactory) *ServerController {
	return &ServerController{
		factory: f,
	}
}
