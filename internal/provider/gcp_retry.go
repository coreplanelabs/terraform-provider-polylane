package provider

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
)

func activateGCPWithRetry(ctx context.Context, apiClient *client.Client, input client.ActivateGCPConnectionInput, wait func(context.Context, time.Duration) error) (*client.GCPConnection, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	for attempt := 0; ; attempt++ {
		connection, err := apiClient.ActivateGCPConnection(ctx, input)
		if err == nil || attempt >= 5 || !retryableGCPActivation(err) {
			return connection, err
		}
		if err := wait(ctx, time.Duration(1<<min(attempt+1, 4))*time.Second); err != nil {
			return nil, err
		}
	}
}

func retryableGCPActivation(err error) bool {
	var apiError *client.APIError
	if !errors.As(err, &apiError) {
		return false
	}
	switch apiError.StatusCode {
	case 429, 500, 502, 503, 504:
		return true
	case 400:
		// Durable Object RPC preserves the setup detail but flattens retry metadata.
		switch strings.TrimPrefix(apiError.Message, "[400] ") {
		case "Google Cloud has not accepted the connection identity. Confirm setup completed, wait a minute, then try again.",
			"Google Cloud has not granted access to the reader identity. Confirm setup completed, wait a minute, then try again.":
			return true
		}
	}
	return false
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
