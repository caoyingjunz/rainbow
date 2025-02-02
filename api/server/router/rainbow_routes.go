package router

import (
	"github.com/caoyingjunz/pixiulib/httputils"
	"github.com/gin-gonic/gin"
)

func (cr *rainbowRouter) createTask(c *gin.Context) {
	r := httputils.NewResponse()
	httputils.SetSuccess(c, r)
}

func (cr *rainbowRouter) updateTask(c *gin.Context) {}

func (cr *rainbowRouter) deleteTask(c *gin.Context) {}

func (cr *rainbowRouter) getTask(c *gin.Context) {}

func (cr *rainbowRouter) listTasks(c *gin.Context) {}
