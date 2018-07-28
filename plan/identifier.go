package Plan

import (
	"fmt"

	"github.com/xitongsys/guery/Config"
	"github.com/xitongsys/guery/Metadata"
	"github.com/xitongsys/guery/Row"
	"github.com/xitongsys/guery/Type"
	"github.com/xitongsys/guery/parser"
)

type IdentifierNode struct {
	Str         *string
	Digit       *int
	NonReserved *string
}

func NewIdentifierNode(runtime *Config.ConfigRuntime, t parser.IIdentifierContext) *IdentifierNode {
	tt := t.(*parser.IdentifierContext)
	res := &IdentifierNode{}
	var (
		str string
		dig int
	)

	if id := tt.IDENTIFIER(); id != nil {
		str = id.GetText()
		res.Str = &str

	} else if qid := tt.QUOTED_IDENTIFIER(); qid != nil {
		str = qid.GetText()
		ln := len(str)
		str = str[1 : ln-1]
		res.Str = &str

	} else if nr := tt.NonReserved(); nr != nil {
		str = nr.GetText()
		res.NonReserved = &str

	} else if did := tt.DIGIT_IDENTIFIER(); did != nil {
		str = did.GetText()
		fmt.Sscanf(str, "%d", &dig)
		res.Digit = &dig
	}
	return res
}

func (self *IdentifierNode) GetType(md *Metadata.Metadata) (Type.Type, error) {
	if self.Digit != nil {
		index := *self.Digit
		return md.GetTypeByIndex(int(index))

	} else if self.Str != nil {
		return md.GetTypeByName(*self.Str)
	}
	return Type.UNKNOWNTYPE, fmt.Errorf("Wrong IdentifierNode")
}

func (self *IdentifierNode) GetColumns() ([]string, error) {
	if self.Digit != nil {
		return []string{}, nil
	} else if self.Str != nil {
		return []string{*self.Str}, nil
	}
	return []string{}, fmt.Errorf("wrong identifierNode")
}

func (self *IdentifierNode) Init(md *Metadata.Metadata) error {
	if self.Str != nil {
		index, err := md.GetIndexByName(*self.Str)
		if err != nil {
			return err
		}
		self.Digit = &index
	}
	return nil
}

func (self *IdentifierNode) Result(input *Row.RowsGroup) (interface{}, error) {
	rn := input.GetRowsNumber()
	index := 0

	if self.Digit != nil {
		if *self.Digit >= input.GetColumnsNumber() {
			return nil, fmt.Errorf("index out of range")
		}
		index = *self.Digit

	} else if self.Str != nil {
		index = input.GetIndex(*self.Str)
		if index >= input.GetColumnsNumber() {
			return nil, fmt.Errorf("index out of range")
		}
	} else {
		return nil, fmt.Errorf("wrong IdentifierNode")
	}

	res := make([]interface{}, rn)
	for i := 0; i < rn; i++ {
		res[i] = input.Vals[index][i]
	}
	return res, nil
}

func (self *IdentifierNode) GetText() string {
	if self.Str != nil {
		return *self.Str
	} else if self.Digit != nil {
		return fmt.Sprintf("%d", *self.Digit)
	}
	return ""
}
