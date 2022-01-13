/**
 * REST API router
 * Rosbit Xu
 */
package main

import (
	"github.com/rosbit/mgin"
	"net/http"
	"fmt"
	"delay-tasks/conf"
	"delay-tasks/rest"
	"delay-tasks/utils"
)

func StartService() error {
	utils.StartLoop()

	api := mgin.NewMgin(mgin.WithLogger("delay-tasks"))

	api.POST("/handler/:cate", rest.RegisterCateHandler)
	api.POST("/task/:cate",    rest.AddTask)
	api.DELETE("/task/:cate",  rest.DelTask)
	api.GET("/task/:cate/:key",rest.FetchTask)
	api.Get("/tasks",          rest.ListTask)

	// health check
	api.GET("/health", func(c *mgin.Context) {
		c.String(http.StatusOK, "OK\n")
	})

	serviceConf := conf.ServiceConf
	listenParam := fmt.Sprintf("%s:%d", serviceConf.ListenHost, serviceConf.ListenPort)
	fmt.Printf("I am listening at %s...\n", listenParam)
	return http.ListenAndServe(listenParam, api)
}

