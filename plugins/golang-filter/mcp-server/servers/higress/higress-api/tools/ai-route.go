package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alibaba/higress/plugins/golang-filter/mcp-server/servers/higress"
	"github.com/alibaba/higress/plugins/golang-filter/mcp-session/common"
	"github.com/mark3labs/mcp-go/mcp"
)

type AIRoute struct {
	Name               string            `json:"name"`
	Version            string            `json:"version,omitempty"`
	Domains            []string          `json:"domains,omitempty"`
	PathPredicate      *RoutePath        `json:"pathPredicate,omitempty"`
	HeaderPredicates   []RouteMatch      `json:"headerPredicates,omitempty"`
	URLParamPredicates []RouteMatch      `json:"urlParamPredicates,omitempty"`
	Upstreams          []AIUpstream      `json:"upstreams,omitempty"`
	ModelPredicates    []RouteMatch      `json:"modelPredicates,omitempty"`
	AuthConfig         *RouteAuthConfig  `json:"authConfig,omitempty"`
	FallbackConfig     *FallbackConfig   `json:"fallbackConfig,omitempty"`
}

type AIUpstream struct {
	Provider      string            `json:"provider"`
	Weight        int               `json:"weight"`
	ModelMapping  map[string]string `json:"modelMapping,omitempty"`
}

type FallbackConfig struct {
	Enabled          bool         `json:"enabled"`
	Upstreams        []AIUpstream `json:"upstreams,omitempty"`
	FallbackStrategy string       `json:"fallbackStrategy,omitempty"`
	ResponseCodes    []string     `json:"responseCodes,omitempty"`
}

type AIRouteResponse = higress.APIResponse[AIRoute]

func RegisterAIRouteTools(mcpServer *common.MCPServer, client *higress.HigressClient) {
	mcpServer.AddTool(
		mcp.NewTool("list-ai-routes", mcp.WithDescription("List all available AI routes")),
		handleListAIRoutes(client),
	)
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema("get-ai-route", "Get detailed information about a specific AI route", getAIRouteSchema()),
		handleGetAIRoute(client),
	)
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema("add-ai-route", "Add a new AI route", getAddAIRouteSchema()),
		handleAddAIRoute(client),
	)
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema("update-ai-route", "Update an existing AI route", getUpdateAIRouteSchema()),
		handleUpdateAIRoute(client),
	)
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema("delete-ai-route", "Delete an existing AI route", getAIRouteSchema()),
		handleDeleteAIRoute(client),
	)
}

