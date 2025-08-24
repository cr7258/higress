package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alibaba/higress/plugins/golang-filter/mcp-server/servers/higress"
	"github.com/alibaba/higress/plugins/golang-filter/mcp-session/common"
	"github.com/mark3labs/mcp-go/mcp"
)

type TokenFailoverConfig struct {
	Enabled             bool   `json:"enabled"`
	FailureThreshold    int    `json:"failureThreshold,omitempty"`
	SuccessThreshold    int    `json:"successThreshold,omitempty"`
	HealthCheckInterval int    `json:"healthCheckInterval,omitempty"`
	HealthCheckTimeout  int    `json:"healthCheckTimeout,omitempty"`
	HealthCheckModel    string `json:"healthCheckModel,omitempty"`
}

type AIProvider struct {
	Name                 string               `json:"name"`
	Type                 string               `json:"type"`
	Protocol             string               `json:"protocol,omitempty"`
	Tokens               []string             `json:"tokens,omitempty"`
	TokenFailoverConfig  *TokenFailoverConfig `json:"tokenFailoverConfig,omitempty"`
}

type AIProviderResponse = higress.APIResponse[AIProvider]

func RegisterAIProviderTools(mcpServer *common.MCPServer, client *higress.HigressClient) {
	mcpServer.AddTool(
		mcp.NewTool("list-ai-providers", mcp.WithDescription("List all available AI providers")),
		handleListAIProviders(client),
	)
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema("get-ai-provider", "Get detailed information about a specific AI provider", getAIProviderSchema()),
		handleGetAIProvider(client),
	)
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema("add-ai-provider", "Add a new AI provider", getAddAIProviderSchema()),
		handleAddAIProvider(client),
	)
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema("update-ai-provider", "Update an existing AI provider", getUpdateAIProviderSchema()),
		handleUpdateAIProvider(client),
	)
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema("delete-ai-provider", "Delete an existing AI provider", getAIProviderSchema()),
		handleDeleteAIProvider(client),
	)
}

func handleListAIProviders(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		respBody, err := client.Get("/v1/ai-providers")
		if err != nil {
			return nil, fmt.Errorf("failed to list AI providers: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(respBody),
				},
			},
		}, nil
	}
}

func handleGetAIProvider(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments
		name, ok := arguments["name"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'name' argument")
		}

		respBody, err := client.Get(fmt.Sprintf("/v1/ai-providers/%s", name))
		if err != nil {
			return nil, fmt.Errorf("failed to get AI provider '%s': %w", name, err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(respBody),
				},
			},
		}, nil
	}
}

func handleAddAIProvider(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments
		configurations, ok := arguments["configurations"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'configurations' argument")
		}

		// Validate required fields
		if _, ok := configurations["name"]; !ok {
			return nil, fmt.Errorf("missing required field 'name' in configurations")
		}
		if _, ok := configurations["type"]; !ok {
			return nil, fmt.Errorf("missing required field 'type' in configurations")
		}

		respBody, err := client.Post("/v1/ai-providers", configurations)
		if err != nil {
			return nil, fmt.Errorf("failed to add AI provider: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(respBody),
				},
			},
		}, nil
	}
}

func handleUpdateAIProvider(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments
		name, ok := arguments["name"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'name' argument")
		}

		configurations, ok := arguments["configurations"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'configurations' argument")
		}

		// Get current AI provider configuration to merge with updates
		currentBody, err := client.Get(fmt.Sprintf("/v1/ai-providers/%s", name))
		if err != nil {
			return nil, fmt.Errorf("failed to get current AI provider configuration: %w", err)
		}

		var response AIProviderResponse
		if err := json.Unmarshal(currentBody, &response); err != nil {
			return nil, fmt.Errorf("failed to parse current AI provider response: %w", err)
		}

		currentConfig := response.Data

		// Update configurations using JSON marshal/unmarshal for type conversion
		configBytes, err := json.Marshal(configurations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal configurations: %w", err)
		}

		var newConfig AIProvider
		if err := json.Unmarshal(configBytes, &newConfig); err != nil {
			return nil, fmt.Errorf("failed to parse AI provider configurations: %w", err)
		}

		// Merge configurations (overwrite with new values where provided)
		// Note: name and type cannot be updated
		if newConfig.Protocol != "" {
			currentConfig.Protocol = newConfig.Protocol
		}
		if newConfig.Tokens != nil {
			currentConfig.Tokens = newConfig.Tokens
		}
		if newConfig.TokenFailoverConfig != nil {
			currentConfig.TokenFailoverConfig = newConfig.TokenFailoverConfig
		}

		respBody, err := client.Put(fmt.Sprintf("/v1/ai-providers/%s", name), currentConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to update AI provider '%s': %w", name, err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(respBody),
				},
			},
		}, nil
	}
}

