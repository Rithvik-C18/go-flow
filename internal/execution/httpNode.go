package execution

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type HttpNode struct {
	url         string
	httpMethod  string
	queryParams map[string]any
	headers     map[string]any
	body        any
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

	node := &HttpNode{
		url:        url,
		httpMethod: method,
	}

	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
	default:
		return nil, fmt.Errorf("http node: unsupported method %q", method)
	}

	if v, ok := config["queryParameters"]; ok {
		qp, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("http node: queryParameters must be an object")
		}
		node.queryParams = qp
	}

	if v, ok := config["headers"]; ok {
		headers, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("http node: headers must be an object")
		}
		node.headers = headers
	}

	if v, ok := config["body"]; ok {
		node.body = v
	}

	return node, nil
}

func (h *HttpNode) Execute(ctx *ExecutionContext) error {
	if err := validateHTTPURL(h.url); err != nil {
		return err
	}
	client := resty.New().SetTransport(publicHTTPTransport()).SetTimeout(30 * time.Second)
	req := client.R()

	if len(h.queryParams) > 0 {
		queryParams := make(map[string]string, len(h.queryParams))
		for key, value := range h.queryParams {
			resolved, err := resolveValue(value, ctx.Input, ctx.Input, ctx.Params)
			if err != nil {
				return fmt.Errorf("http node: query parameter %q: %w", key, err)
			}
			queryParams[key] = fmt.Sprint(resolved)
		}
		req.SetQueryParams(queryParams)
	}

	if len(h.headers) > 0 {
		headers := make(map[string]string, len(h.headers))
		for key, value := range h.headers {
			resolved, err := resolveValue(value, ctx.Input, ctx.Input, ctx.Params)
			if err != nil {
				return fmt.Errorf("http node: header %q: %w", key, err)
			}
			headers[key] = fmt.Sprint(resolved)
		}
		req.SetHeaders(headers)
	}

	if h.body != nil {
		switch body := h.body.(type) {
		case map[string]any:
			resolved := make(map[string]any, len(body))
			for key, value := range body {
				val, err := resolveValue(value, ctx.Input, ctx.Input, ctx.Params)
				if err != nil {
					return fmt.Errorf("http node: body field %q: %w", key, err)
				}
				resolved[key] = val
			}
			req.SetBody(resolved)
		default:
			req.SetBody(h.body)
		}
	}

	var (
		resp *resty.Response
		err  error
	)

	switch h.httpMethod {
	case "GET":
		resp, err = req.Get(h.url)
	case "POST":
		resp, err = req.Post(h.url)
	case "PUT":
		resp, err = req.Put(h.url)
	case "PATCH":
		resp, err = req.Patch(h.url)
	case "DELETE":
		resp, err = req.Delete(h.url)
	}

	if err != nil {
		return fmt.Errorf("resty request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("http request returned status: %d", resp.StatusCode())
	}

	result := map[string]any{}
	if len(resp.Body()) > 0 {
		var body any
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			if fields, ok := body.(map[string]any); ok {
				for key, value := range fields {
					result[key] = value
				}
			} else {
				result["body"] = body
			}
		} else {
			result["body"] = string(resp.Body())
		}
	}
	result["statusCode"] = resp.StatusCode()
	ctx.Output = result

	return nil
}
