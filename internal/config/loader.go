package config

import (
	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig() // optional .env file

	return &Config{
		App: App{
			Name: viper.GetString("APP_NAME"),
			Port: viper.GetInt("APP_PORT"),
			Env:  viper.GetString("APP_ENV"),
		},
		DB: Database{
			DSN:             viper.GetString("DB_DSN"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
		},
		Redis: Redis{
			Addr:     viper.GetString("REDIS_ADDR"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
	}, nil
}
