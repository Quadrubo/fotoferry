package config

import (
	"fmt"
	"io/fs"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type Config struct {
	StateDB      string
	LogFormat    string
	DryRun       bool
	RequirePaths []string
	FileMode     fs.FileMode
	DirMode      fs.FileMode
	OwnerUID     int
	OwnerGID     int
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
	v.SetDefault("FILE_MODE", "0644")
	v.SetDefault("DIR_MODE", "0755")
	v.SetDefault("OWNER_UID", -1)
	v.SetDefault("OWNER_GID", -1)
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
	fileMode, err := parseMode(v.GetString("FILE_MODE"))
	if err != nil {
		return nil, fmt.Errorf("config: FILE_MODE: %w", err)
	}
	dirMode, err := parseMode(v.GetString("DIR_MODE"))
	if err != nil {
		return nil, fmt.Errorf("config: DIR_MODE: %w", err)
	}

	cfg := &Config{
		StateDB:      v.GetString("STATE_DB"),
		LogFormat:    v.GetString("LOG_FORMAT"),
		DryRun:       v.GetBool("DRY_RUN"),
		RequirePaths: splitList(v.GetString("REQUIRE_PATHS")),
		FileMode:     fileMode,
		DirMode:      dirMode,
		OwnerUID:     v.GetInt("OWNER_UID"),
		OwnerGID:     v.GetInt("OWNER_GID"),
		Mappings:     mappings,
	}

	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

func parseMode(s string) (fs.FileMode, error) {
	n, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid octal mode %q", s)
	}
	return fs.FileMode(n), nil
}
