package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// New instantiates and returns a new api
func New() *gin.Engine {
	// TODO: gin.New() instead to avoid logger etc
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	return router
}
