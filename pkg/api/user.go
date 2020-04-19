package api

import (
	"fmt"
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
	m.SetPassword(req.PasswordPlain)
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

func (a *Api) handlerUserLogin(c *gin.Context) {
	var req schema.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	var user models.User
	if err := a.db.Model(&user).Where("email = ?", req.Email).Limit(1).Select(); err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	reqUser := user
	reqUser.SetPassword(req.PasswordPlain)
	bearer:=user.GetAuthKey()
	if reqUser.GetAuthKey() != bearer {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("bad password"))
		return
	}
	c.JSON(http.StatusOK, schema.LoginResponse{
		Bearer: bearer,
	})
}