func handleDeleteAIProvider(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments
		name, ok := arguments["name"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'name' argument")
		}

		respBody, err := client.Delete(fmt.Sprintf("/v1/ai-providers/%s", name))
		if err != nil {
			return nil, fmt.Errorf("failed to delete AI provider '%s': %w", name, err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(respBody),
				},
			},
		}, nil
	}
}

func getAIProviderSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"description": "The name of the AI provider to retrieve"
			}
		},
		"required": ["name"],
		"additionalProperties": false
	}`)
}

// TODO: some providers have special configurations (e.g., AWS Bedrock, Google Vertex AI), we need to support them in the future
func getAddAIProviderSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"configurations": {
				"type": "object",
				"properties": {
					"name": {
						"type": "string",
						"description": "The name of the AI provider"
					},
					"type": {
						"type": "string",
						"enum": [
							"moonshot", "azure", "ai360", "github", "qwen", "openai", 
							"groq", "grok", "baichuan", "yi", "deepseek", "zhipuai", 
							"ollama", "claude", "baidu", "hunyuan", "stepfun", "minimax", 
							"cloudflare", "spark", "gemini", "deepl", "mistral", "cohere", 
							"doubao", "coze", "together-ai", "dify", "bedrock", "vertex"
						],
						"description": "The type of AI provider"
					},
					"protocol": {
						"type": "string",
						"enum": ["openai/v1"],
						"description": "The protocol used by the AI provider (currently only openai/v1 is supported)"
					},
					"tokens": {
						"type": "array",
						"items": {"type": "string"},
						"description": "API tokens for authentication"
					},
					"tokenFailoverConfig": {
						"type": "object",
						"properties": {
							"enabled": {
								"type": "boolean",
								"description": "Whether token failover is enabled"
							},
							"failureThreshold": {
								"type": "integer",
								"minimum": 1,
								"description": "Number of failures before marking token as unhealthy"
							},
							"successThreshold": {
								"type": "integer",
								"minimum": 1,
								"description": "Number of successes before marking token as healthy"
							},
							"healthCheckInterval": {
								"type": "integer",
								"description": "Health check interval in milliseconds"
							},
							"healthCheckTimeout": {
								"type": "integer",
								"description": "Health check timeout in milliseconds"
							},
							"healthCheckModel": {
								"type": "string",
								"description": "Model to use for health checks"
							}
					},
					"additionalProperties": false
				}
			},
			"required": ["name", "type"],
				"additionalProperties": false
			}
		},
		"required": ["configurations"],
		"additionalProperties": false
	}`)
}

func getUpdateAIProviderSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"description": "The name of the AI provider to update"
			},
			"configurations": {
				"type": "object",
				"properties": {
					"protocol": {
						"type": "string",
						"enum": ["openai/v1"],
						"description": "The protocol used by the AI provider (currently only openai/v1 is supported)"
					},
					"tokens": {
						"type": "array",
						"items": {"type": "string"},
						"description": "API tokens for authentication"
					},
					"tokenFailoverConfig": {
						"type": "object",
						"properties": {
							"enabled": {
								"type": "boolean",
								"description": "Whether token failover is enabled"
							},
							"failureThreshold": {
								"type": "integer",
								"minimum": 1,
								"description": "Number of failures before marking token as unhealthy"
							},
							"successThreshold": {
								"type": "integer",
								"minimum": 1,
								"description": "Number of successes before marking token as healthy"
							},
							"healthCheckInterval": {
								"type": "integer",
								"description": "Health check interval in milliseconds"
							},
							"healthCheckTimeout": {
								"type": "integer",
								"description": "Health check timeout in milliseconds"
							},
							"healthCheckModel": {
								"type": "string",
								"description": "Model to use for health checks"
						}
					},
					"additionalProperties": false
				}
			},
			"additionalProperties": false
			}
		},
		"required": ["name", "configurations"],
		"additionalProperties": false
	}`)
}
