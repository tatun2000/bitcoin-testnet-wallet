package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tatun2000/bitcoin-testnet-wallet/infrastructure"
	"github.com/tatun2000/golang-lib/pkg/wrap"
)

var walletCommand = &cobra.Command{
	Use:   "wallet",
	Short: "testnet wallet commands.",
	Long:  "testnet wallet commands.",
}

var walletAddressCommand = &cobra.Command{
	Use:                   "address",
	Short:                 "retrieve wallet address.",
	Long:                  "retrieve wallet address.",
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		walletService := infrastructure.App.InjectWalletService()

		address, err := walletService.GetWalletAddress()
		if err != nil {
			return wrap.Wrap(err)
		}

		fmt.Fprintf(os.Stdout, "Wallet address: %s\n", address)

		return nil
	},
}

var walletBalanceCommand = &cobra.Command{
	Use:                   "balance",
	Short:                 "retrieve wallet balance.",
	Long:                  "retrieve wallet balance.",
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		walletService := infrastructure.App.InjectWalletService()

		balance, err := walletService.GetWalletBalance()
		if err != nil {
			return wrap.Wrap(err)
		}

		fmt.Fprintf(os.Stdout, "Wallet balance: %d satoshi\n", balance)

		return nil
	},
}

func init() {
	rootCommand.AddCommand(walletCommand)
	walletCommand.AddCommand(walletAddressCommand)
	walletCommand.AddCommand(walletBalanceCommand)
}

func walletResetFlags() {
	walletCommand.Flags().Set("help", "")        //nolint:errcheck // err can be always
	walletAddressCommand.Flags().Set("help", "") //nolint:errcheck // err can be always
	walletBalanceCommand.Flags().Set("help", "") //nolint:errcheck // err can be always
}
