package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type Config struct {
	StateDB      string
	LogFormat    string
	DryRun       bool
	RequirePaths []string
	Mappings     []Mapping `validate:"required,dive"`
}

type Mapping struct {
	ID     string `env:"ID"     validate:"required"`
	Source string `env:"SOURCE" validate:"required"`
	Dest   string `env:"DEST"   validate:"required"`
}

func Load(envFile string) (*Config, error) {
	v := viper.New()
	v.SetDefault("STATE_DB", "/data/state.db")
	v.SetDefault("LOG_FORMAT", "text")
	v.AutomaticEnv()

	if envFile != "" {
		v.SetConfigFile(envFile)
		v.SetConfigType("env")
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("config: reading %s: %w", envFile, err)
		}
	}

	mappings, err := parseGroup[Mapping](v, "MAPPING", "ID")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		StateDB:      v.GetString("STATE_DB"),
		LogFormat:    v.GetString("LOG_FORMAT"),
		DryRun:       v.GetBool("DRY_RUN"),
		RequirePaths: splitList(v.GetString("REQUIRE_PATHS")),
		Mappings:     mappings,
	}

	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
