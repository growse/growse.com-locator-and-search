package main

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func TestGinRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	BuildRoutes(router, "/")
}
