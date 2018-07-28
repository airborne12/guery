package Plan

import (
	"fmt"

	"github.com/xitongsys/guery/Config"
	"github.com/xitongsys/guery/Metadata"
	"github.com/xitongsys/guery/Row"
	"github.com/xitongsys/guery/Type"
	"github.com/xitongsys/guery/parser"
)

type BooleanExpressionNode struct {
	Name                    string
	Predicated              *PredicatedNode
	NotBooleanExpression    *NotBooleanExpressionNode
	BinaryBooleanExpression *BinaryBooleanExpressionNode
}

func NewBooleanExpressionNode(runtime *Config.ConfigRuntime, t parser.IBooleanExpressionContext) *BooleanExpressionNode {
	tt := t.(*parser.BooleanExpressionContext)
	res := &BooleanExpressionNode{}
	children := tt.GetChildren()
	switch len(children) {
	case 1: //Predicated
		res.Predicated = NewPredicatedNode(runtime, tt.Predicated())
		res.Name = res.Predicated.Name

	case 2: //NOT
		res.NotBooleanExpression = NewNotBooleanExpressionNode(runtime, tt.BooleanExpression(0))
		res.Name = res.NotBooleanExpression.Name

	case 3: //Binary
		var o Type.Operator
		if tt.AND() != nil {
			o = Type.AND
		} else if tt.OR() != nil {
			o = Type.OR
		}
		res.BinaryBooleanExpression = NewBinaryBooleanExpressionNode(runtime, tt.GetLeft(), tt.GetRight(), o)
		res.Name = res.BinaryBooleanExpression.Name

	}
	return res
}

func (self *BooleanExpressionNode) ExtractAggFunc(res *[]*FuncCallNode) {
	if self.Predicated != nil {
		self.Predicated.ExtractAggFunc(res)
	} else if self.NotBooleanExpression != nil {
		self.NotBooleanExpression.ExtractAggFunc(res)
	} else if self.BinaryBooleanExpression != nil {
		self.BinaryBooleanExpression.ExtractAggFunc(res)
	}
}

func (self *BooleanExpressionNode) GetType(md *Metadata.Metadata) (Type.Type, error) {
	if self.Predicated != nil {
		return self.Predicated.GetType(md)
	} else if self.NotBooleanExpression != nil {
		return self.NotBooleanExpression.GetType(md)
	} else if self.BinaryBooleanExpression != nil {
		return self.BinaryBooleanExpression.GetType(md)
	}
	return Type.UNKNOWNTYPE, fmt.Errorf("GetType: wrong BooleanExpressionNode")
}

func (self *BooleanExpressionNode) GetColumns() ([]string, error) {
	if self.Predicated != nil {
		return self.Predicated.GetColumns()
	} else if self.NotBooleanExpression != nil {
		return self.NotBooleanExpression.GetColumns()
	} else if self.BinaryBooleanExpression != nil {
		return self.BinaryBooleanExpression.GetColumns()
	}
	return nil, fmt.Errorf("GetColumns: wrong BooleanExpressionNode")
}

func (self *BooleanExpressionNode) Init(md *Metadata.Metadata) error {
	if self.Predicated != nil {
		return self.Predicated.Init(md)
	} else if self.NotBooleanExpression != nil {
		return self.NotBooleanExpression.Init(md)
	} else if self.BinaryBooleanExpression != nil {
		return self.BinaryBooleanExpression.Init(md)
	}
	return fmt.Errorf("wrong BooleanExpressionNode")
}

func (self *BooleanExpressionNode) Result(input *Row.RowsGroup) (interface{}, error) {
	if self.Predicated != nil {
		return self.Predicated.Result(input)
	} else if self.NotBooleanExpression != nil {
		return self.NotBooleanExpression.Result(input)
	} else if self.BinaryBooleanExpression != nil {
		return self.BinaryBooleanExpression.Result(input)
	}
	return nil, fmt.Errorf("wrong BooleanExpressionNode")
}

func (self *BooleanExpressionNode) IsAggregate() bool {
	if self.Predicated != nil {
		return self.Predicated.IsAggregate()
	} else if self.NotBooleanExpression != nil {
		return self.NotBooleanExpression.IsAggregate()
	} else if self.BinaryBooleanExpression != nil {
		return self.BinaryBooleanExpression.IsAggregate()
	}
	return false
}

////////////////////////
type NotBooleanExpressionNode struct {
	Name              string
	BooleanExpression *BooleanExpressionNode
}

func NewNotBooleanExpressionNode(runtime *Config.ConfigRuntime, t parser.IBooleanExpressionContext) *NotBooleanExpressionNode {
	res := &NotBooleanExpressionNode{
		BooleanExpression: NewBooleanExpressionNode(runtime, t),
	}
	res.Name = "NOT_" + res.BooleanExpression.Name
	return res
}

