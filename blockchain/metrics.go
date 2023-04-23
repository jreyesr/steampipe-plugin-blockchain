package blockchain

import (
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
)

var requestCounter syncint64.Counter
var rowsCounter syncint64.Counter

func InitMetrics() {
	meter := global.Meter("steampipe_plugin_blockchain")
	requestCounter, _ = meter.SyncInt64().Counter("steampipe_plugin_blockchain.requests.count", instrument.WithUnit(unit.Dimensionless))
	rowsCounter, _ = meter.SyncInt64().Counter("steampipe_plugin_blockchain.rows.count", instrument.WithUnit(unit.Dimensionless))
}
