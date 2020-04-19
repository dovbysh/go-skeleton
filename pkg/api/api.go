package api

import (
	"fmt"
	"github.com/dovbysh/go-skeleton/pkg/api/middleware"
	"github.com/dovbysh/go-skeleton/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v9"
)

type Api struct {
	db     *pg.DB
	router *gin.RouterGroup
}

func NewApi(db *pg.DB) *Api {
	return &Api{
		db: db,
	}
}

func (a *Api) InitRouter(router *gin.RouterGroup) {
	a.router = router

	a.router.GET("/health", a.handlerHealth)
	user := a.router.Group("/user")

	user.Use(
		middleware.UserAuth(
			[]middleware.SkipperFunc{
				middleware.AllowPathPrefixSkipper(
					"/api/user/register",
					"/api/user/login",
				),
			},
			a.searchUser,
		),
	)
	user.POST("/register", a.handlerUserRegister)
	user.POST("/login", a.handlerUserLogin)
	user.GET("/hello", a.handlerUserHello)

}

func (a *Api) searchUser(AuthKey string) (*models.User, error) {
	u, e := models.AuthKeyToUserSearch(AuthKey)
	if e != nil {
		return nil, e
	}
	user := models.User{ID: u.ID}
	if err := a.db.Model(&user).WherePK().Select(); err != nil {
		return nil, err
	}
	if user.AuthHash() != u.Password {
		return nil, fmt.Errorf("wrong password")
	}
	return &user, nil
}
