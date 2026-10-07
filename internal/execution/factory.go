package execution

import (
	"fmt"

	"github.com/Rithvik-C18/go-flow/internal/config"
	"github.com/Rithvik-C18/go-flow/internal/workflow"
)

var appConfig config.Config

func SetConfig(cfg config.Config) {
	appConfig = cfg
}

func IsValidNodeType(nodeType string) bool {
	switch nodeType {
	case "http", "if", "gemini", "google_docs", "google_sheets":
		return true
	default:
		return false
	}
}

func buildNode(def *workflow.NodeDefinition) (Node, error) {
	switch def.Type {
	case "http":
		return NewHttpNode(def.Config)
	case "if":
		return NewIfNode(def.Config)
	case "gemini":
		return NewGeminiNode(def.Config, appConfig.Integrations.GeminiAPIKey)
	case "google_docs":
		return NewGoogleDocsNode(def.Config, appConfig.Integrations.GoogleCredentials)
	case "google_sheets":
		return NewGoogleSheetsNode(def.Config, appConfig.Integrations.GoogleCredentials)
	default:
		return nil, fmt.Errorf("Unknown node type")
	}
}
