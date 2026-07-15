package csvparser_test

import (
	"strings"
	"testing"

	"keeneye_practice/app/internal/csvparser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseValidRows(t *testing.T) {
	csvData := `fio,role,email,group_name
Иванов Иван,student,ivan@example.com,ИТ-101
Петров Пётр,teacher,petrov@example.com,
`
	rows, errs, err := csvparser.Parse(strings.NewReader(csvData))
	require.NoError(t, err)
	assert.Empty(t, errs)
	require.Len(t, rows, 2)
	assert.Equal(t, "student", rows[0].Role)
	assert.Equal(t, "ИТ-101", rows[0].GroupName)
}

func TestParseDuplicateEmailInFile(t *testing.T) {
	csvData := `fio,role,email,group_name
A,student,a@x.com,G1
B,student,a@x.com,G1
`
	rows, errs, err := csvparser.Parse(strings.NewReader(csvData))
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	require.Len(t, errs, 1)
	assert.Equal(t, "conflict", errs[0].Code)
}

func TestParseInvalidRole(t *testing.T) {
	csvData := `fio,role,email,group_name
A,admin,a@x.com,G1
`
	_, errs, err := csvparser.Parse(strings.NewReader(csvData))
	require.NoError(t, err)
	require.Len(t, errs, 1)
	assert.Equal(t, "validation_error", errs[0].Code)
}

func TestParseEmptyFile(t *testing.T) {
	_, _, err := csvparser.Parse(strings.NewReader(""))
	require.Error(t, err)
}

func TestParseBOMHeader(t *testing.T) {
	csvData := "\ufefffio,role,email,group_name\nA,student,a@x.com,G1\n"
	rows, errs, err := csvparser.Parse(strings.NewReader(csvData))
	require.NoError(t, err)
	assert.Empty(t, errs)
	require.Len(t, rows, 1)
}

func TestParseStudentWithoutGroup(t *testing.T) {
	csvData := `fio,role,email,group_name
A,student,a@x.com,
`
	rows, errs, err := csvparser.Parse(strings.NewReader(csvData))
	require.NoError(t, err)
	assert.Empty(t, rows)
	require.Len(t, errs, 1)
}
