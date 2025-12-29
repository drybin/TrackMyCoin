package webapi

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type IGoogleSheets interface {
	ReadSpreadsheet(ctx context.Context, spreadsheetID string, readRange string) (*sheets.ValueRange, error)
	GetSpreadsheetInfo(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error)
	UpdateSpreadsheet(ctx context.Context, spreadsheetID string, writeRange string, values [][]interface{}) error
}

type GoogleSheets struct {
	service *sheets.Service
}

// NewGoogleSheetsWithServiceAccount creates a client using service account file
func NewGoogleSheetsWithServiceAccount(ctx context.Context, credentialsFilePath string) (*GoogleSheets, error) {
	credentialsJSON, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read service account file: %w", err)
	}

	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("unable to create Sheets client: %w", err)
	}

	return &GoogleSheets{
		service: srv,
	}, nil
}

// NewGoogleSheetsWithCredentialsJSON creates a client using credentials JSON bytes
func NewGoogleSheetsWithCredentialsJSON(ctx context.Context, credentialsJSON []byte) (*GoogleSheets, error) {
	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("unable to create Sheets client: %w", err)
	}

	return &GoogleSheets{
		service: srv,
	}, nil
}

// NewGoogleSheetsWithAPIKey creates a client using API key (simpler but more limited)
func NewGoogleSheetsWithAPIKey(ctx context.Context, apiKey string) (*GoogleSheets, error) {
	srv, err := sheets.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("unable to create Sheets client: %w", err)
	}

	return &GoogleSheets{
		service: srv,
	}, nil
}

func (g *GoogleSheets) ReadSpreadsheet(ctx context.Context, spreadsheetID string, readRange string) (*sheets.ValueRange, error) {
	resp, err := g.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %w", err)
	}

	return resp, nil
}

// GetSpreadsheetInfo returns basic information about the spreadsheet including sheet names
func (g *GoogleSheets) GetSpreadsheetInfo(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
	spreadsheet, err := g.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve spreadsheet info: %w", err)
	}

	return spreadsheet, nil
}

// UpdateSpreadsheet записывает данные в таблицу
func (g *GoogleSheets) UpdateSpreadsheet(ctx context.Context, spreadsheetID string, writeRange string, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := g.service.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).
		ValueInputOption("RAW").
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("unable to update data in sheet: %w", err)
	}

	return nil
}

