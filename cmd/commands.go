package main

import (
	"strings"

	"github.com/tatun2000/golang-lib/pkg/wrap"
)

func ExecuteCommand(command string) (err error) {
	rootCommand.SetArgs(strings.Fields(command))
	if err := rootCommand.Execute(); err != nil {
		return wrap.Wrap(err)
	}

	resetHelpFlags()

	return nil
}

func resetHelpFlags() {
	rootCommand.Flags().Set("help", "") //nolint:errcheck // err can be always

	walletResetFlags()
}
