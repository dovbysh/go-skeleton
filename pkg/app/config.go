package app

type Config struct {
	Listen     string   `yaml:"listen"`
	Postgres   Postgres `yaml:"postgres"`
	Orm        Orm      `yaml:"orm"`
	SwaggerDir string   ``
	DevEnv     bool     `yaml:"dev_env"`
}

type Orm struct {
	Debug bool `yaml:"debug"`
}

type Postgres struct {
	Addr         string `yaml:"addr"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	MinIdleConns int    `yaml:"min_idle_conns"`
	PoolSize     int    `yaml:"pool_size"`
}
