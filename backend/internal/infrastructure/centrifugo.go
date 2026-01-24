package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/centrifugal/gocent/v3"
	"github.com/rs/zerolog/log"
)

type CentrifugoClient interface {
	Publish(ctx context.Context, channel string, data interface{}) error
}

type centrifugoClient struct {
	client *gocent.Client
}

func NewCentrifugoClient(url, apiKey string) CentrifugoClient {
	// Ensure URL ends with /api if not present, though gocent might handle it.
	// gocent expects the full API endpoint usually.
	// But let's check standard gocent usage.
	// Actually gocent.New config takes Addr.
	
	// Check if user provided WS url or HTTP url. We need HTTP API url.
	// If user provided ws://...:8000, we likely need http://...:8000/api
	
	apiURL := url
	if strings.HasPrefix(url, "ws://") {
		apiURL = strings.Replace(url, "ws://", "http://", 1)
	} else if strings.HasPrefix(url, "wss://") {
		apiURL = strings.Replace(url, "wss://", "https://", 1)
	}
	
	// Remove /connection/websocket if present
	apiURL = strings.Replace(apiURL, "/connection/websocket", "", 1)
	
	// Append /api if not present
	if !strings.HasSuffix(apiURL, "/api") {
		// It might be just host:port
		if !strings.Contains(apiURL, "/api") {
			apiURL = fmt.Sprintf("%s/api", apiURL)
		}
	}

	return &centrifugoClient{
		client: gocent.New(gocent.Config{
			Addr: apiURL,
			Key:  apiKey,
		}),
	}
}

func (c *centrifugoClient) Publish(ctx context.Context, channel string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Publish to channel
	_, err = c.client.Publish(ctx, channel, jsonData)
	if err != nil {
		log.Error().Err(err).Str("channel", channel).Msg("Failed to publish to centrifugo")
		return err
	}
	
	log.Debug().Str("channel", channel).Msg("Published to centrifugo")
	return nil
}
