package config

import "time"

type Config struct {
	App   App
	DB    Database
	Redis Redis
}

type App struct {
	Name string
	Port int
	Env  string
}

type Database struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type Redis struct {
	Addr     string
	Password string
	DB       int
}
