package main

import (
	"gopkg.in/gin-gonic/gin.v1"
	"testing"
)

func TestGinRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	BuildRoutes(router)
}
