package entities

type TxOutput struct {
	TxID   string   `json:"txid"`
	Vout   int      `json:"vout"`
	Status TxStatus `json:"status"`
	Value  int64    `json:"value"`
}

type TxOutputs []TxOutput
