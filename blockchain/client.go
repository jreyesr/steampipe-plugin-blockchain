package blockchain

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const API_BASE_URL = "https://blockchain.info"

type BlockchainClient struct{}

type WalletBalance struct {
	FinalBalance       int `json:"final_balance"`
	NumberTransactions int `json:"n_tx"`
	TotalReceived      int `json:"total_received"`
}

func (b WalletBalance) String() string {
	return fmt.Sprintf(
		"WalletBalance{final_balance=%d,num_transactions=%d,total_received=%d}",
		b.FinalBalance, b.NumberTransactions, b.TotalReceived,
	)
}

func (c BlockchainClient) GetBalance(address string) (WalletBalance, error) {
	url := fmt.Sprintf("%s/balance?active=%s", API_BASE_URL, address)

	res, err := http.Get(url)
	if err != nil {
		return WalletBalance{}, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var originalData map[string]WalletBalance
	if err := json.Unmarshal(body, &originalData); err != nil {
		return WalletBalance{}, err
	}
	data, found := originalData[address]

	if found {
		return data, nil
	} else {
		return WalletBalance{}, nil
	}
}

type WalletInfo struct {
	Hash160            string `json:"hash160"`
	Address            string `json:"address"`
	NumberTransactions int    `json:"n_tx"`
	NumberUnredeemed   int    `json:"n_unredeemed"`
	TotalReceived      int    `json:"total_received"`
	TotalSent          int    `json:"total_sent"`
	FinalBalance       int    `json:"final_balance"`
}

func (i WalletInfo) String() string {
	return fmt.Sprintf(
		"WalletInfo{address=%s/%s,txns=%d total/%d unred,funds=%d in/%d out,balance=%d}",
		i.Address, i.Hash160, i.NumberTransactions, i.NumberUnredeemed, i.TotalReceived, i.TotalSent, i.FinalBalance,
	)
}

func (c BlockchainClient) GetWalletInfo(address string) (WalletInfo, error) {
	url := fmt.Sprintf("%s/rawaddr/%s", API_BASE_URL, address)

	res, err := http.Get(url)
	if err != nil {
		return WalletInfo{}, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var data WalletInfo
	if err := json.Unmarshal(body, &data); err != nil {
		return WalletInfo{}, err
	}

	return data, nil
}

type UnixTime struct {
	time.Time
}

func (u *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	u.Time = time.Unix(timestamp, 0)
	return nil
}

type TransactionInfo struct {
	Hash    string
	Fee     int
	Time    UnixTime
	Balance int
}

func (i TransactionInfo) String() string {
	return fmt.Sprintf(
		"TransactionInfo{hash=%s,value=%d fee/%d balance,time=%s}",
		i.Hash, i.Fee, i.Balance, i.Time,
	)
}

func (c BlockchainClient) GetTransactionsForWallet(address string, offset int) ([]TransactionInfo, error) {
	url := fmt.Sprintf("%s/rawaddr/%s?offset=%d", API_BASE_URL, address, offset)

	res, err := http.Get(url)
	if err != nil {
		return make([]TransactionInfo, 0), err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var data struct {
		Transactions []TransactionInfo `json:"txs"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return make([]TransactionInfo, 0), err
	}

	return data.Transactions, nil
}
