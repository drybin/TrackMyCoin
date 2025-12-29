package config

import (
	"errors"
	"time"

	"github.com/drybin/TrackMyCoin/pkg/env"
	"github.com/drybin/TrackMyCoin/pkg/wrap"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	ServiceName              string
	TgConfig                 TgConfig
	GoogleAPIKey             string
	GoogleServiceAccountFile string
	GoogleSheetID            string
	GoogleSheetRange         string
}

type TgConfig struct {
	BotToken string
	ChatId   string
	Timeout  time.Duration
}

func (c Config) Validate() error {
	var errs []error

	err := validation.ValidateStruct(&c,
		validation.Field(&c.ServiceName, validation.Required),
	)
	if err != nil {
		return wrap.Errorf("failed to validate cli config: %w", err)
	}

	return errors.Join(errs...)
}

func InitConfig() (*Config, error) {
	config := Config{
		ServiceName:              env.GetString("APP_NAME", "TrackMyCoin"),
		TgConfig:                 initTgConfig(),
		GoogleAPIKey:             env.GetString("GOOGLE_API_KEY", ""),
		GoogleServiceAccountFile: env.GetString("GOOGLE_SERVICE_ACCOUNT_FILE", "service-account-file.json"),
		GoogleSheetID:            env.GetString("GOOGLE_SHEET_ID", "1zDO5I9ZWnT9AbD--RT9NZX3aQgem6d1FEleq0ISsElk"),
		GoogleSheetRange:         env.GetString("GOOGLE_SHEET_RANGE", ""), // Пусто = читать первый лист полностью
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func initTgConfig() TgConfig {
	return TgConfig{
		BotToken: env.GetString("TG_BOT_TOKEN", ""),
		ChatId:   env.GetString("TG_CHAT_ID", ""),
		Timeout:  10 * time.Second,
	}
}
