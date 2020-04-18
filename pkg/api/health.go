package api

import (
	"github.com/dovbysh/go-skeleton/pkg/schema"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func (a *Api) handlerHealth(c *gin.Context) {
	c.JSON(http.StatusOK, &schema.HealthResponse{
		R:    "healthy",
		Time: time.Now(),
	})
}
