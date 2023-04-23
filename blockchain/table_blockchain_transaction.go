package blockchain

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
	"go.opentelemetry.io/otel/attribute"
)

func tableBlockchainTransaction() *plugin.Table {
	return &plugin.Table{
		Name:        "blockchain_transaction",
		Description: "Returns information about Bitcoin transactions",
		// The List config requires a search key, since you will never list all Bitcoin transactions...
		List: &plugin.ListConfig{
			KeyColumns: plugin.SingleColumn("wallet"),
			Hydrate:    listTransactions,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("hash"),
			Hydrate:    getTransaction,
		},
		Columns: []*plugin.Column{
			{Name: "hash", Type: proto.ColumnType_STRING, Transform: transform.FromField("Hash"), Description: "Transaction hash, unique across all transactions in the blockchain"},
			{Name: "fee", Type: proto.ColumnType_INT, Transform: transform.FromField("Fee"), Description: "Fee paid by the sender to the miner, in satoshis (1e-8 BTC)"},
			{Name: "relayed_by", Type: proto.ColumnType_IPADDR, Transform: transform.FromField("RelayedBy"), Description: "The IP of the node that announced this transaction"},
			{Name: "time", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("Time"), Description: "Timestamp of the transaction"},
			{Name: "inputs_count", Type: proto.ColumnType_INT, Transform: transform.FromField("InputsCount"), Description: "Number of inputs to this transaction"},
			{Name: "inputs_value", Type: proto.ColumnType_INT, Transform: transform.FromField("InputsValue"), Description: "Total value of the inputs to this transaction, in satoshis (1e-8 BTC)"},
			{Name: "outputs_count", Type: proto.ColumnType_INT, Transform: transform.FromField("OutputsCount"), Description: "Number of outputs from this transaction"},
			{Name: "outputs_value", Type: proto.ColumnType_INT, Transform: transform.FromField("OutputsValue"), Description: "Total value of the outputs from this transaction, in satoshis (1e-8 BTC)"},

			// Raw JSON fields for a bunch of data that can't be exposed easily
			{Name: "inputs", Type: proto.ColumnType_JSON, Transform: transform.FromField("Inputs"), Description: "Data about the inputs to the transaction: wallets, value transferred rrom each"},
			{Name: "outputs", Type: proto.ColumnType_JSON, Transform: transform.FromField("Outputs"), Description: "Data about the outputs from the transaction: wallets, value sent to each, transaction that spent the funds"},

			// Search fields
			{Name: "wallet", Type: proto.ColumnType_STRING, Transform: transform.FromQual("wallet"), Description: "Search field to search transactions by wallet. Only set when searching transactions by wallet."},
			{Name: "wallet_balance", Type: proto.ColumnType_INT, Transform: transform.FromField("Balance"), Description: "If searching by a wallet, total amount involved in the transaction FROM THE POINT OF VIEW OF THE WALLET, in satoshis (1e-8 BTC). Only set when searching transactions by wallet."},
		},
	}
}

func listTransactions(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	requestCounter.Add(ctx, 1, attribute.String("op", "listTransactions"))

	plugin.Logger(ctx).Warn("listTransactions")
	quals := d.EqualsQuals
	plugin.Logger(ctx).Warn("listTransactions", "quals", quals)
	wallet := quals["wallet"].GetStringValue()
	plugin.Logger(ctx).Warn("listTransactions", "wallet", wallet)

	client := BlockchainClient{logger: plugin.Logger(ctx)}

	page := 1 // Pagination for this API starts at 1!
	getTransactionsPage := func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
		return client.GetTransactionsForWallet(wallet, page)
	}
	retryConfig := &plugin.RetryConfig{
		ShouldRetryErrorFunc: ShouldRetryBlockchainError,
		MaxAttempts:          3,
		RetryInterval:        1000,
		BackoffAlgorithm:     "Constant",
	}

	for { // Run over all pages until we get an empty one, that means we're done
		transactionsGeneric, err := plugin.RetryHydrate(ctx, d, h, getTransactionsPage, retryConfig)
		plugin.Logger(ctx).Debug("listTransactions", "res", transactionsGeneric, "err", err)

		if err != nil {
			plugin.Logger(ctx).Error(fmt.Sprintf(
				"Unable to fetch transactions for wallet %s at offset %d: %s",
				wallet, page, err),
			)
			return nil, err
		}
		transactions := transactionsGeneric.([]TransactionInfo)

		for _, tx := range transactions {
			d.StreamListItem(ctx, tx)
			rowsCounter.Add(ctx, 1, attribute.String("op", "listTransactions"))
		}

		if len(transactions) == 0 {
			plugin.Logger(ctx).Debug(fmt.Sprintf(
				"Exiting, got 0 results for wallet %s at offset %d",
				wallet, page),
			)
			break
		}

		page++
	}
	return nil, nil
}

func getTransaction(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Warn("getTransaction")

	quals := d.EqualsQuals
	plugin.Logger(ctx).Warn("getTransaction", "quals", quals)
	hash := quals["hash"].GetStringValue()
	plugin.Logger(ctx).Warn("getTransaction", "hash", hash)

	client := BlockchainClient{logger: plugin.Logger(ctx)}

	getTransaction := func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
		return client.GetTransaction(hash)
	}
	txInfo, err := plugin.RetryHydrate(ctx, d, h, getTransaction, &plugin.RetryConfig{
		ShouldRetryErrorFunc: ShouldRetryBlockchainError,
		MaxAttempts:          3,
		RetryInterval:        1000,
		BackoffAlgorithm:     "Constant",
	})
	plugin.Logger(ctx).Debug("getTransaction", "res", txInfo, "err", err)
	if err != nil {
		return nil, err
	}

	return txInfo, nil
}
