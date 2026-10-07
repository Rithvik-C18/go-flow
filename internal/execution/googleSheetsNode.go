package execution

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsNode struct {
	credentials string
	operation   string
	spreadsheetId string
	title       string
	data        [][]interface{}
	range_      string
	values      []interface{}
}

func NewGoogleSheetsNode(config map[string]any, credentials string) (*GoogleSheetsNode, error) {
	if credentials == "" {
		return nil, fmt.Errorf("google sheets node: GOOGLE_CREDENTIALS not configured")
	}

	operation, ok := config["operation"].(string)
	if !ok || operation == "" {
		return nil, fmt.Errorf("google sheets node: missing or invalid operation")
	}

	node := &GoogleSheetsNode{
		credentials: credentials,
		operation:   operation,
	}

	switch operation {
	case "create":
		title, ok := config["title"].(string)
		if !ok || title == "" {
			return nil, fmt.Errorf("google sheets node: missing title for create operation")
		}
		node.title = title

		if data, ok := config["data"].([]interface{}); ok {
			convertedData := make([][]interface{}, len(data))
			for i, row := range data {
				if rowSlice, ok := row.([]interface{}); ok {
					convertedData[i] = rowSlice
				}
			}
			node.data = convertedData
		}

	case "read":
		spreadsheetId, ok := config["spreadsheetId"].(string)
		if !ok || spreadsheetId == "" {
			return nil, fmt.Errorf("google sheets node: missing spreadsheetId for read operation")
		}
		node.spreadsheetId = spreadsheetId

		range_, ok := config["range"].(string)
		if !ok || range_ == "" {
			node.range_ = "Sheet1!A1:Z1000"
		} else {
			node.range_ = range_
		}

	case "append":
		spreadsheetId, ok := config["spreadsheetId"].(string)
		if !ok || spreadsheetId == "" {
			return nil, fmt.Errorf("google sheets node: missing spreadsheetId for append operation")
		}
		node.spreadsheetId = spreadsheetId

		range_, ok := config["range"].(string)
		if !ok || range_ == "" {
			node.range_ = "Sheet1!A1"
		} else {
			node.range_ = range_
		}

		values, ok := config["values"].([]interface{})
		if !ok || len(values) == 0 {
			return nil, fmt.Errorf("google sheets node: missing values for append operation")
		}
		node.values = values

	case "update":
		spreadsheetId, ok := config["spreadsheetId"].(string)
		if !ok || spreadsheetId == "" {
			return nil, fmt.Errorf("google sheets node: missing spreadsheetId for update operation")
		}
		node.spreadsheetId = spreadsheetId

		range_, ok := config["range"].(string)
		if !ok || range_ == "" {
			return nil, fmt.Errorf("google sheets node: missing range for update operation")
		}
		node.range_ = range_

		values, ok := config["values"].([]interface{})
		if !ok || len(values) == 0 {
			return nil, fmt.Errorf("google sheets node: missing values for update operation")
		}
		node.values = values

	default:
		return nil, fmt.Errorf("google sheets node: unsupported operation %q (supported: create, read, append, update)", operation)
	}

	return node, nil
}

func (g *GoogleSheetsNode) Execute(ctx *ExecutionContext) error {
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

	sheetsService, err := sheets.NewService(bgCtx, opts...)
	if err != nil {
		return fmt.Errorf("google sheets node: failed to create sheets service: %w", err)
	}

	switch g.operation {
	case "create":
		return g.createSpreadsheet(bgCtx, sheetsService, ctx)
	case "read":
		return g.readSpreadsheet(bgCtx, sheetsService, ctx)
	case "append":
		return g.appendValues(bgCtx, sheetsService, ctx)
	case "update":
		return g.updateValues(bgCtx, sheetsService, ctx)
	default:
		return fmt.Errorf("google sheets node: unsupported operation %q", g.operation)
	}
}

