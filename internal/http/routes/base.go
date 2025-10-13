package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterBaseRoutes(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data":"hello world"})
	})
}
