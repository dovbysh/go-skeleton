package main

import (
	"flag"
	"github.com/dovbysh/go-skeleton/pkg/app"
	cfg "github.com/rowdyroad/go-yaml-config"
)

func main() {
	var (
		config     app.Config
		configFile string
		swaggerDir string
	)
	flag.StringVar(&configFile, "c", "api.yaml", "Config file")
	flag.StringVar(&swaggerDir, "swagger", "", "swagger")

	flag.Parse()
	cfg.LoadConfigFromFile(&config, configFile, &app.Config{})
	config.SwaggerDir = swaggerDir

	app := app.NewApp(config)
	defer app.Close()
	app.Run()
}
