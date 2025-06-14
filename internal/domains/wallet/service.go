package wallet

import (
	"bufio"
	"log"
	"os"

	"github.com/tatun2000/bitcoin-testnet-wallet/internal/constants"
	"github.com/tatun2000/golang-lib/pkg/wrap"
)

type (
	IAddressService interface {
		GenerateBitcoinBIP84AddressForTestNet() (result string, err error)
	}

	Service struct {
		addressService IAddressService
	}
)

func NewService(addressService IAddressService) *Service {
	s := &Service{
		addressService: addressService,
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
