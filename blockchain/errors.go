package blockchain

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

type RateLimitError struct {
	url string
}

func (e RateLimitError) Error() string { return fmt.Sprintf("Rate Limit %s", e.url) }

func ShouldRetryBlockchainError(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData, err error) bool {
	if err, ok := err.(RateLimitError); ok {
		plugin.Logger(ctx).Warn("ShouldRetryBlockchainError", "err", err)
		return true
	}
	return false
}
