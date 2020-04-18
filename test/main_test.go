package test

import (
	"fmt"
	"github.com/dovbysh/go-skeleton/pkg/app"
	"github.com/dovbysh/tests_common"
	"github.com/go-pg/pg/v9"
	"github.com/parnurzeal/gorequest"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

var (
	appAddr  string
	config   app.Config
	pgConfig *pg.Options
	wd       string
)

func TestMain(t *testing.M) {
	var tResult int
	func() {
		wd, _ = os.Getwd()
		var wg sync.WaitGroup
		var pgCloser func()
		o, pgCloser, pgPort, pgContainer := tests_common.PostgreSQLContainer(&wg)
		pgConfig = o
		defer pgCloser()

		wg.Wait()

		sqlMigrateWorkDir := filepath.Dir(wd)
		sqlMigrate := exec.Command("make", "docker/migrate/up")
		sqlMigrate.Env = append(os.Environ(),
			fmt.Sprintf("DB_HOST=%s", pgContainer.Inspection.NetworkSettings.Gateway),
			fmt.Sprintf("DB_PORT=%d", pgPort),
			fmt.Sprintf("DB_NAME=%s", pgConfig.Database),
			fmt.Sprintf("DB_USER=%s", pgConfig.User),
			fmt.Sprintf("DB_PASSWORD=%s", pgConfig.Password),
		)
		sqlMigrate.Dir = sqlMigrateWorkDir
		var out io.ReadCloser
		out, err := sqlMigrate.StdoutPipe()
		if err != nil {
			panic(err)
		}
		var outErr io.ReadCloser
		outErr, err = sqlMigrate.StderrPipe()
		if err != nil {
			panic(err)
		}
		go io.Copy(os.Stdout, out)
		go io.Copy(os.Stderr, outErr)
		err = sqlMigrate.Run()
		if err != nil {
			panic(err)
		}

		appAddr, _, _ = tests_common.GetFreeLocalAddr()
		config = getAppConfig(*pgConfig, appAddr)
		a := app.NewApp(config)
		defer a.Close()
		go a.Run()

		for {
			r, _, _ := gorequest.New().Post("http://" + appAddr + "/api/health").End()
			if r == nil {
				time.Sleep(time.Microsecond)
				continue
			}
			break

		}
		tResult = t.Run()
	}()
	os.Exit(tResult)
}

func getAppConfig(pgConfig pg.Options, appAddr string) app.Config {
	return app.Config{
		Listen:     appAddr,
		SwaggerDir: fmt.Sprintf("%s/api/openapi_spec", path.Dir(wd)),
		Postgres: app.Postgres{
			Addr:         pgConfig.Addr,
			User:         pgConfig.User,
			Password:     pgConfig.Password,
			Database:     pgConfig.Database,
			MinIdleConns: pgConfig.MinIdleConns,
			PoolSize:     pgConfig.PoolSize,
		},
		Orm: app.Orm{
			Debug: true,
		},
		DevEnv: true,
	}
}

func TestSwagger(t *testing.T) {
	r, _, errs := gorequest.New().Get("http://" + appAddr + "/swagger/").End()
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.Empty(t, errs)
}
