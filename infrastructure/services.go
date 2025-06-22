package infrastructure

import (
	"sync"

	"github.com/tatun2000/bitcoin-testnet-wallet/internal/domains/address"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/domains/transaction"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/domains/wallet"
)

var (
	walletService     *wallet.Service
	walletServiceOnce sync.Once
)

func (k *Kernel) InjectWalletService() *wallet.Service {
	walletServiceOnce.Do(func() {
		walletService = wallet.NewService(
			k.InjectAddressService(),
			k.InjectTransactionService(),
		)
	})

	return walletService
}

var (
	addressService     *address.Service
	addressServiceOnce sync.Once
)

func (k *Kernel) InjectAddressService() *address.Service {
	addressServiceOnce.Do(func() {
		addressService = address.NewService(
			k.cfg.SecretPassphrase,
			k.cfg.UniqueSeed,
		)
	})

	return addressService
}

var (
	transactionService     *transaction.Service
	transactionServiceOnce sync.Once
)

func (k *Kernel) InjectTransactionService() *transaction.Service {
	transactionServiceOnce.Do(func() {
		transactionService = transaction.NewService(
			k.InjectAddressService(),
		)
	})

	return transactionService
}
