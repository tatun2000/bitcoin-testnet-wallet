package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tatun2000/bitcoin-testnet-wallet/internal/config"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/domains/wallet"
)

const (
	cfgPath = "../config"
)

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	cfg, err := config.NewConfig(ctx, cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	service := wallet.NewService(cfg.SecretPassphrase, cfg.UniqueSeed)

	address, err := service.GenerateBitcoinAddressForTestNet()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bitcoin Testnet address:", address)
	<-ctx.Done()
	// graceful shutdown
}
