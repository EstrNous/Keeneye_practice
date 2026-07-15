package csvparser

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	colFIO       = "fio"
	colRole      = "role"
	colEmail     = "email"
	colGroupName = "group_name"
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Row struct {
	RowNumber int
	FIO       string
	Role      string
	Email     string
	GroupName string
}

type ParseError struct {
	Row     int
	Email   string
	Code    string
	Message string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("row %d: %s", e.Row, e.Message)
}

func Parse(reader io.Reader) ([]Row, []ParseError, error) {
	r := csv.NewReader(reader)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	header, err := r.Read()
	if err == io.EOF {
		return nil, nil, fmt.Errorf("csv file is empty")
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read header: %w", err)
	}

	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}

	colIndex, err := mapHeader(header)
	if err != nil {
		return nil, nil, err
	}

	var rows []Row
	var errors []ParseError
	seenEmails := make(map[string]int)
	rowNum := 1

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			errors = append(errors, ParseError{
				Row:     rowNum,
				Code:    "validation_error",
				Message: fmt.Sprintf("invalid csv row: %v", err),
			})
			continue
		}

		if len(record) < len(colIndex) {
			errors = append(errors, ParseError{
				Row:     rowNum,
				Code:    "validation_error",
				Message: "invalid column count",
			})
			continue
		}

		row := Row{
			RowNumber: rowNum,
			FIO:       strings.TrimSpace(record[colIndex[colFIO]]),
			Role:      strings.ToLower(strings.TrimSpace(record[colIndex[colRole]])),
			Email:     strings.TrimSpace(record[colIndex[colEmail]]),
			GroupName: strings.TrimSpace(record[colIndex[colGroupName]]),
		}

		if row.FIO == "" {
			errors = append(errors, ParseError{Row: rowNum, Email: row.Email, Code: "validation_error", Message: "fio is required"})
			continue
		}
		if row.Role != "student" && row.Role != "teacher" {
			errors = append(errors, ParseError{Row: rowNum, Email: row.Email, Code: "validation_error", Message: "role must be student or teacher"})
			continue
		}
		if row.Email == "" || !emailRe.MatchString(row.Email) {
			errors = append(errors, ParseError{Row: rowNum, Email: row.Email, Code: "validation_error", Message: "invalid email"})
			continue
		}
		emailKey := strings.ToLower(row.Email)
		if prev, ok := seenEmails[emailKey]; ok {
			errors = append(errors, ParseError{
				Row: rowNum, Email: row.Email, Code: "conflict",
				Message: fmt.Sprintf("duplicate email in file (first seen at row %d)", prev),
			})
			continue
		}
		seenEmails[emailKey] = rowNum

		if row.Role == "student" && row.GroupName == "" {
			errors = append(errors, ParseError{Row: rowNum, Email: row.Email, Code: "validation_error", Message: "group_name is required for student"})
			continue
		}

		rows = append(rows, row)
	}

	if len(rows) == 0 && len(errors) == 0 {
		return nil, nil, fmt.Errorf("csv file has no data rows")
	}

	return rows, errors, nil
}

func mapHeader(header []string) (map[string]int, error) {
	normalized := make(map[string]int, len(header))
	for i, h := range header {
		key := strings.ToLower(strings.TrimSpace(h))
		normalized[key] = i
	}

	required := []string{colFIO, colRole, colEmail, colGroupName}
	for _, col := range required {
		if _, ok := normalized[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}
	return normalized, nil
}
