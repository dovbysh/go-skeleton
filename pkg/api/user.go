package api

import (
	"github.com/dovbysh/go-skeleton/pkg/models"
	"github.com/dovbysh/go-skeleton/pkg/schema"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v9"
	"net/http"
)

func (a *Api) handlerUserRegister(c *gin.Context) {
	var req schema.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	m := models.User{
		Email: req.Email,
		Name:  req.Name,
	}
	if _, err := a.db.Model(&m).Insert(); err != nil {
		pgErr, ok := err.(pg.Error)
		if ok && pgErr.IntegrityViolation() {
			c.AbortWithError(http.StatusBadRequest, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}
	c.JSON(http.StatusOK, schema.RegisterResponse{
		User: &m,
	})
}
