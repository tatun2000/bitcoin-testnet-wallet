package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/tatun2000/bitcoin-testnet-wallet/infrastructure"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/constants"
	"github.com/tatun2000/golang-lib/pkg/wrap"
)

var rootCommand = &cobra.Command{
	Use:   "",
	Short: "testnet wallet shell.",
	Long:  "testnet wallet shell.",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func setRootUID() (err error) {
	if err := syscall.Setuid(0); err != nil {
		return wrap.Wrap(err)
	}

	return nil
}

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	infrastructure.App = infrastructure.NewKernel(ctx)

	instance, err := initReadline()
	if err != nil {
		log.Fatal(err)
	}
	defer instance.Close()

	go listenUserCommands(cancelFunc, instance)

	<-ctx.Done()
	// graceful shutdown
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("wallet",
		readline.PcItem("address"),
		readline.PcItem("balance"),
		readline.PcItem("send"),
	),
	readline.PcItem("help"),
	readline.PcItem("exit"),
)

// initReadline inits readline Instance.
func initReadline() (instance *readline.Instance, err error) {
	instance, err = readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("%s> ", constants.WalletShell),
		InterruptPrompt: "^C",
		EOFPrompt:       constants.EOFCommand,
		AutoComplete:    completer,
	})
	if err != nil {
		return nil, wrap.Wrap(err)
	}

	return instance, nil
}

func listenUserCommands(cancel context.CancelFunc, instance *readline.Instance) {
	for {
		line, err := instance.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				continue
			}
			cancel()
			return
		}
		line = strings.TrimSpace(line)
		if line == constants.EOFCommand {
			cancel()
			return
		}
		// when user tab Enter
		if lo.IsEmpty(line) {
			continue
		}
		if err = ExecuteCommand(line); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
	}
}
