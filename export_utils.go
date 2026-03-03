package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/xuri/excelize/v2"
)

// ExportToExcel is a generic function that takes any slice of data,
// a set of headers, and a mapper function to convert each item into a string slice (row).
// It then writes the resulting Excel file to the provided io.Writer.
func ExportToExcel[T any](w io.Writer, data []T, headers []string, mapper func(T) []string) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing excel file: %v\n", err)
		}
	}()

	sheetName := "Sheet1"
	// Set headers
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Set rows
	for i, item := range data {
		row := mapper(item)
		for j, val := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			f.SetCellValue(sheetName, cell, val)
		}
	}

	// Write to the output
	if err := f.Write(w); err != nil {
		return fmt.Errorf("writing excel to output: %w", err)
	}

	return nil
}

// SetExcelHeaders sets the necessary HTTP headers for a file download.
func SetExcelHeaders(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xlsx", filename))
}
