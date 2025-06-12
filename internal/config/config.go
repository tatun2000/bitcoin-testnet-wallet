package config

import (
	"context"

	"github.com/spf13/viper"
	"github.com/tatun2000/golang-lib/pkg/validator"
	"github.com/tatun2000/golang-lib/pkg/wrap"
)

type Config struct {
	SecretPassphrase string `mapstructure:"secretPassphrase" validate:"min=10,max=100"`
	UniqueSeed       bool   `mapstructure:"uniqueSeed"`
}

func NewConfig(ctx context.Context, cfgPath string) (cfg *Config, err error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(cfgPath)
	if err = v.ReadInConfig(); err != nil {
		return nil, wrap.Wrap(err)
	}

	if err = v.Unmarshal(&cfg); err != nil {
		return cfg, wrap.Wrap(err)
	}

	if err = validator.Validator.StructCtx(ctx, cfg); err != nil {
		return cfg, wrap.Wrap(err)
	}

	return cfg, nil
}
