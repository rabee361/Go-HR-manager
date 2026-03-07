package main

import (
	"bytes"
	"testing"
)

// TestExportToExcel demonstrates a table-driven test for our generic Excel exporter.
func TestExportToExcel(t *testing.T) {
	// 1. Define our test data structure (The "Table")
	type testCase struct {
		name    string
		data    []Department
		headers []string
		wantErr bool
	}

	// 2. Define our specific cases
	tests := []testCase{
		{
			name: "Successful export with data",
			data: []Department{
				{ID: 1, Name: "HR", Description: "Human Resources"},
				{ID: 2, Name: "IT", Description: "IT Support"},
			},
			headers: []string{"ID", "Name", "Description"},
			wantErr: false,
		},
		{
			name:    "Successful export with empty data",
			data:    []Department{},
			headers: []string{"ID", "Name", "Description"},
			wantErr: false,
		},
	}

	// 3. Define the mapper function (same as in main.go/handlers)
	mapper := func(d Department) []string {
		return []string{
			"1", // Simplified for testing
			d.Name,
			d.Description,
		}
	}

	// 4. Run the tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// bytes.Buffer satisfies io.Writer, so we don't need a real file!
			var buf bytes.Buffer

			err := ExportToExcel(&buf, tc.data, tc.headers, mapper)

			// Check if we expected an error but didn't get one (or vice versa)
			if (err != nil) != tc.wantErr {
				t.Errorf("ExportToExcel() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// If success, verify we actually wrote something to the buffer
			if !tc.wantErr && buf.Len() == 0 {
				t.Error("ExportToExcel() wrote 0 bytes to the buffer")
			}
		})
	}
}
