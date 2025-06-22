package utils

import "github.com/tatun2000/bitcoin-testnet-wallet/internal/constants"

// 1 input = P2WPKH ~ 68 vbytes
// 1 output = P2WPKH ~ 31 vbytes
// header and additional bytes ~ 10 vbytes
func CalculateFee(inputs int, outputs int) int64 {
	txSize := inputs*68 + outputs*31 + 10
	feeRate := constants.DefaultFeeRate
	return int64(txSize * feeRate)
}
