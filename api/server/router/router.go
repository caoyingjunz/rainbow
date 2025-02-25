/*
Copyright 2021 The Pixiu Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

type RegisterFunc func(o *options.ServerOptions)

func InstallRouters(o *options.ServerOptions) {
	fs := []RegisterFunc{
		NewMiddlewares,
		NewRouter,
	}

	install(o, fs...)

	// 启动健康检查
	o.HttpEngine.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
}

func install(o *options.ServerOptions, fs ...RegisterFunc) {
	for _, f := range fs {
		f(o)
	}
}
