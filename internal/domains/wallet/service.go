package wallet

import (
	"log"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/tatun2000/golang-lib/pkg/wrap"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

const (
	defaultMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
)

type Service struct {
	masterPrivateKey *bip32.Key
	publicKey        *bip32.Key
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
	publicKey := masterKey.PublicKey()

	return &Service{
		masterPrivateKey: masterKey,
		publicKey:        publicKey,
	}
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
		mnemonic = defaultMnemonic
	}

	seed = bip39.NewSeed(mnemonic, secretPhrase)
	return seed, nil
}

// m/44'/1'/0'/0/0
// 44 - BIP44
// 1 - testnet (mainnet = 0)
// 0 - account
// 0 - external addresses
// 0 - first address in leaf
func (s *Service) GenerateBitcoinAddressForTestNet() (result string, err error) {
	key44, err := s.masterPrivateKey.NewChildKey(bip32.FirstHardenedChild + 44)
	if err != nil {
		return result, wrap.Wrap(err)
	}

	key1, err := key44.NewChildKey(bip32.FirstHardenedChild + 1) // testnet
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

	pubKey := addrKey.PublicKey().Key
	address, err := btcutil.NewAddressPubKey(pubKey, &chaincfg.TestNet3Params)
	if err != nil {
		return result, wrap.Wrap(err)
	}

	return address.EncodeAddress(), nil
}
