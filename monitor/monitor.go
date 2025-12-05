package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Config struct {
	// MCP Server settings
	Description  string
	MCPServerURL string
	APIKey       string

	// Health check settings
	Interval   time.Duration
	Timeout    time.Duration
	MaxRetries int

	// Process management
	// StartCommand string
	// WorkingDir   string
}

type Monitor struct {
	client *http.Client
	cfg    *Config
}

func NewMonitor(cfg *Config) *Monitor {
	return &Monitor{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		cfg: cfg,
	}
}

type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string `json:"protocolVersion"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
	Capabilities map[string]any `json:"capabilities"`
}

func (m *Monitor) makeInitializeRequest(ctx context.Context) error {
	initParams := InitializeParams{
		ProtocolVersion: "2025-06-18",
		Capabilities: map[string]any{
			"tools": map[string]bool{
				"list": true,
				"call": true,
			},
		},
	}
	initParams.ClientInfo.Name = "my-go-client"
	initParams.ClientInfo.Version = "0.1.0"

	initReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  initParams,
	}
	reqBody, err := json.Marshal(initReq)
	if err != nil {
		return fmt.Errorf("marshaling MCP request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.cfg.MCPServerURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("creating MCP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("MCP request failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[%s] Health check successful (%s)", m.cfg.Description, resp.Status)
	return nil
}

func (m *Monitor) StartMonitorLoop(ctx context.Context, failureCb func(string)) {
	log.Printf("[%s] Starting MCP Health Monitor...", m.cfg.Description)
	log.Printf("[%s] Monitoring MCP server at URL %s every %v", m.cfg.Description, m.cfg.MCPServerURL, m.cfg.Interval)

	ticker := time.NewTicker(m.cfg.Interval)
	defer ticker.Stop()

	consecutiveFailures := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.performHealthCheck(ctx, &consecutiveFailures); err != nil {
				failureCb(fmt.Sprintf("[%s] Health check cycle failed: %v", m.cfg.Description, err))
				return
			}
		}
	}
}

func (m *Monitor) performHealthCheck(ctx context.Context, consecutiveFailures *int) error {
	checkCtx, cancel := context.WithTimeout(ctx, m.cfg.Timeout)
	defer cancel()

	err := m.makeInitializeRequest(checkCtx)
	if err == nil {
		if *consecutiveFailures > 0 {
			log.Printf("[%s] Health check successful after %d failures", m.cfg.Description, *consecutiveFailures)
			*consecutiveFailures = 0
		}
		return nil
	}

	*consecutiveFailures++
	log.Printf("[%s] Health check failed (attempt %d/%d): %v", m.cfg.Description, *consecutiveFailures, m.cfg.MaxRetries, err)

	if *consecutiveFailures >= m.cfg.MaxRetries {
		return fmt.Errorf("max retries (%d) reached for %s", m.cfg.MaxRetries, m.cfg.Description)
	}

	return nil
}
