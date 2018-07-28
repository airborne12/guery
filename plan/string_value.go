package plan

import (
	"github.com/xitongsys/guery/config"
	"github.com/xitongsys/guery/gtype"
	"github.com/xitongsys/guery/metadata"
	"github.com/xitongsys/guery/parser"
	"github.com/xitongsys/guery/row"
)

type StringValueNode struct {
	Name string
	Str  string
}

func NewStringValueNode(runtime *Config.ConfigRuntime, t parser.IStringValueContext) *StringValueNode {
	s := t.GetText()
	ls := len(s)
	return &StringValueNode{
		Str:  s[1 : ls-1],
		Name: s[1 : ls-1],
	}
}

func (self *StringValueNode) Init(md *Metadata.Metadata) error {
	return nil
}

func (self *StringValueNode) Result(input *Row.RowsGroup) (interface{}, error) {
	rn := input.GetRowsNumber()
	res := make([]interface{}, rn)
	for i := 0; i < rn; i++ {
		res[i] = self.Str
	}
	return res, nil
}

func (self *StringValueNode) GetType(md *Metadata.Metadata) (Type.Type, error) {
	return Type.STRING, nil
}
