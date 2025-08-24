package plugins

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alibaba/higress/plugins/golang-filter/mcp-server/servers/higress"
	"github.com/alibaba/higress/plugins/golang-filter/mcp-session/common"
	"github.com/mark3labs/mcp-go/mcp"
)

const RequestBlockPluginName = "request-block"

type RequestBlockConfig struct {
	BlockBodies   []string `json:"block_bodies,omitempty"`
	BlockHeaders  []string `json:"block_headers,omitempty"`
	BlockUrls     []string `json:"block_urls,omitempty"`
	BlockedCode   int      `json:"blocked_code,omitempty"`
	CaseSensitive bool     `json:"case_sensitive,omitempty"`
}

type RequestBlockInstance = PluginInstance[RequestBlockConfig]

type RequestBlockResponse = higress.APIResponse[RequestBlockInstance]

func RegisterRequestBlockPluginTools(mcpServer *common.MCPServer, client *higress.HigressClient) {
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema(fmt.Sprintf("update-%s-plugin", RequestBlockPluginName), "Update request block plugin configuration", getAddOrUpdateRequestBlockConfigSchema()),
		handleAddOrUpdateRequestBlockConfig(client),
	)
}

func handleAddOrUpdateRequestBlockConfig(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, err := ParsePluginUpdateParams(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Merge current config with new config
		mergeFunc := func(current *RequestBlockInstance, newConfig RequestBlockConfig) {
			if newConfig.BlockBodies != nil {
				current.Configurations.BlockBodies = newConfig.BlockBodies
			}
			if newConfig.BlockHeaders != nil {
				current.Configurations.BlockHeaders = newConfig.BlockHeaders
			}
			if newConfig.BlockUrls != nil {
				current.Configurations.BlockUrls = newConfig.BlockUrls
			}
			if newConfig.BlockedCode != 0 {
				current.Configurations.BlockedCode = newConfig.BlockedCode
			}
			current.Configurations.CaseSensitive = newConfig.CaseSensitive
		}

		respBody, err := UpdatePluginConfig(client, RequestBlockPluginName, params, mergeFunc)
		if err != nil {
			return nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: respBody,
				},
			},
		}, nil
	}
}

func getAddOrUpdateRequestBlockConfigSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"scope": {
				"type": "string",
				"enum": ["GLOBAL", "DOMAIN", "SERVICE", "ROUTE"],
				"description": "The scope at which the plugin is applied"
			},

			"resource_name": {
				"type": "string",
				"description": "The name of the resource (required for DOMAIN, SERVICE, ROUTE scopes)"
			},
			"enabled": {
				"type": "boolean",
				"description": "Whether the plugin is enabled"
			},
			"configurations": {
				"type": "object",
				"properties": {
					"block_bodies": {
						"type": "array",
						"items": {"type": "string"},
						"description": "List of patterns to match against request body content"
					},
					"block_headers": {
						"type": "array",
						"items": {"type": "string"},
						"description": "List of patterns to match against request headers"
					},
					"block_urls": {
						"type": "array",
						"items": {"type": "string"},
						"description": "List of patterns to match against request URLs"
					},
					"blocked_code": {
						"type": "integer",
						"minimum": 100,
						"maximum": 599,
						"description": "HTTP status code to return when a block is matched"
					},
					"case_sensitive": {
						"type": "boolean",
						"description": "Whether the block matching is case sensitive"
					}
				},
				"additionalProperties": false
			}
		},
		"required": ["scope", "enabled", "configurations"],
		"additionalProperties": false
	}`)
}
