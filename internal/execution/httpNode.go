package execution

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type HttpNode struct {
	url        string
	httpMethod string
}

func NewHttpNode(config map[string]any) (*HttpNode, error) {
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return nil, fmt.Errorf("http node: missing or invalid url")
	}

	method, ok := config["httpMethod"].(string)
	if !ok || method == "" {
		return nil, fmt.Errorf("http node: missing or invalid httpMethod")
	}

	return &HttpNode{
		url:        url,
		httpMethod: method,
	}, nil
}

func (h *HttpNode) Execute(ctx *ExecutionContext) error {

	client := resty.New()

	switch h.httpMethod {
	case "GET":
		resp, err := client.R().Get(h.url)
		if err != nil {
			return fmt.Errorf("resty request failed: %w", err)
		}

		if resp.IsError() {
			return fmt.Errorf("http request returned status: %d", resp.StatusCode())
		}

		var result map[string]any

		err = json.Unmarshal(resp.Body(), &result)
		if err != nil {
			return fmt.Errorf("failed to parse JSON response: %w", err)
		}

		ctx.Output = result
	default:
		return fmt.Errorf("unsupported method: %s", h.httpMethod)
	}

	return nil
}

