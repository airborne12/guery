package plan

import (
	"github.com/xitongsys/guery/config"
	"github.com/xitongsys/guery/gtype"
	"github.com/xitongsys/guery/metadata"
	"github.com/xitongsys/guery/parser"
	"github.com/xitongsys/guery/row"
)

type SortItemNode struct {
	Expression *ExpressionNode
	OrderType  Type.OrderType
}

func NewSortItemNode(runtime *Config.ConfigRuntime, t parser.ISortItemContext) *SortItemNode {
	tt := t.(*parser.SortItemContext)
	res := &SortItemNode{
		Expression: NewExpressionNode(runtime, tt.Expression()),
		OrderType:  Type.ASC,
	}

	if ot := tt.GetOrdering(); ot != nil {
		if ot.GetText() != "ASC" {
			res.OrderType = Type.DESC
		}
	}

	return res
}

func (self *SortItemNode) GetColumns() ([]string, error) {
	return self.Expression.GetColumns()
}

func (self *SortItemNode) Init(md *Metadata.Metadata) error {
	return self.Expression.Init(md)
}

func (self *SortItemNode) Result(input *Row.RowsGroup) (interface{}, error) {
	return self.Expression.Result(input)
}

func (self *SortItemNode) IsAggregate() bool {
	return self.Expression.IsAggregate()
}

func (self *SortItemNode) GetType(md *Metadata.Metadata) (Type.Type, error) {
	return self.Expression.GetType(md)
}