func (g *GoogleSheetsNode) createSpreadsheet(ctx context.Context, sheetsService *sheets.Service, execCtx *ExecutionContext) error {
	title := g.title
	if resolvedTitle, err := resolveValue(g.title, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		title = fmt.Sprint(resolvedTitle)
	}

	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}

	if len(g.data) > 0 {
		spreadsheet.Sheets = []*sheets.Sheet{
			{
				Properties: &sheets.SheetProperties{
					Title: "Sheet1",
				},
				Data: []*sheets.GridData{
					{
						RowData: convertToRowData(g.data),
					},
				},
			},
		}
	}

	createdSpreadsheet, err := sheetsService.Spreadsheets.Create(spreadsheet).Do()
	if err != nil {
		return fmt.Errorf("google sheets node: failed to create spreadsheet: %w", err)
	}

	execCtx.Output = map[string]any{
		"spreadsheetId": createdSpreadsheet.SpreadsheetId,
		"title":         createdSpreadsheet.Properties.Title,
		"url":           fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit", createdSpreadsheet.SpreadsheetId),
	}

	return nil
}

func (g *GoogleSheetsNode) readSpreadsheet(ctx context.Context, sheetsService *sheets.Service, execCtx *ExecutionContext) error {
	spreadsheetId := g.spreadsheetId
	if resolvedId, err := resolveValue(g.spreadsheetId, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		spreadsheetId = fmt.Sprint(resolvedId)
	}

	range_ := g.range_
	if resolvedRange, err := resolveValue(g.range_, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		range_ = fmt.Sprint(resolvedRange)
	}

	resp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetId, range_).Do()
	if err != nil {
		return fmt.Errorf("google sheets node: failed to read spreadsheet: %w", err)
	}

	execCtx.Output = map[string]any{
		"spreadsheetId": spreadsheetId,
		"range":         range_,
		"values":        resp.Values,
		"majorDimension": resp.MajorDimension,
	}

	return nil
}

func (g *GoogleSheetsNode) appendValues(ctx context.Context, sheetsService *sheets.Service, execCtx *ExecutionContext) error {
	spreadsheetId := g.spreadsheetId
	if resolvedId, err := resolveValue(g.spreadsheetId, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		spreadsheetId = fmt.Sprint(resolvedId)
	}

	range_ := g.range_
	if resolvedRange, err := resolveValue(g.range_, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		range_ = fmt.Sprint(resolvedRange)
	}

	values := g.values
	if resolvedValues, err := resolveValue(g.values, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		if resolvedSlice, ok := resolvedValues.([]interface{}); ok {
			values = resolvedSlice
		}
	}

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := sheetsService.Spreadsheets.Values.Append(spreadsheetId, range_, valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("google sheets node: failed to append values: %w", err)
	}

	execCtx.Output = map[string]any{
		"spreadsheetId": spreadsheetId,
		"message":       "Values appended successfully",
	}

	return nil
}

func (g *GoogleSheetsNode) updateValues(ctx context.Context, sheetsService *sheets.Service, execCtx *ExecutionContext) error {
	spreadsheetId := g.spreadsheetId
	if resolvedId, err := resolveValue(g.spreadsheetId, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		spreadsheetId = fmt.Sprint(resolvedId)
	}

	range_ := g.range_
	if resolvedRange, err := resolveValue(g.range_, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		range_ = fmt.Sprint(resolvedRange)
	}

	values := g.values
	if resolvedValues, err := resolveValue(g.values, execCtx.Input, execCtx.Input, execCtx.Params); err == nil {
		if resolvedSlice, ok := resolvedValues.([]interface{}); ok {
			values = resolvedSlice
		}
	}

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := sheetsService.Spreadsheets.Values.Update(spreadsheetId, range_, valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("google sheets node: failed to update values: %w", err)
	}

	execCtx.Output = map[string]any{
		"spreadsheetId": spreadsheetId,
		"message":       "Values updated successfully",
	}

	return nil
}

func convertToRowData(data [][]interface{}) []*sheets.RowData {
	rowData := make([]*sheets.RowData, len(data))
	for i, row := range data {
		cells := make([]*sheets.CellData, len(row))
		for j, value := range row {
			strValue := fmt.Sprint(value)
			cells[j] = &sheets.CellData{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: &strValue,
				},
			}
		}
		rowData[i] = &sheets.RowData{
			Values: cells,
		}
	}
	return rowData
}