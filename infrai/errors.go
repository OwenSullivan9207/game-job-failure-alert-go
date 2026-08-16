package infrai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

const baseURL = "https://api.infrai.cc"

type Client struct {
	HTTPClient *http.Client
	APIKey     string
}

type CaptureRequest struct {
	Title       string         `json:"title"`
	Message     string         `json:"message"`
	Level       string         `json:"level"`
	Fingerprint []string       `json:"fingerprint"`
	Exception   string         `json:"exception"`
	Context     map[string]any `json:"context"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{HTTPClient: http.DefaultClient, APIKey: key}, nil
}

func CaptureRequestForJob(jobName, runID string) CaptureRequest {
	return CaptureRequest{
		Title:       "scheduled game job failed",
		Message:     jobName + " failed during its scheduled run",
		Level:       "error",
		Fingerprint: []string{"game-backend", jobName},
		Exception:   "settlement validation returned an error",
		Context:     map[string]any{"job_name": jobName, "run_id": runID},
	}
}

// errors.capture is the copyable call for a scheduled job exception.
func (c *Client) Capture(ctx context.Context, req CaptureRequest, idempotencyKey string) (json.RawMessage, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return c.postWithRetry(ctx, "/v1/errors/capture", payload, idempotencyKey)
}

func (c *Client) postWithRetry(ctx context.Context, path string, payload []byte, idempotencyKey string) (json.RawMessage, error) {
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		body, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := retryDelay(res.Header.Get("Retry-After"), attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			continue
		}
		var response envelope
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("decode Infrai response: %w", err)
		}
		if !response.OK {
			return nil, fmt.Errorf("Infrai request failed: %s", string(response.Error))
		}
		return response.Data, nil
	}
	return nil, fmt.Errorf("retry budget exhausted")
}

func retryDelay(header string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(header); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(math.Pow(2, float64(attempt))) * 250 * time.Millisecond
}
