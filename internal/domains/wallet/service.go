package wallet

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/entities"
	"github.com/tatun2000/golang-lib/pkg/wrap"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
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

// Путь: m/84'/1'/0'/0/0
// 84 - BIP84
// 1 - testnet (mainnet = 0)
// 0 - account
// 0 - external addresses
// 0 - first address in leaf
func (s *Service) getChildBIP32Key() (result *bip32.Key, err error) {
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
	key, err := s.getChildBIP32Key()
	if err != nil {
		return result, wrap.Wrap(err)
	}

	// SHA256 → RIPEMD160
	pubKey := key.PublicKey().Key
	shaHash := sha256.Sum256(pubKey)
	ripemd := ripemd160.New()
	ripemd.Write(shaHash[:])
	pubKeyHash := ripemd.Sum(nil)

	// Generate P2WPKH (bech32) address
	address, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.TestNet3Params)
	if err != nil {
		return result, wrap.Wrap(err)
	}

	return address.EncodeAddress(), nil
}

func (s *Service) CreateNewTransaction(prevTXID, walletAddress string) (err error) {
	client := resty.New()

	// get transaction
	var respTx entities.Tx
	resp, err := client.R().
		SetResult(&respTx).
		Get(fmt.Sprintf("https://blockstream.info/testnet/api/tx/%s", prevTXID))
	if err != nil {
		return wrap.Wrap(err)
	}

	if resp.Error() != nil {
		return wrap.Wrap(fmt.Errorf("%d %s: api error", resp.StatusCode(), resp.Status()))
	}

	// recepient address
	faucetAddrStr := "tb1qlj64u6fqutr0xue85kl55fx0gt4m4urun25p7q"
	faucetAddr, err := btcutil.DecodeAddress(faucetAddrStr, &chaincfg.TestNet3Params)
	if err != nil {
		return wrap.Wrap(err)
	}

	// UTXO
	utxoTxIDStr := respTx.TxID

	prevOut, index, ok := lo.FindIndexOf(respTx.Vout, func(vout entities.Vout) bool {
		return vout.ScriptPubKeyAddress == walletAddress
	})
	if !ok {
		return wrap.Wrap(fmt.Errorf("current address %s not found", walletAddress))
	}

	utxoVout := uint32(index)
	utxoValue := prevOut.Value

	rawKey, err := s.getChildBIP32Key()
	if err != nil {
		return wrap.Wrap(err)
	}
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), rawKey.Key)
	wif, err := btcutil.NewWIF(privateKey, &chaincfg.TestNet3Params, true)
	if err != nil {
		return wrap.Wrap(err)
	}
	pubKeyHash := btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed())
	senderAddr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.TestNet3Params)
	if err != nil {
		return wrap.Wrap(err)
	}
	// создаём pkScript — locking script для входа
	// именно этот pkScript используется в подписи
	senderPkScript, err := txscript.PayToAddrScript(senderAddr)
	if err != nil {
		return wrap.Wrap(err)
	}

	// create new transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// add input
	utxoHash, err := chainhash.NewHashFromStr(utxoTxIDStr)
	if err != nil {
		return wrap.Wrap(err)
	}
	outPoint := wire.NewOutPoint(utxoHash, utxoVout)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	tx.AddTxIn(txIn)

	// calculate fee and change

	// 1 input = P2WPKH ~ 68 vbytes
	// 1 output = P2WPKH ~ 31 vbytes
	// header and additional bytes ~ 10 vbytes
	txSize := 68 + 31 + 10
	feeRate := 2 // sat/vbyte
	fee := txSize * feeRate

	sendAmount := utxoValue - int64(fee)

	// add output
	// P2WPKH
	pkScript, err := txscript.PayToAddrScript(faucetAddr)
	if err != nil {
		return wrap.Wrap(err)
	}
	txOut := wire.NewTxOut(sendAmount, pkScript)
	tx.AddTxOut(txOut)

	// sign (P2WPKH)
	//fetcher := txscript.NewCannedPrevOutputFetcher(senderPkScript, utxoValue)
	sigHashes := txscript.NewTxSigHashes(tx)

	witnessScript, err := txscript.WitnessSignature(
		tx,
		sigHashes,
		0, // input index
		utxoValue,
		senderPkScript, // locking script from UTXO
		txscript.SigHashAll,
		wif.PrivKey,
		true,
	)
	if err != nil {
		return wrap.Wrap(err)
	}
	tx.TxIn[0].Witness = witnessScript

	var buf bytes.Buffer
	tx.Serialize(&buf)

	hexTx := hex.EncodeToString(buf.Bytes())
	fmt.Println("Raw TX (hex):", hexTx)

	resp, err = client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody(hexTx).
		Post("https://blockstream.info/testnet/api/tx")
	if err != nil {
		return wrap.Wrap(err)
	}

	if resp.Error() != nil {
		return wrap.Wrap(fmt.Errorf("%d %s: api error", resp.StatusCode(), resp.Status()))
	}

	fmt.Println(resp.StatusCode())
	fmt.Println("Raw response bytes:", resp.Body())

	fmt.Println("response string:", string(resp.Body()))

	return nil
}
