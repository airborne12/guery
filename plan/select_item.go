package plan

import (
	"github.com/xitongsys/guery/config"
	"github.com/xitongsys/guery/gtype"
	"github.com/xitongsys/guery/metadata"
	"github.com/xitongsys/guery/parser"
	"github.com/xitongsys/guery/row"
)

type SelectItemNode struct {
	Expression    *ExpressionNode
	QualifiedName *QualifiedNameNode
	Identifier    *IdentifierNode
	Names         []string
}

func NewSelectItemNode(runtime *config.ConfigRuntime, t parser.ISelectItemContext) *SelectItemNode {
	res := &SelectItemNode{}
	tt := t.(*parser.SelectItemContext)
	if id := tt.Identifier(); id != nil {
		res.Identifier = NewIdentifierNode(runtime, id)
	}

	if ep := tt.Expression(); ep != nil {
		res.Expression = NewExpressionNode(runtime, ep)
		res.Names = []string{res.Expression.Name}

	} else if qn := tt.QualifiedName(); qn != nil {
		res.QualifiedName = NewQulifiedNameNode(runtime, qn)
	}

	if res.Identifier != nil {
		res.Names = []string{tt.Identifier().(*parser.IdentifierContext).GetText()}
	}
	return res
}

func (self *SelectItemNode) GetNames() []string {
	return self.Names
}

func (self *SelectItemNode) GetNamesAndTypes(md *metadata.Metadata) ([]string, []gtype.Type, error) {
	types := []gtype.Type{}
	if self.Expression != nil {
		t, err := self.Expression.GetType(md)
		if err != nil {
			return self.Names, types, err
		}
		types = append(types, t)
		return self.Names, types, nil

	} else {
		return md.GetColumnNames(), md.GetColumnTypes(), nil
	}
}

//get the columns needed in SelectItem
func (self *SelectItemNode) GetColumns(md *metadata.Metadata) ([]string, error) {
	if self.Expression != nil {
		return self.Expression.GetColumns()
	} else { //*
		return md.GetColumnNames(), nil
	}
}

func (self *SelectItemNode) Init(md *metadata.Metadata) error {
	if self.Expression != nil { //some items
		if err := self.Expression.Init(md); err != nil {
			return err
		}

	}
	return nil
}

func (self *SelectItemNode) ExtractAggFunc(res *[]*FuncCallNode) {
	if self.Expression != nil { //some items
		self.Expression.ExtractAggFunc(res)
	} else { //*
	}
}

func (self *SelectItemNode) Result(input *row.RowsGroup) ([]interface{}, error) {
	res := []interface{}{}
	if self.Expression != nil { //some items
		rec, err := self.Expression.Result(input)
		if err != nil {
			return res, err
		}
		res = append(res, rec)

	} else { //*
		for _, v := range input.Vals {
			res = append(res, v)
		}
		self.Names = input.Metadata.GetColumnNames()
	}

	return res, nil
}

func (self *SelectItemNode) IsAggregate() bool {
	if self.Expression != nil {
		return self.Expression.IsAggregate()
	}
	return false
}
