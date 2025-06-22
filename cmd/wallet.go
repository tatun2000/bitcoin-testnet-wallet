package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tatun2000/bitcoin-testnet-wallet/infrastructure"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/entities"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/utils"
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
	RunE: func(cmd *cobra.Command, _ []string) error {
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
	RunE: func(cmd *cobra.Command, _ []string) error {
		walletService := infrastructure.App.InjectWalletService()

		confirmedBalance, unconfirmedBalance, err := walletService.GetWalletBalance()
		if err != nil {
			return wrap.Wrap(err)
		}

		fmt.Fprintf(os.Stdout, "Wallet balance: \n\t\tAvailable: %d satoshi\n\t\tOn hold: %d satoshi\n", confirmedBalance, unconfirmedBalance)

		return nil
	},
}

var walletSendToCommand = &cobra.Command{
	Use:   "send",
	Short: "Send money to bitcoin address.",
	Long: utils.GenLongMessage("Send money to bitcoin address", map[string]entities.HelpArg{
		"address": {
			Description: "Destination address",
			SeqNumber:   1,
			Required:    true,
		},
		"amount": {
			Description: "Amount of satoshi",
			SeqNumber:   2,
			Required:    true,
		},
	}),
	Args:                  cobra.RangeArgs(2, 2),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		address := args[0]
		amount, err := strconv.Atoi(args[1])
		if err != nil {
			return wrap.Wrap(err)
		}

		walletService := infrastructure.App.InjectWalletService()

		txid, err := walletService.SendTo(address, int64(amount))
		if err != nil {
			return wrap.Wrap(err)
		}

		fmt.Fprintf(os.Stdout, "Successfully sent %d satoshi to: %s\n", amount, address)
		fmt.Fprintf(os.Stdout, "Transaction ID: %s\n", txid)

		return nil
	},
}

func init() {
	rootCommand.AddCommand(walletCommand)
	walletCommand.AddCommand(walletAddressCommand)
	walletCommand.AddCommand(walletBalanceCommand)
	walletCommand.AddCommand(walletSendToCommand)
}

func walletResetFlags() {
	walletCommand.Flags().Set("help", "")        //nolint:errcheck // err can be always
	walletAddressCommand.Flags().Set("help", "") //nolint:errcheck // err can be always
	walletBalanceCommand.Flags().Set("help", "") //nolint:errcheck // err can be always
	walletSendToCommand.Flags().Set("help", "")  //nolint:errcheck // err can be always
}
