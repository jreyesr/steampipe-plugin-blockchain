package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hashicorp/go-hclog"
)

const API_BASE_URL = "https://chain.api.btc.com/v3"

type BlockchainClient struct {
	logger hclog.Logger
}

type ApiWrapper[T any] struct {
	Data      T      `json:"data"`
	Status    string `json:"status"` // Can be "success" or "fail"
	Message   string `json:"msg"`
	ErrorCode int    `json:"err_code"` // 0 if request was successful
}
type PaginatedWrapper[T any] struct {
	Items    []T `json:"list"`
	Page     int `json:"page"`
	PageSize int `json:"pagesize"`
	NumPages int `json:"page_total"`
	NumItems int `json:"total_count"`
}

type WalletInfo struct {
	Hash160            string `json:"-"` // This will be manually extracted from .Address
	Address            string `json:"address"`
	NumberTransactions int    `json:"tx_count"`
	NumberUnredeemed   int    `json:"unspent_tx_count"`
	TotalReceived      int    `json:"received"`
	TotalSent          int    `json:"sent"`
	FinalBalance       int    `json:"balance"`
}

// Fills the .Hash160 field from the .Address field, by reversing the Base58Check encoding used
// See the process in "Creating a Base58Check string" here: https://en.bitcoin.it/wiki/Base58Check_encoding
func (wi *WalletInfo) FillHash160() {
	result, _, err := base58.CheckDecode(wi.Address)
	if err != nil {
		panic(fmt.Errorf("invalid address %s, can't extract hash160", wi.Address))
	}
	wi.Hash160 = hex.EncodeToString(result)
}

func (i WalletInfo) String() string {
	return fmt.Sprintf(
		"WalletInfo{address=%s/%s,txns=%d total/%d unred,funds=%d in/%d out,balance=%d}",
		i.Address, i.Hash160, i.NumberTransactions, i.NumberUnredeemed, i.TotalReceived, i.TotalSent, i.FinalBalance,
	)
}

// Returns information about a single wallet, identified by its Base58 address
func (c BlockchainClient) GetWalletInfo(address string) (WalletInfo, error) {
	url := fmt.Sprintf("%s/address/%s", API_BASE_URL, address)
	c.logger.Debug("getWallet", "url", url)

	res, err := http.Get(url)
	if err != nil {
		return WalletInfo{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return WalletInfo{}, fmt.Errorf("GetWalletInfo - %s", res.Status)
	}
	body, _ := io.ReadAll(res.Body)
	c.logger.Debug("getWallet", "statusCode", res.StatusCode, "response", string(body))
	if string(body) == "Don't abuse the API. Please contact support@btcm.group\n" {
		c.logger.Debug("getWalletRateLimit", "url", url)
		return WalletInfo{}, RateLimitError{url}
	}

	var data ApiWrapper[WalletInfo]
	if err := json.Unmarshal(body, &data); err != nil {
		return WalletInfo{}, err
	}
	if data.Status != "success" {
		return WalletInfo{}, fmt.Errorf("%s", data.Message)
	}
	data.Data.FillHash160()

	return data.Data, nil
}

type TransactionInfo struct {
	Hash         string    `json:"hash"`
	Fee          int       `json:"fee"`
	RelayedBy    string    `json:"-"` // This does not work :(
	OriginalTime int       `json:"block_time"`
	Time         time.Time `json:"-"` // This will be filled manually from .OriginalTime
	InputsCount  int       `json:"inputs_count"`
	InputsValue  int       `json:"inputs_value"`
	OutputsCount int       `json:"outputs_count"`
	OutputsValue int       `json:"outputs_value"`
	Balance      int       `json:"balance_diff"`

	Inputs  []map[string]any `json:"inputs"`
	Outputs []map[string]any `json:"outputs"`
}

// Fills the .Time field by interpreting the .OriginalTime field as the number of seconds since the Unix epoch
func (ti *TransactionInfo) FillTime() {
	ti.Time = time.Unix(int64(ti.OriginalTime), 0)
}

func (i TransactionInfo) String() string {
	return fmt.Sprintf(
		"TransactionInfo{hash=%s,value=%d fee/%d balance,time=%s}",
		i.Hash, i.Fee, i.Balance, i.Time,
	)
}

// Returns a page of results for transactions that involve a certain wallet, identified by its Base58 address
func (c BlockchainClient) GetTransactionsForWallet(address string, page int) ([]TransactionInfo, error) {
	url := fmt.Sprintf("%s/address/%s/tx?page=%d", API_BASE_URL, address, page)
	c.logger.Debug("GetTransactionsForWallet", "url", url)
	res, err := http.Get(url)
	if err != nil {
		return make([]TransactionInfo, 0), err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return make([]TransactionInfo, 0), fmt.Errorf("GetTransactionsForWallet - %s", res.Status)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) == "Don't abuse the API. Please contact support@btcm.group\n" {
		c.logger.Debug("getWalletRateLimit", "url", url)
		return make([]TransactionInfo, 0), RateLimitError{url}
	}

	var data ApiWrapper[PaginatedWrapper[TransactionInfo]]
	if err := json.Unmarshal(body, &data); err != nil {
		return make([]TransactionInfo, 0), err
	}
	if data.Status != "success" {
		return make([]TransactionInfo, 0), fmt.Errorf("%s", data.Message)
	}

	// Manually parse the transaction time
	// NOTE Don't iterate over the value! range creates a copy of it, and we need the original
	for i := range data.Data.Items {
		data.Data.Items[i].FillTime()
	}

	return data.Data.Items, nil
}

// Returns data about a single transaction, identified by its hash
func (c BlockchainClient) GetTransaction(hash string) (TransactionInfo, error) {
	url := fmt.Sprintf("%s/tx/%s", API_BASE_URL, hash)

	res, err := http.Get(url)
	if err != nil {
		return TransactionInfo{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return TransactionInfo{}, fmt.Errorf("GetTransaction - %s", res.Status)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) == "Don't abuse the API. Please contact support@btcm.group\n" {
		c.logger.Debug("getWalletRateLimit", "url", url)
		return TransactionInfo{}, RateLimitError{url}
	}

	var data ApiWrapper[TransactionInfo]
	if err := json.Unmarshal(body, &data); err != nil {
		return TransactionInfo{}, err
	}
	if data.Status != "success" {
		return TransactionInfo{}, fmt.Errorf("%s", data.Message)
	}

	data.Data.FillTime()

	return data.Data, nil
}
