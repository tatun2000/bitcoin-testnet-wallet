package wallet

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/entities"
	"github.com/tatun2000/golang-lib/pkg/wrap"
)

type (
	IAddressService interface {
		RetrieveAddress() (result string, err error)
	}

	ITransactionService interface {
		CreateNewTransaction(recepientAddress string, amount int64, txIDs ...string) (txID string, err error)
	}

	Service struct {
		addressService     IAddressService
		transactionService ITransactionService
		client             *resty.Client
	}
)

func NewService(addressService IAddressService, transactionService ITransactionService) *Service {
	s := &Service{
		addressService:     addressService,
		transactionService: transactionService,
		client: resty.New().
			SetBaseURL("https://blockstream.info/testnet/api/"),
	}

	return s
}

func (s *Service) GetWalletAddress() (result string, err error) {
	result, err = s.addressService.RetrieveAddress()
	if err != nil {
		return result, wrap.Wrap(err)
	}

	return result, nil
}

func (s *Service) GetWalletBalance() (confirmed, unconfirmed int64, err error) {
	address, err := s.GetWalletAddress()
	if err != nil {
		return confirmed, unconfirmed, wrap.Wrap(err)
	}

	txs, err := s.getConfirmedUTXOTransactions(address)
	if err != nil {
		return confirmed, unconfirmed, wrap.Wrap(err)
	}

	confirmed = lo.SumBy(txs, func(vout entities.TxOutput) int64 { return vout.Value })

	txs, err = s.getUnconfirmedUTXOTransactions(address)
	if err != nil {
		return confirmed, unconfirmed, wrap.Wrap(err)
	}

	unconfirmed = lo.SumBy(txs, func(vout entities.TxOutput) int64 { return vout.Value })

	return confirmed, unconfirmed, nil
}

func (s *Service) SendTo(address string, amount int64) (txid string, err error) {
	balance, _, err := s.GetWalletBalance()
	if err != nil {
		return "", wrap.Wrap(err)
	}

	if balance < amount {
		return "", wrap.Wrap(fmt.Errorf("insufficient available balance"))
	}

	walletAddress, err := s.GetWalletAddress()
	if err != nil {
		return "", wrap.Wrap(err)
	}

	txs, err := s.getConfirmedUTXOTransactions(walletAddress)
	if err != nil {
		return "", wrap.Wrap(err)
	}

	var (
		necessarySum int64
		txIDs        []string
	)
	for _, tx := range txs {
		txIDs = append(txIDs, tx.TxID)
		necessarySum += tx.Value
		if necessarySum >= amount {
			break
		}
	}

	txid, err = s.transactionService.CreateNewTransaction(address, amount, txIDs...)
	if err != nil {
		return "", wrap.Wrap(err)
	}

	return txid, nil
}

func (s *Service) getConfirmedUTXOTransactions(address string) (confirmedUTXOs []entities.TxOutput, err error) {
	var respUTXOs entities.TxOutputs
	resp, err := s.client.R().
		SetResult(&respUTXOs).
		Get(fmt.Sprintf("address/%s/utxo", address))
	if err != nil {
		return confirmedUTXOs, wrap.Wrap(err)
	}

	if resp.Error() != nil {
		return confirmedUTXOs, wrap.Wrap(fmt.Errorf("%d %s: api error", resp.StatusCode(), resp.Status()))
	}

	return lo.Filter(respUTXOs, func(vout entities.TxOutput, _ int) bool { return vout.Status.Confirmed }), nil
}

func (s *Service) getUnconfirmedUTXOTransactions(address string) (confirmedUTXOs []entities.TxOutput, err error) {
	var respUTXOs entities.TxOutputs
	resp, err := s.client.R().
		SetResult(&respUTXOs).
		Get(fmt.Sprintf("address/%s/utxo", address))
	if err != nil {
		return confirmedUTXOs, wrap.Wrap(err)
	}

	if resp.Error() != nil {
		return confirmedUTXOs, wrap.Wrap(fmt.Errorf("%d %s: api error", resp.StatusCode(), resp.Status()))
	}

	return lo.Filter(respUTXOs, func(vout entities.TxOutput, _ int) bool { return !vout.Status.Confirmed }), nil
}
