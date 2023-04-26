package blockchain

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func tableBlockchainWallet() *plugin.Table {
	return &plugin.Table{
		Name:        "blockchain_wallet",
		Description: "Returns information about Bitcoin wallets (also known as addresses)",
		// There is no List config, since you will never ever list all Bitcoin wallets...
		Get: &plugin.GetConfig{
			KeyColumns:     plugin.SingleColumn("address"),
			Hydrate:        getWallet,
			MaxConcurrency: 1,
		},
		Columns: []*plugin.Column{
			{Name: "address", Type: proto.ColumnType_STRING, Transform: transform.FromField("Address"), Description: "Wallet address, in the Base58 format"},
			{Name: "hash160", Type: proto.ColumnType_STRING, Transform: transform.FromField("Hash160"), Description: "Wallet address, as a 160-character hash"},
			{Name: "number_transactions", Type: proto.ColumnType_INT, Transform: transform.FromField("NumberTransactions"), Description: "Number of transactions involving this account"},
			{Name: "number_unredeemed", Type: proto.ColumnType_INT, Transform: transform.FromField("NumberUnredeemed"), Description: "Number of unredeemed transactions involving this account"},
			{Name: "total_received", Type: proto.ColumnType_INT, Transform: transform.FromField("TotalReceived"), Description: "Total funds sent TO this wallet, in satoshis (1e-8 BTC)"},
			{Name: "total_sent", Type: proto.ColumnType_INT, Transform: transform.FromField("TotalSent"), Description: "Total funds sent FROM this wallet, in satoshis (1e-8 BTC)"},
			{Name: "final_balance", Type: proto.ColumnType_INT, Transform: transform.FromField("FinalBalance"), Description: "Final balance of the wallet, in satoshis (1e-8 BTC)"},
		},
	}
}

func getWallet(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	requestCounter.Add(ctx, 1, attribute.String("op", "getWallet"))
	plugin.Logger(ctx).Warn("getWallet")

	quals := d.EqualsQuals
	plugin.Logger(ctx).Warn("getWallet", "quals", quals)
	address := quals["address"].GetStringValue()
	plugin.Logger(ctx).Warn("getWallet", "address", address)

	client := BlockchainClient{logger: plugin.Logger(ctx)}

	getWalletInfo := func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
		data, err := client.GetWalletInfo(address)
		if err != nil {
			span := trace.SpanFromContext(ctx)
			span.SetStatus(codes.Error, err.Error())

			span.AddEvent(
				"error",
				trace.WithAttributes(
					attribute.String("wallet", address),
				),
			)
		}
		return data, err
	}
	walletInfo, err := plugin.RetryHydrate(ctx, d, h, getWalletInfo, &plugin.RetryConfig{
		ShouldRetryErrorFunc: ShouldRetryBlockchainError,
		MaxAttempts:          3,
		RetryInterval:        1000,
		BackoffAlgorithm:     "Constant",
	})
	plugin.Logger(ctx).Debug("getWallet", "res", walletInfo, "err", err)
	if err != nil {
		span := trace.SpanFromContext(ctx)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	rowsCounter.Add(ctx, 1, attribute.String("op", "getWallet"))
	return walletInfo, nil
}
