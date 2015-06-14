package main
import (
	"testing"
	"github.com/gin-gonic/gin"
)

func TestGinRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router:=gin.Default()
	BuildRoutes(router)
}