func handleListAIRoutes(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		respBody, err := client.Get("/v1/ai-routes")
		if err != nil {
			return nil, fmt.Errorf("failed to list AI routes: %w", err)
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

func handleGetAIRoute(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments
		name, ok := arguments["name"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'name' argument")
		}

		respBody, err := client.Get(fmt.Sprintf("/v1/ai-routes/%s", name))
		if err != nil {
			return nil, fmt.Errorf("failed to get AI route '%s': %w", name, err)
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

func handleAddAIRoute(client *higress.HigressClient) common.ToolHandlerFunc {
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
		if _, ok := configurations["pathPredicate"]; !ok {
			return nil, fmt.Errorf("missing required field 'pathPredicate' in configurations")
		}
		if _, ok := configurations["upstreams"]; !ok {
			return nil, fmt.Errorf("missing required field 'upstreams' in configurations")
		}

		respBody, err := client.Post("/v1/ai-routes", configurations)
		if err != nil {
			return nil, fmt.Errorf("failed to add AI route: %w", err)
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

func handleUpdateAIRoute(client *higress.HigressClient) common.ToolHandlerFunc {
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

		// Get current AI route configuration to merge with updates
		currentBody, err := client.Get(fmt.Sprintf("/v1/ai-routes/%s", name))
		if err != nil {
			return nil, fmt.Errorf("failed to get current AI route configuration: %w", err)
		}

		var response AIRouteResponse
		if err := json.Unmarshal(currentBody, &response); err != nil {
			return nil, fmt.Errorf("failed to parse current AI route response: %w", err)
		}

		currentConfig := response.Data

		// Update configurations using JSON marshal/unmarshal for type conversion
		configBytes, err := json.Marshal(configurations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal configurations: %w", err)
		}

		var newConfig AIRoute
		if err := json.Unmarshal(configBytes, &newConfig); err != nil {
			return nil, fmt.Errorf("failed to parse AI route configurations: %w", err)
		}

		// Merge configurations (overwrite with new values where provided)
		// Note: name cannot be updated
		if newConfig.Domains != nil {
			currentConfig.Domains = newConfig.Domains
		}
		if newConfig.PathPredicate != nil {
			currentConfig.PathPredicate = newConfig.PathPredicate
		}
		if newConfig.HeaderPredicates != nil {
			currentConfig.HeaderPredicates = newConfig.HeaderPredicates
		}
		if newConfig.URLParamPredicates != nil {
			currentConfig.URLParamPredicates = newConfig.URLParamPredicates
		}
		if newConfig.Upstreams != nil {
			currentConfig.Upstreams = newConfig.Upstreams
		}
		if newConfig.ModelPredicates != nil {
			currentConfig.ModelPredicates = newConfig.ModelPredicates
		}
		if newConfig.AuthConfig != nil {
			currentConfig.AuthConfig = newConfig.AuthConfig
		}
		if newConfig.FallbackConfig != nil {
			currentConfig.FallbackConfig = newConfig.FallbackConfig
		}

		respBody, err := client.Put(fmt.Sprintf("/v1/ai-routes/%s", name), currentConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to update AI route '%s': %w", name, err)
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

func handleDeleteAIRoute(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments
		name, ok := arguments["name"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'name' argument")
		}

		respBody, err := client.Delete(fmt.Sprintf("/v1/ai-routes/%s", name))
		if err != nil {
			return nil, fmt.Errorf("failed to delete AI route '%s': %w", name, err)
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

func getAIRouteSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"description": "The name of the AI route"
			}
		},
		"required": ["name"],
		"additionalProperties": false
	}`)
}

func getAddAIRouteSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"configurations": {
				"type": "object",
				"properties": {
					"name": {
						"type": "string",
						"description": "The name of the AI route"
					},
					"domains": {
						"type": "array",
						"items": {"type": "string"},
						"description": "List of domain names"
					},
					"pathPredicate": {
						"type": "object",
						"properties": {
							"matchType": {"type": "string", "enum": ["PRE", "EQUAL", "REGULAR"], "description": "Match type of path"},
							"matchValue": {"type": "string", "description": "Value to match"},
							"caseSensitive": {"type": "boolean", "description": "Whether matching is case sensitive"}
						},
						"required": ["matchType", "matchValue"],
						"description": "Path matching configuration"
					},
					"headerPredicates": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"key": {"type": "string", "description": "Header key name"},
								"matchType": {"type": "string", "enum": ["PRE", "EQUAL", "REGULAR"], "description": "Match type of header"},
								"matchValue": {"type": "string", "description": "Value to match"}
							},
							"required": ["key", "matchType", "matchValue"]
						},
						"description": "List of header match conditions"
					},
					"urlParamPredicates": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"key": {"type": "string", "description": "Parameter key name"},
								"matchType": {"type": "string", "enum": ["PRE", "EQUAL", "REGULAR"], "description": "Match type of URL parameter"},
								"matchValue": {"type": "string", "description": "Value to match"}
							},
							"required": ["key", "matchType", "matchValue"]
						},
						"description": "List of URL parameter match conditions"
					},
					"upstreams": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"provider": {"type": "string", "description": "AI provider name"},
								"weight": {"type": "integer", "minimum": 0, "maximum": 100, "description": "Weight for load balancing"},
								"modelMapping": {
									"type": "object",
									"additionalProperties": {"type": "string"},
									"description": "Model mapping configuration, use '*' for wildcard"
								}
							},
							"required": ["provider", "weight"]
						},
						"description": "List of AI provider upstreams"
					},
					"modelPredicates": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"key": {"type": "string", "enum": ["model"], "description": "Key of the model"},
								"matchType": {"type": "string", "enum": ["PRE", "EQUAL"], "description": "Match type for model"},
								"matchValue": {"type": "string", "description": "Model value to match"}
							},
							"required": ["matchType", "matchValue"]
						},
						"description": "List of model match conditions"
					},
					"authConfig": {
						"type": "object",
						"properties": {
							"enabled": {"type": "boolean", "description": "Whether authentication is enabled"},
							"allowedConsumers": {
								"type": ["array", "null"],
								"items": {"type": "string"},
								"description": "List of allowed consumer names"
							}
						},
						"description": "Authentication configuration"
					},
					"fallbackConfig": {
						"type": "object",
						"properties": {
							"enabled": {"type": "boolean", "description": "Whether fallback is enabled"},
							"upstreams": {
								"type": "array",
								"items": {
									"type": "object",
									"properties": {
										"provider": {"type": "string", "description": "Fallback AI provider name"},
										"weight": {"type": "integer", "minimum": 0, "maximum": 100, "description": "Weight for load balancing"},
										"modelMapping": {
											"type": "object",
											"additionalProperties": {"type": "string"},
											"description": "Model mapping configuration, use '*' for wildcard"
										}
									},
									"required": ["provider", "weight"]
								},
								"description": "List of fallback AI provider upstreams"
							},
							"fallbackStrategy": {
								"type": "string",
								"enum": ["RAND", "ROUND_ROBIN", "WEIGHTED"],
								"description": "Fallback strategy for selecting upstreams"
							},
							"responseCodes": {
								"type": "array",
								"items": {"type": "string"},
								"description": "Response codes that trigger fallback (e.g., '5xx', '4xx', '503')"
							}
						},
						"required": ["enabled"],
						"description": "Fallback configuration for error handling"
					}
				},
				"required": ["name", "pathPredicate", "upstreams"],
				"additionalProperties": false
			}
		},
		"required": ["configurations"],
		"additionalProperties": false
	}`)
}

func getUpdateAIRouteSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"description": "The name of the AI route to update"
			},
			"configurations": {
				"type": "object",
				"properties": {
					"domains": {
						"type": "array",
						"items": {"type": "string"},
						"description": "List of domain names"
					},
					"pathPredicate": {
						"type": "object",
						"properties": {
							"matchType": {"type": "string", "enum": ["PRE", "EQUAL", "REGULAR"], "description": "Match type of path"},
							"matchValue": {"type": "string", "description": "Value to match"},
							"caseSensitive": {"type": "boolean", "description": "Whether matching is case sensitive"}
						},
						"required": ["matchType", "matchValue"],
						"description": "Path matching configuration"
					},
					"headerPredicates": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"key": {"type": "string", "description": "Header key name"},
								"matchType": {"type": "string", "enum": ["PRE", "EQUAL", "REGULAR"], "description": "Match type of header"},
								"matchValue": {"type": "string", "description": "Value to match"},
								"caseSensitive": {"type": "boolean", "description": "Whether matching is case sensitive"}
							},
							"required": ["key", "matchType", "matchValue"]
						},
						"description": "List of header match conditions"
					},
					"urlParamPredicates": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"key": {"type": "string", "description": "Parameter key name"},
								"matchType": {"type": "string", "enum": ["PRE", "EQUAL", "REGULAR"], "description": "Match type of URL parameter"},
								"matchValue": {"type": "string", "description": "Value to match"},
								"caseSensitive": {"type": "boolean", "description": "Whether matching is case sensitive"}
							},
							"required": ["key", "matchType", "matchValue"]
						},
						"description": "List of URL parameter match conditions"
					},
					"upstreams": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"provider": {"type": "string", "description": "AI provider name"},
								"weight": {"type": "integer", "minimum": 0, "maximum": 100, "description": "Weight for load balancing"},
								"modelMapping": {
									"type": "object",
									"additionalProperties": {"type": "string"},
									"description": "Model mapping configuration, use '*' for wildcard"
								}
							},
							"required": ["provider", "weight"]
						},
						"description": "List of AI provider upstreams"
					},
					"modelPredicates": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"matchType": {"type": "string", "enum": ["PRE", "EQUAL"], "description": "Match type for model"},
								"matchValue": {"type": "string", "description": "Model value to match"},
								"caseSensitive": {"type": ["boolean", "null"], "description": "Whether matching is case sensitive"}
							},
							"required": ["matchType", "matchValue"]
						},
						"description": "List of model match conditions"
					},
					"authConfig": {
						"type": "object",
						"properties": {
							"enabled": {"type": "boolean", "description": "Whether authentication is enabled"},
							"allowedConsumers": {
								"type": ["array", "null"],
								"items": {"type": "string"},
								"description": "List of allowed consumer names"
							}
						},
						"description": "Authentication configuration"
					},
					"fallbackConfig": {
						"type": "object",
						"properties": {
							"enabled": {"type": "boolean", "description": "Whether fallback is enabled"},
							"upstreams": {
								"type": "array",
								"items": {
									"type": "object",
									"properties": {
										"provider": {"type": "string", "description": "Fallback AI provider name"},
										"weight": {"type": "integer", "minimum": 0, "maximum": 100, "description": "Weight for load balancing"},
										"modelMapping": {
											"type": "object",
											"additionalProperties": {"type": "string"},
											"description": "Model mapping configuration, use '*' for wildcard"
										}
									},
									"required": ["provider", "weight"]
								},
								"description": "List of fallback AI provider upstreams"
							},
							"fallbackStrategy": {
								"type": "string",
								"enum": ["RAND", "ROUND_ROBIN", "WEIGHTED"],
								"description": "Fallback strategy for selecting upstreams"
							},
							"responseCodes": {
								"type": "array",
								"items": {"type": "string"},
								"description": "Response codes that trigger fallback (e.g., '5xx', '4xx', '503')"
							}
						},
						"description": "Fallback configuration for error handling"
					}
				},
				"additionalProperties": false
			}
		},
		"required": ["name", "configurations"],
		"additionalProperties": false
	}`)
}
