package Plan

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/xitongsys/guery/Config"
	"github.com/xitongsys/guery/Metadata"
	"github.com/xitongsys/guery/Row"
	"github.com/xitongsys/guery/Type"
	"github.com/xitongsys/guery/parser"
)

type ValueExpressionNode struct {
	Name                  string
	PrimaryExpression     *PrimaryExpressionNode
	Operator              *Type.Operator
	ValueExpression       *ValueExpressionNode
	BinaryVauleExpression *BinaryValueExpressionNode
}

func NewValueExpressionNode(runtime *Config.ConfigRuntime, t parser.IValueExpressionContext) *ValueExpressionNode {
	tt := t.(*parser.ValueExpressionContext)
	res := &ValueExpressionNode{}
	children := t.GetChildren()
	switch len(children) {
	case 1: //PrimaryExpression
		res.PrimaryExpression = NewPrimaryExpressionNode(runtime, tt.PrimaryExpression())
		res.Name = res.PrimaryExpression.Name

	case 2: //ValueExpression
		ops := "+"
		if tt.MINUS() != nil {
			res.Operator = Type.NewOperatorFromString("-")
			ops = "-"
		} else {
			res.Operator = Type.NewOperatorFromString("+")
			ops = "+"
		}
		res.ValueExpression = NewValueExpressionNode(runtime, children[1].(parser.IValueExpressionContext))
		res.Name = ops + res.ValueExpression.Name

	case 3: //BinaryValueExpression
		op := Type.NewOperatorFromString(children[1].(*antlr.TerminalNodeImpl).GetText())
		res.BinaryVauleExpression = NewBinaryValueExpressionNode(runtime, tt.GetLeft(), tt.GetRight(), op)
		res.Name = res.BinaryVauleExpression.Name
	}
	return res
}

func (self *ValueExpressionNode) ExtractAggFunc(res *[]*FuncCallNode) {
	if self.PrimaryExpression != nil {
		self.PrimaryExpression.ExtractAggFunc(res)
	} else if self.ValueExpression != nil {
		self.ValueExpression.ExtractAggFunc(res)
	} else if self.BinaryVauleExpression != nil {
		self.BinaryVauleExpression.ExtractAggFunc(res)
	}
}

func (self *ValueExpressionNode) GetType(md *Metadata.Metadata) (Type.Type, error) {
	if self.PrimaryExpression != nil {
		return self.PrimaryExpression.GetType(md)
	} else if self.ValueExpression != nil {
		return self.ValueExpression.GetType(md)
	} else if self.BinaryVauleExpression != nil {
		return self.BinaryVauleExpression.GetType(md)
	}
	return Type.UNKNOWNTYPE, fmt.Errorf("ValueExpressionNode type error")
}

func (self *ValueExpressionNode) GetColumns() ([]string, error) {
	if self.PrimaryExpression != nil {
		return self.PrimaryExpression.GetColumns()
	} else if self.ValueExpression != nil {
		return self.PrimaryExpression.GetColumns()
	} else if self.BinaryVauleExpression != nil {
		return self.BinaryVauleExpression.GetColumns()
	}
	return []string{}, fmt.Errorf("ValueExpression node error")
}

func (self *ValueExpressionNode) Init(md *Metadata.Metadata) error {
	if self.PrimaryExpression != nil {
		return self.PrimaryExpression.Init(md)

	} else if self.ValueExpression != nil {
		return self.ValueExpression.Init(md)

	} else if self.BinaryVauleExpression != nil {
		return self.BinaryVauleExpression.Init(md)
	}
	return fmt.Errorf("wrong ValueExpressionNode")
}

