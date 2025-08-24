package plugins

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alibaba/higress/plugins/golang-filter/mcp-server/servers/higress"
	"github.com/alibaba/higress/plugins/golang-filter/mcp-session/common"
	"github.com/mark3labs/mcp-go/mcp"
)

const CustomResponsePluginName = "custom-response"

type CustomResponseConfig struct {
	StatusCode     int      `json:"status_code,omitempty"`
	Headers        []string `json:"headers,omitempty"`
	Body           string   `json:"body,omitempty"`
	EnableOnStatus []int    `json:"enable_on_status,omitempty"`
}

type CustomResponseInstance = PluginInstance[CustomResponseConfig]

type CustomResponseResponse = higress.APIResponse[CustomResponseInstance]

func RegisterCustomResponsePluginTools(mcpServer *common.MCPServer, client *higress.HigressClient) {
	mcpServer.AddTool(
		mcp.NewToolWithRawSchema(fmt.Sprintf("update-%s-plugin", CustomResponsePluginName), "Update custom response plugin configuration", getAddOrUpdateCustomResponseConfigSchema()),
		handleAddOrUpdateCustomResponseConfig(client),
	)
}

func handleAddOrUpdateCustomResponseConfig(client *higress.HigressClient) common.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, err := ParsePluginUpdateParams(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Merge current config with new config
		mergeFunc := func(current *CustomResponseInstance, newConfig CustomResponseConfig) {
			if newConfig.StatusCode != 0 {
				current.Configurations.StatusCode = newConfig.StatusCode
			}
			if newConfig.Headers != nil {
				current.Configurations.Headers = newConfig.Headers
			}
			if newConfig.Body != "" {
				current.Configurations.Body = newConfig.Body
			}
			if newConfig.EnableOnStatus != nil {
				current.Configurations.EnableOnStatus = newConfig.EnableOnStatus
			}
		}

		respBody, err := UpdatePluginConfig(client, CustomResponsePluginName, params, mergeFunc)
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

func getAddOrUpdateCustomResponseConfigSchema() json.RawMessage {
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
					"status_code": {
						"type": "integer",
						"minimum": 100,
						"maximum": 599,
						"default": 200,
						"description": "Custom HTTP response status code"
					},
					"headers": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Custom HTTP response headers, keys and values separated by ="
					},
					"body": {
						"type": "string",
						"description": "Custom HTTP response body"
					},
					"enable_on_status": {
						"type": "array",
						"items": {
							"type": "integer",
							"minimum": 100,
							"maximum": 599
						},
						"description": "Match original status codes to generate custom responses; if not specified, the original status code is not checked"
					}
				},
				"additionalProperties": false
			}
		},
		"required": ["scope", "enabled", "configurations"],
		"additionalProperties": false
	}`)
}
