package api

import (
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

	a.router.Any("/health", a.handlerHealth)
}
