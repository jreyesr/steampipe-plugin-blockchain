package blockchain

import (
	"context"
	"reflect"
	"runtime"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
)

var requestCounter syncint64.Counter
var rowsCounter syncint64.Counter
var timeTaken syncint64.Histogram

func InitMetrics() {
	meter := global.Meter("steampipe_plugin_blockchain")
	requestCounter, _ = meter.SyncInt64().Counter(
		"steampipe_plugin_blockchain.requests.count",
		instrument.WithDescription("A counter of requests made to the Blockchain API"),
		instrument.WithUnit(unit.Dimensionless))
	rowsCounter, _ = meter.SyncInt64().Counter(
		"steampipe_plugin_blockchain.rows.count",
		instrument.WithDescription("A counter of rows returned across all API calls"),
		instrument.WithUnit(unit.Dimensionless))
	timeTaken, _ = meter.SyncInt64().Histogram(
		"steampipe_plugin_blockchain.requests.time",
		instrument.WithDescription("A histogram of the times taken for each hydrate function"),
		instrument.WithUnit(unit.Milliseconds))
}

// Higher-order function that wraps a normal hydrate function with some more functionality:
//
// 1. It increments the requestCounter metric by 1 as the function is called
//
// 2. It times the execution of the hydrate function
//
// 3. Once the hydrate function returns, it records the time taken in the timeTaken histogram
//
// This wrapper function can be used in place of an ordinary hydration function, e.g.:
//
//	List: &plugin.ListConfig{
//		  KeyColumns: plugin.SingleColumn("search"),
//		  Hydrate:    wrapWithTimer(originalHydrateFunction),
//	},
func wrapWithTimer(f plugin.HydrateFunc) plugin.HydrateFunc {
	funcName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	var wrapped plugin.HydrateFunc = func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
		requestCounter.Add(
			ctx, 1,
			attribute.String("op", funcName))

		start := time.Now()
		ret, err := f(ctx, d, h)
		end := time.Now()

		timeTaken.Record(
			ctx, end.Sub(start).Milliseconds(),
			attribute.String("op", funcName),
			attribute.Bool("success", err == nil))
		return ret, err
	}
	return wrapped
}
