package api

import (
	"github.com/dovbysh/go-skeleton/pkg/models"
	"github.com/dovbysh/go-skeleton/pkg/schema"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (a *Api) handlerUserRegister(c *gin.Context) {
	var req schema.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	c.AbortWithStatusJSON(http.StatusOK, schema.RegisterResponse{User: &models.User{
		ID:    1,
		Email: req.Email,
		Name:  req.Name,
	}})
}
