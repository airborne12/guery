package Plan

import (
	"fmt"

	"github.com/xitongsys/guery/Common"
	"github.com/xitongsys/guery/Util"
	"github.com/xitongsys/guery/parser"
)

type FuncCallNode struct {
	FuncName    string
	Expressions []*ExpressionNode
}

func NewFuncCallNode(name string, expressions []parser.IExpressionContext) *FuncCallNode {
	res := &FuncCallNode{
		FuncName:    name,
		Expressions: make([]*ExpressionNode, len(expressions)),
	}
	for i := 0; i < len(expressions); i++ {
		res.Expressions[i] = NewExpressionNode(expressions[i])
	}
	return res
}

func (self *FuncCallNode) Result(input *Util.RowsBuffer) (interface{}, error) {
	switch self.FuncName {
	case "SUM":
		return SUM(input, self.Expressions[0])
	case "MIN":
		return MIN(input, self.Expressions[0])
	case "MAX":
		return MAX(input, self.Expressions[0])
	case "ABS":
		return ABS(input)
	}
	return nil, fmt.Errorf("Unkown function %v", self.FuncName)
}

func (self *FuncCallNode) IsAggregate() bool {
	switch self.FuncName {
	case "SUM":
		return true
	case "MIN":
		return true
	case "MAX":
		return true
	case "ABS":
		return false
	}
	return false
}

func SUM(input *Util.RowsBuffer, t *ExpressionNode) (interface{}, error) {
	var (
		err      error
		res, tmp interface{}
		rb       *Util.RowsBuffer
		row      *Util.Row
	)

	for {
		row, err = input.Read()
		if err != nil {
			break
		}
		rb = Util.NewRowsBuffer(input.Metadata)
		rb.Write(row)
		tmp, err = t.Result(rb)
		if err != nil {
			break
		}

		if res == nil {
			res = tmp
		} else {
			res = Common.Arithmetic(res, tmp, Common.PLUS)
		}
	}
	return res, err
}

func MIN(input *Util.RowsBuffer, t *ExpressionNode) (interface{}, error) {
	var (
		err      error
		res, tmp interface{}
		rb       *Util.RowsBuffer
		row      *Util.Row
	)

	for {
		row, err = input.Read()
		if err != nil {
			break
		}
		rb = Util.NewRowsBuffer(input.Metadata)
		rb.Write(row)
		tmp, err = t.Result(rb)
		if err != nil {
			break
		}

		if res == nil {
			res = tmp
		} else {
			if Common.Cmp(res, tmp) > 0 {
				res = tmp
			}
		}
	}
	return res, err
}

func MAX(input *Util.RowsBuffer, t *ExpressionNode) (interface{}, error) {
	var (
		err      error
		res, tmp interface{}
		rb       *Util.RowsBuffer
		row      *Util.Row
	)

	for {
		row, err = input.Read()
		if err != nil {
			break
		}
		rb = Util.NewRowsBuffer(input.Metadata)
		rb.Write(row)
		tmp, err = t.Result(rb)
		if err != nil {
			break
		}

		if res == nil {
			res = tmp
		} else {
			if Common.Cmp(res, tmp) < 0 {
				res = tmp
			}
		}
	}
	return res, err
}

func ABS(input *Util.RowsBuffer) (interface{}, error) {
	return nil, nil
}