package main

import (
	"common/std"
	"net/http"

	"github.com/gin-gonic/gin"
	"{{snake .ServiceName}}/api"
)

func NewRouter(service api.Service) *gin.Engine {
	std.SetupGinValidator()
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/chatter/v1")
	v1.POST("/hello", func(c *gin.Context) {
		var req api.HelloRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := service.SayHello(ctx, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	return router
}