func (self *NotBooleanExpressionNode) ExtractAggFunc(res *[]*FuncCallNode) {
	self.BooleanExpression.ExtractAggFunc(res)
}

func (self *NotBooleanExpressionNode) GetType(md *Metadata.Metadata) (Type.Type, error) {
	t, err := self.BooleanExpression.GetType(md)
	if err != nil {
		return t, err
	}
	if t != Type.BOOL {
		return t, fmt.Errorf("expression type error")
	}
	return t, nil
}

func (self *NotBooleanExpressionNode) GetColumns() ([]string, error) {
	return self.BooleanExpression.GetColumns()
}

func (self *NotBooleanExpressionNode) Init(md *Metadata.Metadata) error {
	return self.BooleanExpression.Init(md)
}

func (self *NotBooleanExpressionNode) Result(input *Row.RowsGroup) (interface{}, error) {
	resi, err := self.BooleanExpression.Result(input)
	if err != nil {
		return false, err
	}

	res := resi.([]interface{})
	for i := 0; i < len(res); i++ {
		res[i] = !(res[i].(bool))
	}
	return res, nil
}

func (self *NotBooleanExpressionNode) IsAggregate() bool {
	return self.BooleanExpression.IsAggregate()
}

////////////////////////
type BinaryBooleanExpressionNode struct {
	Name                   string
	LeftBooleanExpression  *BooleanExpressionNode
	RightBooleanExpression *BooleanExpressionNode
	Operator               *Type.Operator
}

func NewBinaryBooleanExpressionNode(
	runtime *Config.ConfigRuntime,
	left parser.IBooleanExpressionContext,
	right parser.IBooleanExpressionContext,
	op Type.Operator) *BinaryBooleanExpressionNode {

	res := &BinaryBooleanExpressionNode{
		LeftBooleanExpression:  NewBooleanExpressionNode(runtime, left),
		RightBooleanExpression: NewBooleanExpressionNode(runtime, right),
		Operator:               &op,
	}
	res.Name = res.LeftBooleanExpression.Name + "_" + res.RightBooleanExpression.Name
	return res
}

func (self *BinaryBooleanExpressionNode) ExtractAggFunc(res *[]*FuncCallNode) {
	self.LeftBooleanExpression.ExtractAggFunc(res)
	self.RightBooleanExpression.ExtractAggFunc(res)
}

func (self *BinaryBooleanExpressionNode) GetType(md *Metadata.Metadata) (Type.Type, error) {
	lt, err1 := self.LeftBooleanExpression.GetType(md)
	if err1 != nil {
		return Type.UNKNOWNTYPE, err1
	}
	if lt != Type.BOOL {
		return lt, fmt.Errorf("expression type error")
	}
	rt, err2 := self.RightBooleanExpression.GetType(md)
	if err2 != nil {
		return Type.UNKNOWNTYPE, err2
	}
	if rt != Type.BOOL {
		return rt, fmt.Errorf("expression type error")
	}

	return Type.BOOL, nil
}

func (self *BinaryBooleanExpressionNode) GetColumns() ([]string, error) {
	resmp := make(map[string]int)
	res := []string{}
	rl, errl := self.LeftBooleanExpression.GetColumns()
	if errl != nil {
		return res, errl
	}
	rr, errr := self.RightBooleanExpression.GetColumns()
	if errr != nil {
		return res, errr
	}
	for _, c := range rl {
		resmp[c] = 1
	}
	for _, c := range rr {
		resmp[c] = 1
	}
	for key, _ := range resmp {
		res = append(res, key)
	}
	return res, nil
}

func (self *BinaryBooleanExpressionNode) Init(md *Metadata.Metadata) error {
	if err := self.LeftBooleanExpression.Init(md); err != nil {
		return err
	}
	if err := self.RightBooleanExpression.Init(md); err != nil {
		return err
	}
	return nil
}

func (self *BinaryBooleanExpressionNode) Result(input *Row.RowsGroup) (interface{}, error) {
	leftResi, err := self.LeftBooleanExpression.Result(input)
	if err != nil {
		return nil, err
	}
	rightResi, err := self.RightBooleanExpression.Result(input)
	if err != nil {
		return nil, err
	}

	leftRes, rightRes := leftResi.([]interface{}), rightResi.([]interface{})
	for i := 0; i < input.GetRowsNumber(); i++ {
		if *self.Operator == Type.AND {
			leftRes[i] = leftRes[i].(bool) && rightRes[i].(bool)
		} else if *self.Operator == Type.OR {
			leftRes[i] = leftRes[i].(bool) || rightRes[i].(bool)
		}
	}
	return leftRes, nil
}

func (self *BinaryBooleanExpressionNode) IsAggregate() bool {
	return self.LeftBooleanExpression.IsAggregate() || self.RightBooleanExpression.IsAggregate()
}
