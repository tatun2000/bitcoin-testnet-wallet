package main

import (
	"context"
	"log"

	"github.com/tatun2000/bitcoin-testnet-wallet/internal/config"
)

type (
	Kernel struct {
		cfg *config.Config
	}
)

func NewKernel(ctx context.Context) *Kernel {
	cfg, err := config.NewConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return &Kernel{
		cfg: cfg,
	}
}
