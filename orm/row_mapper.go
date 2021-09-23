package orm

import (
	"github.com/jmoiron/sqlx"
)

type RowScan interface {
	TotalRows() int
	Columns() []string
	RowsValues() []*IterValue
	SliceScan(row int) []interface{}
	PaginateTotal() int
	Error() error
}

type RowValue interface {
	CurrentRow() int
	Columns() []string
	Values() []interface{}
}

type rowScan struct {
	// result of rows
	totalRows  int
	columns    []string
	rowsValues []*IterValue
	rowsErr    error

	// สำหรับ นับ เพื่อทำ paginate จะหาค่าเฉพาะ key total_row
	paginateTotal int
}

type IterValue struct {
	currentRow int
	columns    []string
	values     []interface{}
}

func NewRowsScan(rows *sqlx.Rows) (RowScan, error) {
	var m = &rowScan{
		totalRows:     0,
		columns:       make([]string, 0),
		rowsValues:    make([]*IterValue, 0),
		rowsErr:       rows.Err(),
		paginateTotal: 0,
	}

	if m.rowsErr != nil {
		return nil, m.rowsErr
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	m.columns = columns

	for rows.Next() {
		m.totalRows++
		values, err := rows.SliceScan()
		if err != nil {
			return nil, err
		}
		if len(values) > 0 {
			if columns != nil && values != nil {
				if len(columns) > 0 && len(values) > 0 {
					for index, column := range columns {
						if column == "total_row" {
							total := int(values[index].(int64))
							m.paginateTotal = total
						}
					}
				}
			}

			m.rowsValues = append(m.rowsValues, &IterValue{
				currentRow: m.totalRows,
				columns:    columns,
				values:     values,
			})
		}
	}

	return m, nil
}

func (r rowScan) TotalRows() int {
	return r.totalRows
}

func (r rowScan) Columns() []string {
	return r.columns
}

func (r rowScan) RowsValues() []*IterValue {
	return r.rowsValues
}

func (r rowScan) PaginateTotal() int {
	return r.paginateTotal
}

func (r rowScan) SliceScan(row int) []interface{} {
	if len(r.rowsValues) > 0 {
		for _, item := range r.rowsValues {
			if item.currentRow == row {
				return item.values
			}
		}
	}
	return make([]interface{}, 0)
}

func (r rowScan) Error() error {
	return r.rowsErr
}

func (r IterValue) CurrentRow() int {
	return r.currentRow
}

func (r IterValue) Columns() []string {
	return r.columns
}

func (r IterValue) Values() []interface{} {
	return r.values
}
