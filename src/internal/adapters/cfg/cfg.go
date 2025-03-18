package cfg

import (
	"net/url"
	"time"

	"github.com/pkg/errors"
)

func NewConfig() (*Config, error) {
	envParser := envReader{}

	cfg := &Config{
		Debug:       envParser.toBoolSupressErr("DEBUG"),
		CoolingDown: envParser.toTimeDuration("COOLING_DOWN_PERIOD"),
		SomeEnv:     envParser.toString("SOME_ENV"),
		SomeUrl:     envParser.toUrl("SOME_URL"),
		SomePeriod:  envParser.toTimeDuration("SOME_PERIOD"),
		PG: pgCfg{
			Host:          envParser.toString("PG_HOST"),
			Port:          envParser.toString("PG_PORT"),
			User:          envParser.toString("PG_USER"),
			Password:      envParser.toString("PG_PASSWORD"),
			DbName:        envParser.toString("PG_DB_NAME"),
			DbNameDefault: envParser.toString("PG_DB_NAME_DEFAULT"),
			SslMode:       envParser.toString("PG_SSL_MODE"),
			ConnPool: pgCfgConnPull{
				IdleConns:    envParser.toInt("PG_CONNS_IDLE"),
				MaxOpenConns: envParser.toInt("PG_CONNS_MAX"),
			},
		},
	}

	if err := envParser.parsingErr("\n"); err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}

	return cfg, nil
}

type Config struct {
	Debug       bool
	CoolingDown time.Duration
	SomeEnv     string
	SomeUrl     url.URL
	SomePeriod  time.Duration
	PG          pgCfg
}

type pgCfg struct {
	Host          string
	Port          string
	User          string
	Password      string
	DbName        string
	DbNameDefault string
	SslMode       string
	ConnPool      pgCfgConnPull
}

type pgCfgConnPull struct {
	IdleConns    int
	MaxOpenConns int
}
