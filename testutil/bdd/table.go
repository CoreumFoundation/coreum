package bdd

import "github.com/cucumber/godog"

type Table struct {
	Rows      map[string][]string
	RowsCount int
}

func ParseTable(input *godog.Table) Table {
	table := Table{
		Rows:      make(map[string][]string),
		RowsCount: len(input.Rows) - 1,
	}
	if len(input.Rows) < 1 {
		return table
	}
	for _, cell := range input.Rows[0].Cells {
		table.Rows[cell.Value] = make([]string, 0, len(input.Rows))
	}
	for i, row := range input.Rows {
		if i == 0 {
			continue
		}
		for j, cell := range row.Cells {
			table.Rows[input.Rows[0].Cells[j].Value] = append(table.Rows[input.Rows[0].Cells[j].Value], cell.Value)
		}
	}
	return table
}
