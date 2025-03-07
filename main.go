package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	port        int
	minLatency  time.Duration
	maxLatency  time.Duration
	configFile  string
	defaultPort = 8080
)

func init() {
	flag.IntVar(&port, "port", defaultPort, "Port to run the server on")
	flag.DurationVar(&minLatency, "min-latency", 0, "Minimum latency to simulate (e.g., 100ms)")
	flag.DurationVar(&maxLatency, "max-latency", 0, "Maximum latency to simulate (e.g., 500ms)")
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.Parse()
}

func main() {
	// Initialize configuration
	config, err := loadConfig(configFile)
	if err != nil {
		log.Printf("Warning: Failed to load config file: %v. Using default and command-line settings.", err)
		config = &Config{
			Port:       defaultPort,
			MinLatency: 0,
			MaxLatency: 0,
			Models: Models{
				Embedding: []string{"text-embedding-ada-002"},
				Chat:      []string{"gpt-3.5-turbo", "gpt-4"},
			},
		}
	}

	// Command-line flags override config file
	if port != defaultPort {
		config.Port = port
	}
	if minLatency != 0 {
		config.MinLatency = minLatency
	}
	if maxLatency != 0 {
		config.MaxLatency = maxLatency
	}

	// Set up the router
	router := setupRouter(config)

	// Start the server
	serverAddr := fmt.Sprintf(":%d", config.Port)
	log.Printf("Starting OpenAI mock server on %s", serverAddr)
	log.Printf("Latency settings: min=%v, max=%v", config.MinLatency, config.MaxLatency)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(config *Config) *gin.Engine {
	router := gin.Default()

	// Add middleware for simulating latency
	router.Use(simulateLatencyMiddleware(config.MinLatency, config.MaxLatency))

	// API routes
	apiGroup := router.Group("/v1")
	{
		// Chat completions endpoint
		apiGroup.POST("/chat/completions", func(c *gin.Context) {
			handleChatCompletions(c, config)
		})

		// Embeddings endpoint
		apiGroup.POST("/embeddings", func(c *gin.Context) {
			handleEmbeddings(c, config)
		})

		// Models endpoint
		apiGroup.GET("/models", func(c *gin.Context) {
			handleModels(c, config)
		})
	}

	return router
}

func simulateLatencyMiddleware(min, max time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if min > 0 {
			latency := min
			if max > min {
				// Calculate a random latency between min and max
				latency = min + time.Duration(float64(max-min)*rand.Float64())
			}
			time.Sleep(latency)
		}
		c.Next()
	}
}

// Config represents the server configuration
type Config struct {
	Port       int           `json:"port"`
	MinLatency time.Duration `json:"min_latency"`
	MaxLatency time.Duration `json:"max_latency"`
	Models     Models        `json:"models"`
}

type Models struct {
	Embedding []string `json:"embedding"`
	Chat      []string `json:"chat"`
}

// ConfigFile represents the JSON structure of the config file
type ConfigFile struct {
	Port       int    `json:"port"`
	MinLatency string `json:"min_latency"`
	MaxLatency string `json:"max_latency"`
	Models     Models `json:"models"`
}

func loadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return &Config{}, nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var configFile ConfigFile
	if err := json.NewDecoder(file).Decode(&configFile); err != nil {
		return nil, err
	}

	// Convert string durations to time.Duration
	minLatency, err := time.ParseDuration(configFile.MinLatency)
	if err != nil {
		return nil, fmt.Errorf("invalid min_latency: %v", err)
	}

	maxLatency, err := time.ParseDuration(configFile.MaxLatency)
	if err != nil {
		return nil, fmt.Errorf("invalid max_latency: %v", err)
	}

	config := &Config{
		Port:       configFile.Port,
		MinLatency: minLatency,
		MaxLatency: maxLatency,
		Models:     configFile.Models,
	}

	return config, nil
}