func (self *ValueExpressionNode) Result(input *Row.RowsGroup) (interface{}, error) {
	if self.PrimaryExpression != nil {
		return self.PrimaryExpression.Result(input)

	} else if self.ValueExpression != nil {
		if *self.Operator == Type.MINUS {
			resi, err := self.ValueExpression.Result(input)
			if err != nil {
				return nil, err
			}
			res := resi.([]interface{})
			for i := 0; i < len(res); i++ {
				res[i] = Type.OperatorFunc(-1, res[i], Type.ASTERISK)
			}
			return res, nil
		}
		return self.ValueExpression.Result(input)

	} else if self.BinaryVauleExpression != nil {
		return self.BinaryVauleExpression.Result(input)
	}
	return nil, fmt.Errorf("wrong ValueExpressionNode")
}

func (self *ValueExpressionNode) IsAggregate() bool {
	if self.PrimaryExpression != nil {
		return self.PrimaryExpression.IsAggregate()

	} else if self.ValueExpression != nil {
		return self.ValueExpression.IsAggregate()

	} else if self.BinaryVauleExpression != nil {
		return self.BinaryVauleExpression.IsAggregate()
	}
	return false
}

/////////////////
type BinaryValueExpressionNode struct {
	Name                 string
	LeftValueExpression  *ValueExpressionNode
	RightValueExpression *ValueExpressionNode
	Operator             *Type.Operator
}

func NewBinaryValueExpressionNode(
	runtime *Config.ConfigRuntime,
	left parser.IValueExpressionContext,
	right parser.IValueExpressionContext,
	op *Type.Operator) *BinaryValueExpressionNode {

	res := &BinaryValueExpressionNode{
		LeftValueExpression:  NewValueExpressionNode(runtime, left),
		RightValueExpression: NewValueExpressionNode(runtime, right),
		Operator:             op,
	}
	res.Name = res.LeftValueExpression.Name + "_" + res.RightValueExpression.Name
	return res
}

func (self *BinaryValueExpressionNode) ExtractAggFunc(res *[]*FuncCallNode) {
	self.LeftValueExpression.ExtractAggFunc(res)
	self.RightValueExpression.ExtractAggFunc(res)
}

func (self *BinaryValueExpressionNode) GetType(md *Metadata.Metadata) (Type.Type, error) {
	lt, errL := self.LeftValueExpression.GetType(md)
	if errL != nil {
		return lt, errL
	}
	rt, errR := self.RightValueExpression.GetType(md)
	if errR != nil {
		return rt, errR
	}
	return Type.CheckType(lt, rt, *self.Operator)
}

func (self *BinaryValueExpressionNode) GetColumns() ([]string, error) {
	res, resmp := []string{}, map[string]int{}
	rl, err := self.LeftValueExpression.GetColumns()
	if err != nil {
		return res, err
	}
	rr, err := self.RightValueExpression.GetColumns()
	if err != nil {
		return res, err
	}
	for _, c := range rl {
		resmp[c] = 1
	}
	for _, c := range rr {
		resmp[c] = 1
	}
	for c, _ := range resmp {
		res = append(res, c)
	}
	return res, nil
}

func (self *BinaryValueExpressionNode) Init(md *Metadata.Metadata) error {
	if err := self.LeftValueExpression.Init(md); err != nil {
		return err
	}
	if err := self.RightValueExpression.Init(md); err != nil {
		return err
	}
	return nil
}

func (self *BinaryValueExpressionNode) Result(input *Row.RowsGroup) (interface{}, error) {
	leftValsi, errL := self.LeftValueExpression.Result(input)
	if errL != nil {
		return nil, errL
	}
	rightValsi, errR := self.RightValueExpression.Result(input)
	if errR != nil {
		return nil, errR
	}
	leftVals, rightVals := leftValsi.([]interface{}), rightValsi.([]interface{})
	if len(leftVals) != len(rightVals) {
		return nil, fmt.Errorf("BinaryValueExpressionNode: length not math")
	}
	res := make([]interface{}, len(leftVals))
	for i := 0; i < len(leftVals); i++ {
		res[i] = Type.OperatorFunc(leftVals[i], rightVals[i], *self.Operator)
	}
	return res, nil
}

func (self *BinaryValueExpressionNode) IsAggregate() bool {
	return self.LeftValueExpression.IsAggregate() || self.RightValueExpression.IsAggregate()
}
