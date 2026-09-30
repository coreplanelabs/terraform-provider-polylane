package provider

import (
	"context"
	"errors"
	"time"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
)

func activateGCPWithRetry(ctx context.Context, apiClient *client.Client, input client.ActivateGCPConnectionInput, wait func(context.Context, time.Duration) error) (*client.GCPConnection, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	for attempt := 0; ; attempt++ {
		connection, err := apiClient.ActivateGCPConnection(ctx, input)
		var apiError *client.APIError
		if err == nil || attempt >= 5 || !errors.As(err, &apiError) || (apiError.StatusCode != 429 && apiError.StatusCode != 502 && apiError.StatusCode != 503 && apiError.StatusCode != 504) {
			return connection, err
		}
		if err := wait(ctx, time.Duration(1<<min(attempt+1, 4))*time.Second); err != nil {
			return nil, err
		}
	}
}

func waitForGCPRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
