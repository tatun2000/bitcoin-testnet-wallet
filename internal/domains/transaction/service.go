package transaction

import (
	"bytes"
	"encoding/hex"
	"fmt"

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
)

type (
	IAddressService interface {
		GetChildBIP32Key() (result *bip32.Key, err error)
	}

	Service struct {
		addressService IAddressService
		client         *resty.Client
	}
)

func NewService(addressService IAddressService) *Service {
	return &Service{
		addressService: addressService,
		client: resty.New().
			SetBaseURL("https://blockstream.info/testnet/api/"),
	}
}

func (s *Service) CreateNewTransaction(prevTXID, walletAddress string) (err error) {
	// get transaction
	var respTx entities.Tx
	resp, err := s.client.R().
		SetResult(&respTx).
		Get(fmt.Sprintf("tx/%s", prevTXID))
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

	rawKey, err := s.addressService.GetChildBIP32Key()
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

	resp, err = s.client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody(hexTx).
		Post("api/tx")
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

func (s *Service) CreateNewTransactionWithChange(prevTXID, walletAddress string, sendToFaucet int) (err error) {
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

	rawKey, err := s.addressService.GetChildBIP32Key()
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
	fmt.Println("sender address:", senderAddr.EncodeAddress())

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
	txSize := 68 + (31 * 2) + 10
	feeRate := 2 // sat/vbyte
	fee := txSize * feeRate

	changeAmount := utxoValue - int64(sendToFaucet) - int64(fee)
	if changeAmount < 0 {
		return wrap.Wrap(fmt.Errorf("change amount %d is less than 0", changeAmount))
	}

	// add first output
	// P2WPKH
	pkScript, err := txscript.PayToAddrScript(faucetAddr)
	if err != nil {
		return wrap.Wrap(err)
	}
	tx.AddTxOut(wire.NewTxOut(int64(sendToFaucet), pkScript))
	// add second output
	// P2WPKH
	changePkScript, err := txscript.PayToAddrScript(senderAddr)
	if err != nil {
		return wrap.Wrap(err)
	}
	tx.AddTxOut(wire.NewTxOut(int64(changeAmount), changePkScript))

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
