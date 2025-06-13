package entities

type Tx struct {
	TxID     string   `json:"txid"`
	Version  int      `json:"version"`
	LockTime int      `json:"locktime"`
	Vin      []Vin    `json:"vin"`
	Vout     []Vout   `json:"vout"`
	Size     int      `json:"size"`
	Fee      int64    `json:"fee"`
	Status   TxStatus `json:"status"`
}

type Vin struct {
	TxID       string   `json:"txid"`
	Vout       uint32   `json:"vout"`
	Prevout    *Vout    `json:"prevout"`
	Witness    []string `json:"witness"`
	IsCoinbase bool     `json:"is_coinbase"`
	Sequence   uint32   `json:"sequence"`
}

type Vout struct {
	ScriptPubKey        string `json:"scriptpubkey"`
	ScriptPubKeyASM     string `json:"scriptpubkey_asm,omitempty"` // не всегда возвращается
	ScriptPubKeyType    string `json:"scriptpubkey_type"`
	ScriptPubKeyAddress string `json:"scriptpubkey_address"`
	Value               int64  `json:"value"`
}

type TxStatus struct {
	Confirmed   bool   `json:"confirmed"`
	BlockHeight int    `json:"block_height,omitempty"`
	BlockHash   string `json:"block_hash,omitempty"`
	BlockTime   int64  `json:"block_time,omitempty"`
}
