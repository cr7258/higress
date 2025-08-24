package plugins

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alibaba/higress/plugins/golang-filter/mcp-server/servers/higress"
	"github.com/alibaba/higress/plugins/golang-filter/mcp-session/common"
	"github.com/mark3labs/mcp-go/mcp"
)

type PluginUpdateParams struct {
	Scope          string
	ResourceName   string
	Enabled        bool
	Configurations interface{}
}

func ParsePluginUpdateParams(arguments map[string]interface{}) (*PluginUpdateParams, error) {
	// Parse required parameters
	scope, ok := arguments["scope"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'scope' argument")
	}

	if !IsValidScope(scope) {
		return nil, fmt.Errorf("invalid scope '%s', must be one of: %v", scope, ValidScopes)
	}

	enabled, ok := arguments["enabled"].(bool)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'enabled' argument")
	}

	configurations, ok := arguments["configurations"]
	if !ok {
		return nil, fmt.Errorf("missing 'configurations' argument")
	}

	// Parse resource_name for non-global scopes
	var resourceName string
	if scope != ScopeGlobal {
		resourceName, ok = arguments["resource_name"].(string)
		if !ok || resourceName == "" {
			return nil, fmt.Errorf("'resource_name' is required for scope '%s'", scope)
		}
	}

	return &PluginUpdateParams{
		Scope:          scope,
		ResourceName:   resourceName,
		Enabled:        enabled,
		Configurations: configurations,
	}, nil
}

// UpdatePluginConfig is a generic function to update plugin configurations
func UpdatePluginConfig[T any](
	client *higress.HigressClient,
	pluginName string,
	params *PluginUpdateParams,
	mergeFunc func(current *PluginInstance[T], newConfig T),
) (string, error) {
	// Build API path
	path := BuildPluginPath(pluginName, params.Scope, params.ResourceName)

	// Get current configuration
	currentBody, err := client.Get(path)
	if err != nil {
		return "", fmt.Errorf("failed to get current %s configuration: %w", pluginName, err)
	}

	var response higress.APIResponse[PluginInstance[T]]
	if err := json.Unmarshal(currentBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse current %s response: %w", pluginName, err)
	}

	currentConfig := response.Data
	currentConfig.Enabled = params.Enabled
	currentConfig.Scope = params.Scope

	// Convert the input configurations to the specific type and merge
	configBytes, err := json.Marshal(params.Configurations)
	if err != nil {
		return "", fmt.Errorf("failed to marshal configurations: %w", err)
	}

	var newConfig T
	if err := json.Unmarshal(configBytes, &newConfig); err != nil {
		return "", fmt.Errorf("failed to parse %s configurations: %w", pluginName, err)
	}

	// Apply the merge function
	mergeFunc(&currentConfig, newConfig)

	// Update the configuration
	respBody, err := client.Put(path, currentConfig)
	if err != nil {
		return "", fmt.Errorf("failed to update %s config at scope '%s': %w", pluginName, params.Scope, err)
	}

	return string(respBody), nil
}

// RegisterCommonPluginTools registers all common plugin management tools
func RegisterCommonPluginTools(mcpServer *common.MCPServer, client *higress.HigressClient) {
	// Get plugin configuration
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema("get-plugin", "Get configuration for a specific plugin", getPluginConfigSchema()),
		handleGetPluginConfig(client),
	)

	// Delete plugin configuration
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema("delete-plugin", "Delete configuration for a specific plugin", getPluginConfigSchema()),
		handleDeletePluginConfig(client),
	)
}

func handleGetPluginConfig(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments

		// Parse required parameters
		pluginName, ok := arguments["name"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'name' argument")
		}

		scope, ok := arguments["scope"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'scope' argument")
		}

		if !IsValidScope(scope) {
			return nil, fmt.Errorf("invalid scope '%s', must be one of: %v", scope, ValidScopes)
		}

		// Parse resource_name (required for non-global scopes)
		var resourceName string
		if scope != ScopeGlobal {
			resourceName, ok = arguments["resource_name"].(string)
			if !ok || resourceName == "" {
				return nil, fmt.Errorf("'resource_name' is required for scope '%s'", scope)
			}
		}

		// Build API path and make request
		path := BuildPluginPath(pluginName, scope, resourceName)
		respBody, err := client.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin config for '%s' at scope '%s': %w", pluginName, scope, err)
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

func handleDeletePluginConfig(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments

		// Parse required parameters
		pluginName, ok := arguments["name"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'name' argument")
		}

		scope, ok := arguments["scope"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'scope' argument")
		}

		if !IsValidScope(scope) {
			return nil, fmt.Errorf("invalid scope '%s', must be one of: %v", scope, ValidScopes)
		}

		// Parse resource_name (required for non-global scopes)
		var resourceName string
		if scope != ScopeGlobal {
			resourceName, ok = arguments["resource_name"].(string)
			if !ok || resourceName == "" {
				return nil, fmt.Errorf("'resource_name' is required for scope '%s'", scope)
			}
		}

		// Build API path and make request
		path := BuildPluginPath(pluginName, scope, resourceName)
		respBody, err := client.Delete(path)
		if err != nil {
			return nil, fmt.Errorf("failed to delete plugin config for '%s' at scope '%s': %w", pluginName, scope, err)
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

func getPluginConfigSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"description": "The name of the plugin"
			},
			"scope": {
				"type": "string",
				"enum": ["GLOBAL", "DOMAIN", "SERVICE", "ROUTE"],
				"description": "The scope at which the plugin is applied"
			},
			"resource_name": {
				"type": "string",
				"description": "The name of the resource (required for DOMAIN, SERVICE, ROUTE scopes)"
			}
		},
		"required": ["name", "scope"],
		"additionalProperties": false
	}`)
}
