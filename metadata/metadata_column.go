package Metadata

import (
	"fmt"

	"github.com/xitongsys/guery/Type"
)

type ColumnMetadata struct {
	Catalog    string
	Schema     string
	Table      string
	ColumnName string
	ColumnType Type.Type
}

func NewColumnMetadata(t Type.Type, metrics ...string) *ColumnMetadata {
	res := &ColumnMetadata{
		Catalog:    "default",
		Schema:     "default",
		Table:      "default",
		ColumnName: "default",
		ColumnType: t,
	}
	ln := len(metrics)
	if ln >= 1 {
		res.ColumnName = metrics[ln-1]
	}
	if ln >= 2 {
		res.Table = metrics[ln-2]
	}
	if ln >= 3 {
		res.Schema = metrics[ln-3]
	}
	if ln >= 4 {
		res.Catalog = metrics[ln-4]
	}
	return res
}

func (self *ColumnMetadata) Copy() *ColumnMetadata {
	return &ColumnMetadata{
		Catalog:    self.Catalog,
		Schema:     self.Schema,
		Table:      self.Table,
		ColumnName: self.ColumnName,
		ColumnType: self.ColumnType,
	}
}

func (self *ColumnMetadata) GetName() string {
	return fmt.Sprintf("%v.%v.%v.%v", self.Catalog, self.Schema, self.Table, self.ColumnName)
}
