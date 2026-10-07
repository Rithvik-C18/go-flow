package execution

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

type GoogleDocsNode struct {
	credentials string
	operation   string
	documentId  string
	title       string
	content     string
	text        string
}

func NewGoogleDocsNode(config map[string]any, credentials string) (*GoogleDocsNode, error) {
	if credentials == "" {
		return nil, fmt.Errorf("google docs node: GOOGLE_CREDENTIALS not configured")
	}

	operation, ok := config["operation"].(string)
	if !ok || operation == "" {
		return nil, fmt.Errorf("google docs node: missing or invalid operation")
	}

	node := &GoogleDocsNode{
		credentials: credentials,
		operation:   operation,
	}

	switch operation {
	case "create":
		title, ok := config["title"].(string)
		if !ok || title == "" {
			return nil, fmt.Errorf("google docs node: missing title for create operation")
		}
		node.title = title

		if content, ok := config["content"].(string); ok {
			node.content = content
		}

	case "read":
		documentId, ok := config["documentId"].(string)
		if !ok || documentId == "" {
			return nil, fmt.Errorf("google docs node: missing documentId for read operation")
		}
		node.documentId = documentId

	case "append":
		documentId, ok := config["documentId"].(string)
		if !ok || documentId == "" {
			return nil, fmt.Errorf("google docs node: missing documentId for append operation")
		}
		node.documentId = documentId

		text, ok := config["text"].(string)
		if !ok || text == "" {
			return nil, fmt.Errorf("google docs node: missing text for append operation")
		}
		node.text = text

	default:
		return nil, fmt.Errorf("google docs node: unsupported operation %q (supported: create, read, append)", operation)
	}

	return node, nil
}

func (g *GoogleDocsNode) Execute(ctx *ExecutionContext) error {
	bgCtx := context.Background()

	// Check if credentials is a file path or JSON content
	var opts []option.ClientOption
	if _, err := os.Stat(g.credentials); err == nil {
		// It's a file path
		opts = []option.ClientOption{option.WithCredentialsFile(g.credentials)}
	} else {
		// Assume it's JSON content
		opts = []option.ClientOption{option.WithCredentialsJSON([]byte(g.credentials))}
	}

	docsService, err := docs.NewService(bgCtx, opts...)
	if err != nil {
		return fmt.Errorf("google docs node: failed to create docs service: %w", err)
	}

	switch g.operation {
	case "create":
		return g.createDocument(bgCtx, docsService, ctx)
	case "read":
		return g.readDocument(bgCtx, docsService, ctx)
	case "append":
		return g.appendText(bgCtx, docsService, ctx)
	default:
		return fmt.Errorf("google docs node: unsupported operation %q", g.operation)
	}
}

func (g *GoogleDocsNode) createDocument(ctx context.Context, docsService *docs.Service, execCtx *ExecutionContext) error {
	title := g.title
	if resolvedTitle, err := resolveValue(g.title, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		title = fmt.Sprint(resolvedTitle)
	}

	doc := &docs.Document{
		Title: title,
	}

	if g.content != "" {
		content := g.content
		if resolvedContent, err := resolveValue(g.content, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
			content = fmt.Sprint(resolvedContent)
		}
		doc.Body = &docs.Body{
			Content: []*docs.StructuralElement{
				{
					Paragraph: &docs.Paragraph{
						Elements: []*docs.ParagraphElement{
							{
								TextRun: &docs.TextRun{
									Content: content,
								},
							},
						},
					},
				},
			},
		}
	}

	createdDoc, err := docsService.Documents.Create(doc).Do()
	if err != nil {
		return fmt.Errorf("google docs node: failed to create document: %w", err)
	}

	execCtx.Output = map[string]any{
		"documentId": createdDoc.DocumentId,
		"title":      createdDoc.Title,
		"url":        fmt.Sprintf("https://docs.google.com/document/d/%s/edit", createdDoc.DocumentId),
	}

	return nil
}

func (g *GoogleDocsNode) readDocument(ctx context.Context, docsService *docs.Service, execCtx *ExecutionContext) error {
	documentId := g.documentId
	if resolvedId, err := resolveValue(g.documentId, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		documentId = fmt.Sprint(resolvedId)
	}

	doc, err := docsService.Documents.Get(documentId).Do()
	if err != nil {
		return fmt.Errorf("google docs node: failed to read document: %w", err)
	}

	content := extractTextFromDocument(doc)

	execCtx.Output = map[string]any{
		"documentId": doc.DocumentId,
		"title":      doc.Title,
		"content":    content,
		"url":        fmt.Sprintf("https://docs.google.com/document/d/%s/edit", doc.DocumentId),
	}

	return nil
}

func (g *GoogleDocsNode) appendText(ctx context.Context, docsService *docs.Service, execCtx *ExecutionContext) error {
	documentId := g.documentId
	if resolvedId, err := resolveValue(g.documentId, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		documentId = fmt.Sprint(resolvedId)
	}

	text := g.text
	if resolvedText, err := resolveValue(g.text, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		text = fmt.Sprint(resolvedText)
	}

	// Get the document to find the end index
	doc, err := docsService.Documents.Get(documentId).Do()
	if err != nil {
		return fmt.Errorf("google docs node: failed to get document for append: %w", err)
	}

	endIndex := doc.Body.Content[len(doc.Body.Content)-1].EndIndex

	batchUpdate := &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			{
				InsertText: &docs.InsertTextRequest{
					Text:       text,
					Location:   &docs.Location{Index: endIndex},
					EndOfSegmentLocation: &docs.EndOfSegmentLocation{},
				},
			},
		},
	}

	_, err = docsService.Documents.BatchUpdate(documentId, batchUpdate).Do()
	if err != nil {
		return fmt.Errorf("google docs node: failed to append text: %w", err)
	}

	execCtx.Output = map[string]any{
		"documentId": documentId,
		"message":    "Text appended successfully",
	}

	return nil
}

func extractTextFromDocument(doc *docs.Document) string {
	var content string
	for _, element := range doc.Body.Content {
		if element.Paragraph != nil {
			for _, paragraphElement := range element.Paragraph.Elements {
				if paragraphElement.TextRun != nil {
					content += paragraphElement.TextRun.Content
				}
			}
			content += "\n"
		}
	}
	return content
}