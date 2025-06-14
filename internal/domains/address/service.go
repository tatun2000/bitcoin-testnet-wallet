package address

import (
	"log"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/constants"
	"github.com/tatun2000/golang-lib/pkg/wrap"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

type Service struct {
	masterPrivateKey *bip32.Key
}

func generateSeed(secretPhrase string, uniqueSeed bool) (seed []byte, err error) {
	var mnemonic string
	if uniqueSeed {
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			return seed, wrap.Wrap(err)
		}

		mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return seed, wrap.Wrap(err)
		}
	} else {
		mnemonic = constants.DefaultMnemonic
	}

	seed = bip39.NewSeed(mnemonic, secretPhrase)
	return seed, nil
}

func NewService(secretPhrase string, uniqueSeed bool) *Service {
	seed, err := generateSeed(secretPhrase, uniqueSeed)
	if err != nil {
		log.Fatal(err)
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		log.Fatal(err)
	}

	return &Service{
		masterPrivateKey: masterKey,
	}
}

// Path: m/84'/1'/0'/0/0
// 84 - BIP84
// 1 - testnet (mainnet = 0)
// 0 - account
// 0 - external addresses
// 0 - first address in leaf
func (s *Service) GetChildBIP32Key() (result *bip32.Key, err error) {
	key84, err := s.masterPrivateKey.NewChildKey(bip32.FirstHardenedChild + 84)
	if err != nil {
		return result, wrap.Wrap(err)
	}

	key1, err := key84.NewChildKey(bip32.FirstHardenedChild + 1) // testnet
	if err != nil {
		return result, wrap.Wrap(err)
	}

	key0, err := key1.NewChildKey(bip32.FirstHardenedChild + 0)
	if err != nil {
		return result, wrap.Wrap(err)
	}

	ext, err := key0.NewChildKey(0)
	if err != nil {
		return result, wrap.Wrap(err)
	}

	addrKey, err := ext.NewChildKey(0)
	if err != nil {
		return result, wrap.Wrap(err)
	}

	return addrKey, nil
}

func (s *Service) GenerateBitcoinBIP84AddressForTestNet() (result string, err error) {
	key, err := s.GetChildBIP32Key()
	if err != nil {
		return result, wrap.Wrap(err)
	}

	pubKeyHash := btcutil.Hash160(key.PublicKey().Key)

	// Generate P2WPKH (bech32) address
	address, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.TestNet3Params)
	if err != nil {
		return result, wrap.Wrap(err)
	}

	return address.EncodeAddress(), nil
}
