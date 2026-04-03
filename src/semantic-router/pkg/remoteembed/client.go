package remoteembed

import (
"bytes"
"encoding/json"
"fmt"
"io"
"net/http"
"sync"
"time"
)

// Config holds the remote embedding API configuration.
type Config struct {
URL            string `yaml:"url"`
APIKey         string `yaml:"api_key"`
Model          string `yaml:"model"`
TimeoutSeconds int    `yaml:"timeout_seconds,omitempty"`
}

// embeddingRequest is the OpenAI-compatible embedding request body.
type embeddingRequest struct {
Model          string `json:"model"`
Input          string `json:"input"`
EncodingFormat string `json:"encoding_format"`
}

// embeddingResponse is the OpenAI-compatible embedding response.
type embeddingResponse struct {
Data []struct {
Embedding []float32 `json:"embedding"`
} `json:"data"`
}

// Client is a singleton remote embedding client.
var (
globalClient *client
mu           sync.RWMutex
)

type client struct {
httpClient *http.Client
url        string
apiKey     string
model      string
}

// Init initializes the global remote embedding client.
func Init(cfg Config) error {
if cfg.URL == "" {
return fmt.Errorf("remote embedding URL is required")
}
if cfg.Model == "" {
return fmt.Errorf("remote embedding model name is required")
}

timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
if timeout <= 0 {
timeout = 30 * time.Second
}

mu.Lock()
defer mu.Unlock()
globalClient = &client{
httpClient: &http.Client{Timeout: timeout},
url:        cfg.URL,
apiKey:     cfg.APIKey,
model:      cfg.Model,
}
return nil
}

// IsInitialized returns true if the remote embedding client is ready.
func IsInitialized() bool {
mu.RLock()
defer mu.RUnlock()
return globalClient != nil
}

// GetEmbedding calls the remote OpenAI-compatible /v1/embeddings endpoint.
func GetEmbedding(text string) ([]float32, error) {
mu.RLock()
c := globalClient
mu.RUnlock()
if c == nil {
return nil, fmt.Errorf("remote embedding client not initialized; configure remote_embedding in embedding_models")
}

reqBody := embeddingRequest{
Model:          c.model,
Input:          text,
EncodingFormat: "float",
}
bodyBytes, err := json.Marshal(reqBody)
if err != nil {
return nil, fmt.Errorf("failed to marshal remote embedding request: %w", err)
}

req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(bodyBytes))
if err != nil {
return nil, fmt.Errorf("failed to create remote embedding request: %w", err)
}
req.Header.Set("Content-Type", "application/json")
if c.apiKey != "" {
req.Header.Set("Authorization", "Bearer "+c.apiKey)
}

resp, err := c.httpClient.Do(req)
if err != nil {
return nil, fmt.Errorf("remote embedding request failed: %w", err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
body, _ := io.ReadAll(resp.Body)
return nil, fmt.Errorf("remote embedding API returned status %d: %s", resp.StatusCode, string(body))
}

var embResp embeddingResponse
if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
return nil, fmt.Errorf("failed to decode remote embedding response: %w", err)
}

if len(embResp.Data) == 0 || len(embResp.Data[0].Embedding) == 0 {
return nil, fmt.Errorf("remote embedding API returned empty embedding")
}

return embResp.Data[0].Embedding, nil
}
