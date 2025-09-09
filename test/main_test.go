package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dovbysh/go-skeleton/pkg/app"
	"github.com/dovbysh/go-skeleton/pkg/schema"
	"github.com/dovbysh/tests_common"
	"github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/assert"
)

var (
	appAddr  string
	config   app.Config
	pgConfig *pg.Options
	wd       string
)

// HTTPClient provides a simple HTTP client replacement for gorequest
type HTTPClient struct {
	client  *http.Client
	headers map[string]string
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client:  &http.Client{Timeout: 30 * time.Second},
		headers: make(map[string]string),
	}
}

// AddHeader adds a header to the client
func (c *HTTPClient) AddHeader(key, value string) *HTTPClient {
	c.headers[key] = value
	return c
}

// Get performs a GET request
func (c *HTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	return c.client.Do(req)
}

// Post performs a POST request with JSON body
func (c *HTTPClient) Post(url string, body interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	
	return c.client.Do(req)
}

// GetJSON performs a GET request and parses JSON response
func (c *HTTPClient) GetJSON(url string, result interface{}) (*http.Response, error) {
	resp, err := c.Get(url)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()
	
	if result != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}
		err = json.Unmarshal(body, result)
		if err != nil {
			return resp, err
		}
	}
	
	return resp, nil
}

// PostJSON performs a POST request with JSON body and parses JSON response
func (c *HTTPClient) PostJSON(url string, reqBody interface{}, result interface{}) (*http.Response, error) {
	resp, err := c.Post(url, reqBody)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()
	
	if result != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}
		err = json.Unmarshal(body, result)
		if err != nil {
			return resp, err
		}
	}
	
	return resp, nil
}

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
		r, err := NewHTTPClient().Get("http://" + appAddr + "/api/health")
		if r == nil || err != nil {
			time.Sleep(time.Microsecond)
			continue
		}
		r.Body.Close()
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
	r, err := NewHTTPClient().Get("http://" + appAddr + "/swagger/")
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.NoError(t, err)
	if r != nil {
		r.Body.Close()
	}
}

func tHealth(t *testing.T) {
	var response schema.HealthResponse
	now := time.Now()
	r, err := NewHTTPClient().GetJSON(fmt.Sprintf("http://%s/api/health", appAddr), &response)
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.NoError(t, err)
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
	r, err := NewHTTPClient().PostJSON(fmt.Sprintf("http://%s/api/user/register", appAddr), &req, &resp2)
	assert.Equal(t, http.StatusBadRequest, r.StatusCode)
	assert.NoError(t, err)
	assert.Empty(t, resp2)
}

func register(t *testing.T, email string) schema.RegisterResponse {
	req := schema.RegisterRequest{
		Email:         email,
		PasswordPlain: "plainPassword",
		Name:          "RegiserName",
	}
	var response schema.RegisterResponse
	r, err := NewHTTPClient().PostJSON(fmt.Sprintf("http://%s/api/user/register", appAddr), &req, &response)
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.NoError(t, err)
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
	r, err := NewHTTPClient().PostJSON(fmt.Sprintf("http://%s/api/user/login", appAddr), &req2, &response2)
	assert.Equal(t, http.StatusNotFound, r.StatusCode)
	assert.NoError(t, err) // HTTP 404 is not a client error
	assert.Empty(t, response2)
}

func auth(t *testing.T, email string) schema.LoginResponse {
	register(t, email)
	req := schema.LoginRequest{
		Email:         email,
		PasswordPlain: "plainPassword",
	}
	var response schema.LoginResponse
	r, err := NewHTTPClient().PostJSON(fmt.Sprintf("http://%s/api/user/login", appAddr), &req, &response)
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.NoError(t, err)
	return response
}

func tHelloUser(t *testing.T) {
	// api should return hello and User for authorized user
	auth := auth(t, "TestHelloUser")
	assert.NotEmpty(t, auth)

	var response schema.HelloResponse
	client := NewHTTPClient().AddHeader("Authorization", fmt.Sprintf("Bearer %s", auth.Bearer))
	r, err := client.GetJSON(fmt.Sprintf("http://%s/api/user/hello", appAddr), &response)
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.NoError(t, err)
	assert.NotEmpty(t, response)
	assert.Equal(t, "TestHelloUser", response.User.Email)

	// unauthorized request to /api/user/hello should return unauthorized
	r, err = NewHTTPClient().Get(fmt.Sprintf("http://%s/api/user/hello", appAddr))
	assert.Equal(t, http.StatusUnauthorized, r.StatusCode)
	assert.NoError(t, err)
	if r != nil {
		r.Body.Close()
	}
}
