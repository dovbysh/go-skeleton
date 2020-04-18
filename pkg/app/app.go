package app

import (
	"context"
	"github.com/dovbysh/go-skeleton/pkg/api"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v9"
)

type App struct {
	config Config
	server *http.Server
	db     *pg.DB
	api    *api.Api
}

func NewApp(cfg Config) *App {
	router := gin.Default()
	router.Use(cors.Default())
	if !cfg.DevEnv {
		gin.SetMode(gin.ReleaseMode)
	}

	server := &http.Server{
		Addr:    cfg.Listen,
		Handler: router,
	}

	opt := &pg.Options{
		User:         cfg.Postgres.User,
		Password:     cfg.Postgres.Password,
		Database:     cfg.Postgres.Database,
		Addr:         cfg.Postgres.Addr,
		MinIdleConns: cfg.Postgres.MinIdleConns,
		PoolSize:     cfg.Postgres.PoolSize,
	}

	app := &App{
		config: cfg,
		server: server,
		db:     pg.Connect(opt),
	}

	if cfg.Orm.Debug {
		app.db.AddQueryHook(app)
	}

	if cfg.SwaggerDir != "" {
		router.Static("/swagger", cfg.SwaggerDir)
	}
	app.api = api.NewApi(app.db)
	app.api.InitRouter(router.Group("/api"))
	return app
}

func (a *App) Run() {
	a.server.ListenAndServe()
}

func (a *App) Close() {
	a.db.Close()
	a.server.Shutdown(context.Background())
}

func (a *App) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	log.Println(q.FormattedQuery())
	return c, nil
}
func (a *App) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	return nil
}