// ChatCompletionRequest represents the request structure for chat completions
type ChatCompletionRequest struct {
	Model    string                  `json:"model"`
	Messages []ChatCompletionMessage `json:"messages"`
}

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents the response structure for chat completions
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int                   `json:"index"`
		Message      ChatCompletionMessage `json:"message"`
		FinishReason string                `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// EmbeddingRequest represents the request structure for embeddings
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse represents the response structure for embeddings
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// ModelsResponse represents the response structure for the models endpoint
type ModelsResponse struct {
	Object string `json:"object"`
	Data   []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

func handleChatCompletions(c *gin.Context, config *Config) {
	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Log warning but continue processing with default values
		log.Printf("Warning: Malformed chat completion request: %v. Continuing with default values.", err)
	}

	// Check if the requested model is supported
	modelSupported := false
	for _, model := range config.Models.Chat {
		if model == req.Model {
			modelSupported = true
			break
		}
	}

	if !modelSupported && req.Model != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("The model '%s' does not exist", req.Model),
				"type":    "invalid_request_error",
				"param":   "model",
				"code":    "model_not_found",
			},
		})
		return
	}

	// If no model specified, use the first available
	if req.Model == "" && len(config.Models.Chat) > 0 {
		req.Model = config.Models.Chat[0]
	}

	// Ensure we have messages to process
	if req.Messages == nil {
		req.Messages = []ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		}
	}

	// Generate a mock response
	response := ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", randomID(29)),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []struct {
			Index        int                   `json:"index"`
			Message      ChatCompletionMessage `json:"message"`
			FinishReason string                `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: ChatCompletionMessage{
					Role:    "assistant",
					Content: "This is a mock response from the OpenAI API emulator. Your request has been processed successfully.",
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     calculateTokens(req.Messages),
			CompletionTokens: 20,
			TotalTokens:      calculateTokens(req.Messages) + 20,
		},
	}

	c.JSON(http.StatusOK, response)
}

func handleEmbeddings(c *gin.Context, config *Config) {
	var req EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Log warning but continue processing with default values
		log.Printf("Warning: Malformed embedding request: %v. Continuing with default values.", err)
	}

	// Check if the requested model is supported
	modelSupported := false
	for _, model := range config.Models.Embedding {
		if model == req.Model {
			modelSupported = true
			break
		}
	}

	if !modelSupported && req.Model != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("The model '%s' does not exist", req.Model),
				"type":    "invalid_request_error",
				"param":   "model",
				"code":    "model_not_found",
			},
		})
		return
	}

	// If no model specified, use the first available
	if req.Model == "" && len(config.Models.Embedding) > 0 {
		req.Model = config.Models.Embedding[0]
	}

	// Ensure we have input to process
	if req.Input == nil {
		req.Input = []string{"Default input text"}
	}

	// Generate mock embeddings
	response := EmbeddingResponse{
		Object: "list",
		Data: make([]struct {
			Object    string    `json:"object"`
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		}, len(req.Input)),
		Model: req.Model,
		Usage: struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		}{
			PromptTokens: 0,
			TotalTokens:  0,
		},
	}

	// Generate mock embeddings for each input
	for i, input := range req.Input {
		// Create a deterministic but random-looking embedding vector
		embedding := generateMockEmbedding(input, 1536) // OpenAI embeddings are typically 1536 dimensions

		response.Data[i] = struct {
			Object    string    `json:"object"`
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		}{
			Object:    "embedding",
			Embedding: embedding,
			Index:     i,
		}

		// Update token counts
		tokens := len(input) / 4 // Rough approximation
		response.Usage.PromptTokens += tokens
		response.Usage.TotalTokens += tokens
	}

	c.JSON(http.StatusOK, response)
}

func handleModels(c *gin.Context, config *Config) {
	response := ModelsResponse{
		Object: "list",
		Data: []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		}{},
	}

	// Add chat models
	for _, model := range config.Models.Chat {
		response.Data = append(response.Data, struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		}{
			ID:      model,
			Object:  "model",
			Created: time.Now().Unix() - 86400*30, // 30 days ago
			OwnedBy: "openai",
		})
	}

	// Add embedding models
	for _, model := range config.Models.Embedding {
		response.Data = append(response.Data, struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		}{
			ID:      model,
			Object:  "model",
			Created: time.Now().Unix() - 86400*30, // 30 days ago
			OwnedBy: "openai",
		})
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions

// randomID generates a random ID with the specified length
func randomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// calculateTokens estimates the number of tokens in the messages
func calculateTokens(messages []ChatCompletionMessage) int {
	tokens := 0
	for _, msg := range messages {
		tokens += len(msg.Content) / 4 // Rough approximation
	}
	return tokens
}

// generateMockEmbedding creates a deterministic but random-looking embedding vector
func generateMockEmbedding(input string, dimensions int) []float64 {
	// Use a hash of the input as a seed for deterministic randomness
	seed := int64(0)
	for _, c := range input {
		seed = seed*31 + int64(c)
	}
	r := rand.New(rand.NewSource(seed))

	// Generate the embedding vector
	embedding := make([]float64, dimensions)
	for i := range embedding {
		embedding[i] = r.Float64()*2 - 1 // Values between -1 and 1
	}

	// Normalize the vector to unit length
	sum := 0.0
	for _, v := range embedding {
		sum += v * v
	}
	magnitude := math.Sqrt(sum)
	for i := range embedding {
		embedding[i] /= magnitude
	}

	return embedding
}
