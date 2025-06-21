package wallet

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/constants"
	"github.com/tatun2000/bitcoin-testnet-wallet/internal/entities"
	"github.com/tatun2000/golang-lib/pkg/wrap"
)

type (
	IAddressService interface {
		GenerateBitcoinBIP84AddressForTestNet() (result string, err error)
	}

	Service struct {
		addressService IAddressService
		client         *resty.Client
	}
)

func NewService(addressService IAddressService) *Service {
	s := &Service{
		addressService: addressService,
		client: resty.New().
			SetBaseURL("https://blockstream.info/testnet/api/"),
	}

	if err := s.PutWalletAddress(); err != nil {
		log.Fatal(err)
	}

	return s
}

func (s *Service) PutWalletAddress() (err error) {
	if _, err := os.Stat(constants.WalletAddressPath); err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(constants.WalletAddressPath)
			if err != nil {
				return wrap.Wrap(err)
			}
			defer file.Close()

			address, err := s.addressService.GenerateBitcoinBIP84AddressForTestNet()
			if err != nil {
				return wrap.Wrap(err)
			}

			if _, err = file.WriteString(address); err != nil {
				return wrap.Wrap(err)
			}
		} else {
			return wrap.Wrap(err)
		}
	}

	return nil
}

func (s *Service) GetWalletAddress() (result string, err error) {
	file, err := os.Open(constants.WalletAddressPath)
	if err != nil {
		return result, wrap.Wrap(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		result = scanner.Text()
	} else if err := scanner.Err(); err != nil {
		return result, wrap.Wrap(err)
	}

	return result, nil
}

func (s *Service) GetWalletBalance() (result int64, err error) {
	address, err := s.GetWalletAddress()
	if err != nil {
		return result, wrap.Wrap(err)
	}

	var respUTXOs entities.TxOutputs
	resp, err := s.client.R().
		SetResult(&respUTXOs).
		Get(fmt.Sprintf("address/%s/utxo", address))
	if err != nil {
		return result, wrap.Wrap(err)
	}

	if resp.Error() != nil {
		return result, wrap.Wrap(fmt.Errorf("%d %s: api error", resp.StatusCode(), resp.Status()))
	}

	return lo.SumBy(respUTXOs, func(vout entities.TxOutput) int64 { return vout.Value }), nil
}
