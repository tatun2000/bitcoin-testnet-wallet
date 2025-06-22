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
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/utils"
	"github.com/tatun2000/golang-lib/pkg/wrap"
	"github.com/tyler-smith/go-bip32"
)

type (
	IAddressService interface {
		GetChildBIP32Key() (result *bip32.Key, err error)
		RetrieveAddress() (result string, err error)
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

func (s *Service) CreateNewTransaction(recepientAddress string, amount int64, txIDs ...string) (txID string, err error) {
	prevTXs := make([]entities.Tx, 0, len(txIDs))

	walletAddress, err := s.addressService.RetrieveAddress()
	if err != nil {
		return "", wrap.Wrap(err)
	}

	for _, txID := range txIDs {
		var respTx entities.Tx
		resp, err := s.client.R().
			SetResult(&respTx).
			Get(fmt.Sprintf("tx/%s", txID))
		if err != nil {
			return "", wrap.Wrap(err)
		}

		if resp.Error() != nil {
			return "", wrap.Wrap(fmt.Errorf("%d %s: api error", resp.StatusCode(), resp.Status()))
		}

		_, _, ok := lo.FindIndexOf(respTx.Vout, func(vout entities.Vout) bool {
			return vout.ScriptPubKeyAddress == walletAddress
		})
		if !ok {
			return "", wrap.Wrap(fmt.Errorf("current address %s not found", walletAddress))
		}

		prevTXs = append(prevTXs, respTx)
	}

	wif, witness, err := s.generateWifAndWitnessAddress()
	if err != nil {
		return "", wrap.Wrap(err)
	}

	senderPkScript, err := txscript.PayToAddrScript(witness)
	if err != nil {
		return "", wrap.Wrap(err)
	}

	// create new transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// add inputs
	var totalInputValue int64
	for _, prevTX := range prevTXs {
		prevOut, index, ok := lo.FindIndexOf(prevTX.Vout, func(vout entities.Vout) bool {
			return vout.ScriptPubKeyAddress == walletAddress
		})
		if !ok {
			return "", wrap.Wrap(fmt.Errorf("current address %s not found", walletAddress))
		}
		utxoHash, err := chainhash.NewHashFromStr(prevTX.TxID)
		if err != nil {
			return "", wrap.Wrap(err)
		}
		outPoint := wire.NewOutPoint(utxoHash, uint32(index))
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)
		totalInputValue += prevOut.Value
	}

	fee := utils.CalculateFee(len(prevTXs), 2)
	changeAmount := totalInputValue - amount - fee
	if changeAmount < 0 {
		return "", wrap.Wrap(fmt.Errorf("change amount %d is less than 0", changeAmount))
	}

	recepient, err := btcutil.DecodeAddress(recepientAddress, &chaincfg.TestNet3Params)
	if err != nil {
		return "", wrap.Wrap(err)
	}
	pkScript, err := txscript.PayToAddrScript(recepient)
	if err != nil {
		return "", wrap.Wrap(err)
	}
	tx.AddTxOut(wire.NewTxOut(int64(amount), pkScript))
	switch changeAmount {
	case 0:
		// one output
	default:
		// two outputs

		// add second output
		// P2WPKH
		changePkScript, err := txscript.PayToAddrScript(witness)
		if err != nil {
			return "", wrap.Wrap(err)
		}
		tx.AddTxOut(wire.NewTxOut(int64(changeAmount), changePkScript))
	}
	// sign (P2WPKH)
	sigHashes := txscript.NewTxSigHashes(tx)

	for idx, prevTX := range prevTXs {
		prevOut, _, ok := lo.FindIndexOf(prevTX.Vout, func(vout entities.Vout) bool {
			return vout.ScriptPubKeyAddress == walletAddress
		})
		if !ok {
			return "", wrap.Wrap(fmt.Errorf("current address %s not found", walletAddress))
		}
		witnessScript, err := txscript.WitnessSignature(
			tx,
			sigHashes,
			idx,
			prevOut.Value,
			senderPkScript, // для каждого UTXO может быть разный
			txscript.SigHashAll,
			wif.PrivKey, // или свой для каждого
			true,
		)
		if err != nil {
			return "", wrap.Wrap(err)
		}
		tx.TxIn[idx].Witness = witnessScript
	}

	var buf bytes.Buffer
	tx.Serialize(&buf)

	hexTx := hex.EncodeToString(buf.Bytes())
	resp, err := s.client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody(hexTx).
		Post("tx")
	if err != nil {
		return "", wrap.Wrap(err)
	}

	if resp.Error() != nil {
		return "", wrap.Wrap(fmt.Errorf("%d %s: api error", resp.StatusCode(), resp.Status()))
	}

	if resp.StatusCode() != 200 {
		return "", wrap.Wrap(fmt.Errorf("transaction error: %s", string(resp.Body())))
	}

	return string(resp.Body()), nil
}

func (s *Service) generateWifAndWitnessAddress() (wif *btcutil.WIF, witness *btcutil.AddressWitnessPubKeyHash, err error) {
	rawKey, err := s.addressService.GetChildBIP32Key()
	if err != nil {
		return nil, nil, wrap.Wrap(err)
	}
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), rawKey.Key)
	wif, err = btcutil.NewWIF(privateKey, &chaincfg.TestNet3Params, true)
	if err != nil {
		return nil, nil, wrap.Wrap(err)
	}
	pubKeyHash := btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed())
	witness, err = btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.TestNet3Params)
	if err != nil {
		return nil, nil, wrap.Wrap(err)
	}

	return wif, witness, nil
}
