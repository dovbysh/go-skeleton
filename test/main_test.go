package test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/dovbysh/go-skeleton/pkg/app"
	"github.com/dovbysh/go-skeleton/pkg/schema"
	"github.com/dovbysh/tests_common"
	"github.com/go-pg/pg/v9"
	"github.com/parnurzeal/gorequest"
	"github.com/stretchr/testify/assert"
)

var (
	appAddr  string
	config   app.Config
	pgConfig *pg.Options
	wd       string
)

func TestM(t *testing.T) {
	wd, _ = os.Getwd()
	var wg sync.WaitGroup
	var pgCloser func()
	o, pgCloser, pgPort, pgContainer := tests_common.PostgreSQLContainer(&wg)
	pgConfig = &pg.Options{
		Addr:         o.Addr,
		User:         o.User,
		Password:     o.Password,
		Database:     o.Database,
		PoolSize:     o.PoolSize,
		MinIdleConns: o.MinIdleConns,
	}
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
		r, _, _ := gorequest.New().Get("http://" + appAddr + "/api/health").End()
		if r == nil {
			time.Sleep(time.Microsecond)
			continue
		}
		break

	}
	t.Run("tSwagger", tSwagger)
	t.Run("tHealth", tHealth)
	t.Run("tRegisterUser", tRegisterUser)
	t.Run("tLogin", tLogin)
	t.Run("tHelloUser", tHelloUser)
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

func tSwagger(t *testing.T) {
	r, _, errs := gorequest.New().Get("http://" + appAddr + "/swagger/").End()
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.Empty(t, errs)
}

func tHealth(t *testing.T) {
	var response schema.HealthResponse
	now := time.Now()
	r, _, errs := gorequest.New().Get(fmt.Sprintf("http://%s/api/health", appAddr)).EndStruct(&response)
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.Empty(t, errs)
	assert.NotEmpty(t, response)
	assert.True(t, response.Time.UnixNano() >= now.UnixNano())
}

func tRegisterUser(t *testing.T) {
	response := register(t, "RegisterUserEmail")
	assert.True(t, response.User.ID > 0)

	// register user with same Email
	req := schema.RegisterRequest{
		Email:         "RegisterUserEmail",
		PasswordPlain: "plainPassword",
		Name:          "RegiserName",
	}
	var resp2 schema.RegisterResponse
	r, _, errs := gorequest.New().Post(fmt.Sprintf("http://%s/api/user/register", appAddr)).SendStruct(&req).EndStruct(&resp2)
	assert.Equal(t, http.StatusBadRequest, r.StatusCode)
	assert.NotEmpty(t, errs)
	assert.Empty(t, resp2)
}

func register(t *testing.T, email string) schema.RegisterResponse {
	req := schema.RegisterRequest{
		Email:         email,
		PasswordPlain: "plainPassword",
		Name:          "RegiserName",
	}
	var response schema.RegisterResponse
	r, _, errs := gorequest.New().Post(fmt.Sprintf("http://%s/api/user/register", appAddr)).SendStruct(&req).EndStruct(&response)
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.Empty(t, errs)
	assert.NotEmpty(t, response)
	assert.Equal(t, req.Email, response.User.Email)
	assert.Equal(t, req.Name, response.User.Name)
	return response
}

func tLogin(t *testing.T) {
	response := auth(t, "TestAuth")
	assert.NotEmpty(t, response)

	req2 := schema.LoginRequest{
		Email:         "TestLoginUnregistered",
		PasswordPlain: "plainPassword",
	}
	var response2 schema.LoginResponse
	r, _, errs := gorequest.New().Post(fmt.Sprintf("http://%s/api/user/login", appAddr)).SendStruct(&req2).EndStruct(&response2)
	assert.Equal(t, http.StatusNotFound, r.StatusCode)
	assert.NotEmpty(t, errs)
	assert.Empty(t, response2)
}

func auth(t *testing.T, email string) schema.LoginResponse {
	register(t, email)
	req := schema.LoginRequest{
		Email:         email,
		PasswordPlain: "plainPassword",
	}
	var response schema.LoginResponse
	r, _, errs := gorequest.New().Post(fmt.Sprintf("http://%s/api/user/login", appAddr)).SendStruct(&req).EndStruct(&response)
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.Empty(t, errs)
	return response
}

func addAuth(req *gorequest.SuperAgent, auth schema.LoginResponse) *gorequest.SuperAgent {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auth.Bearer))
	return req
}

func tHelloUser(t *testing.T) {
	// api should return hello and User for authorized user
	auth := auth(t, "TestHelloUser")
	assert.NotEmpty(t, auth)

	var response schema.HelloResponse
	r, _, errs := addAuth(gorequest.New().Get(fmt.Sprintf("http://%s/api/user/hello", appAddr)), auth).EndStruct(&response)
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.Empty(t, errs)
	assert.NotEmpty(t, response)
	assert.Equal(t, "TestHelloUser", response.User.Email)

	// unauthorized request to /api/user/hello should return unauthorized
	r, _, _ = gorequest.New().Get(fmt.Sprintf("http://%s/api/user/hello", appAddr)).End()
	assert.Equal(t, http.StatusUnauthorized, r.StatusCode)
}
