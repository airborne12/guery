package Plan

import (
	"fmt"
	"strings"

	"github.com/xitongsys/guery/Metadata"
	"github.com/xitongsys/guery/Row"
	"github.com/xitongsys/guery/Type"
)

func NewLengthFunc() *GueryFunc {
	res := &GueryFunc{
		Name: "LENGTH",
		IsAggregate: func(es []*ExpressionNode) bool {
			if len(es) < 1 {
				return false
			}
			return es[0].IsAggregate()
		},

		GetType: func(md *Metadata.Metadata, es []*ExpressionNode) (Type.Type, error) {
			return Type.INT64, nil
		},

		Result: func(input *Row.RowsGroup, Expressions []*ExpressionNode) (interface{}, error) {
			if len(Expressions) < 1 {
				return nil, fmt.Errorf("not enough parameters in LENGTH")
			}
			var (
				err error
				tmp interface{}
				t   *ExpressionNode = Expressions[0]
			)

			input.Reset()
			if tmp, err = t.Result(input); err != nil {
				return nil, err
			}

			switch Type.TypeOf(tmp) {
			case Type.STRING:
				return int64(len(tmp.(string))), nil

			default:
				return nil, fmt.Errorf("type cann't use LENGTH function")
			}
		},
	}
	return res
}

func NewLowerFunc() *GueryFunc {
	res := &GueryFunc{
		Name: "LOWER",
		IsAggregate: func(es []*ExpressionNode) bool {
			if len(es) < 1 {
				return false
			}
			return es[0].IsAggregate()
		},

		GetType: func(md *Metadata.Metadata, es []*ExpressionNode) (Type.Type, error) {
			return Type.STRING, nil
		},

		Result: func(input *Row.RowsGroup, Expressions []*ExpressionNode) (interface{}, error) {
			if len(Expressions) < 1 {
				return nil, fmt.Errorf("not enough parameters in LOWER")
			}
			var (
				err error
				tmp interface{}
				t   *ExpressionNode = Expressions[0]
			)

			input.Reset()
			if tmp, err = t.Result(input); err != nil {
				return nil, err
			}

			switch Type.TypeOf(tmp) {
			case Type.STRING:
				return strings.ToLower(tmp.(string)), nil

			default:
				return nil, fmt.Errorf("type cann't use LOWER function")
			}
		},
	}
	return res
}

func NewUpperFunc() *GueryFunc {
	res := &GueryFunc{
		Name: "UPPER",
		IsAggregate: func(es []*ExpressionNode) bool {
			if len(es) < 1 {
				return false
			}
			return es[0].IsAggregate()
		},

		GetType: func(md *Metadata.Metadata, es []*ExpressionNode) (Type.Type, error) {
			return Type.STRING, nil
		},

		Result: func(input *Row.RowsGroup, Expressions []*ExpressionNode) (interface{}, error) {
			if len(Expressions) < 1 {
				return nil, fmt.Errorf("not enough parameters in UPPER")
			}
			var (
				err error
				tmp interface{}
				t   *ExpressionNode = Expressions[0]
			)

			input.Reset()
			if tmp, err = t.Result(input); err != nil {
				return nil, err
			}

			switch Type.TypeOf(tmp) {
			case Type.STRING:
				return strings.ToUpper(tmp.(string)), nil

			default:
				return nil, fmt.Errorf("type cann't use UPPER function")
			}
		},
	}
	return res
}

func NewReverseFunc() *GueryFunc {
	res := &GueryFunc{
		Name: "REVERSE",
		IsAggregate: func(es []*ExpressionNode) bool {
			if len(es) < 1 {
				return false
			}
			return es[0].IsAggregate()
		},

		GetType: func(md *Metadata.Metadata, es []*ExpressionNode) (Type.Type, error) {
			return Type.STRING, nil
		},

		Result: func(input *Row.RowsGroup, Expressions []*ExpressionNode) (interface{}, error) {
			if len(Expressions) < 1 {
				return nil, fmt.Errorf("not enough parameters in REVERSE")
			}
			var (
				err error
				tmp interface{}
				t   *ExpressionNode = Expressions[0]
			)

			input.Reset()
			if tmp, err = t.Result(input); err != nil {
				return nil, err
			}

			switch Type.TypeOf(tmp) {
			case Type.STRING:
				bs := []byte(tmp.(string))
				bd := make([]byte, len(bs))
				for i := 0; i < len(bs); i++ {
					bd[len(bs)-i-1] = bs[i]
				}
				return string(bd), nil

			default:
				return nil, fmt.Errorf("type cann't use REVERSE function")
			}
		},
	}
	return res
}

func NewConcatFunc() *GueryFunc {
	res := &GueryFunc{
		Name: "CONCAT",
		IsAggregate: func(es []*ExpressionNode) bool {
			if len(es) < 2 {
				return false
			}
			return es[0].IsAggregate() || es[1].IsAggregate()
		},

		GetType: func(md *Metadata.Metadata, es []*ExpressionNode) (Type.Type, error) {
			return Type.STRING, nil
		},

		Result: func(input *Row.RowsGroup, Expressions []*ExpressionNode) (interface{}, error) {
			if len(Expressions) < 2 {
				return nil, fmt.Errorf("not enough parameters in CONCAT")
			}
			var (
				err        error
				tmp1, tmp2 interface{}
				t1         *ExpressionNode = Expressions[0]
				t2         *ExpressionNode = Expressions[1]
			)

			input.Reset()
			if tmp1, err = t1.Result(input); err != nil {
				return nil, err
			}

			input.Reset()
			if tmp2, err = t2.Result(input); err != nil {
				return nil, err
			}

			if Type.TypeOf(tmp1) != Type.STRING || Type.TypeOf(tmp2) != Type.STRING {
				return nil, fmt.Errorf("type error in CONCAT")
			}

			return tmp1.(string) + tmp2.(string), nil
		},
	}
	return res
}

func NewSubstrFunc() *GueryFunc {
	res := &GueryFunc{
		Name: "SUBSTR",
		IsAggregate: func(es []*ExpressionNode) bool {
			if len(es) < 3 {
				return false
			}
			return es[0].IsAggregate() || es[1].IsAggregate() || es[2].IsAggregate()
		},

		GetType: func(md *Metadata.Metadata, es []*ExpressionNode) (Type.Type, error) {
			return Type.STRING, nil
		},

		Result: func(input *Row.RowsGroup, Expressions []*ExpressionNode) (interface{}, error) {
			if len(Expressions) < 3 {
				return nil, fmt.Errorf("not enough parameters in SUBSTR")
			}
			var (
				err              error
				tmp1, tmp2, tmp3 interface{}
				t1               *ExpressionNode = Expressions[0]
				t2               *ExpressionNode = Expressions[1]
				t3               *ExpressionNode = Expressions[2]
			)

			input.Reset()
			if tmp1, err = t1.Result(input); err != nil {
				return nil, err
			}

			input.Reset()
			if tmp2, err = t2.Result(input); err != nil {
				return nil, err
			}

			input.Reset()
			if tmp3, err = t3.Result(input); err != nil {
				return nil, err
			}

			bgn, end := Type.ToInt64(tmp2), Type.ToInt64(tmp3)

			if Type.TypeOf(tmp1) != Type.STRING {
				return nil, fmt.Errorf("type error in SUBSTR")
			}
			if bgn < 0 || end < 0 {
				return nil, fmt.Errorf("index out of range in SUBSTR")
			}
			return tmp1.(string)[bgn:end], nil

		},
	}
	return res
}

func NewReplaceFunc() *GueryFunc {
	res := &GueryFunc{
		Name: "REPLACE",
		IsAggregate: func(es []*ExpressionNode) bool {
			if len(es) < 3 {
				return false
			}
			return es[0].IsAggregate() || es[1].IsAggregate() || es[2].IsAggregate()
		},

		GetType: func(md *Metadata.Metadata, es []*ExpressionNode) (Type.Type, error) {
			return Type.STRING, nil
		},

		Result: func(input *Row.RowsGroup, Expressions []*ExpressionNode) (interface{}, error) {
			if len(Expressions) < 3 {
				return nil, fmt.Errorf("not enough parameters in REPLACE")
			}
			var (
				err              error
				tmp1, tmp2, tmp3 interface{}
				t1               *ExpressionNode = Expressions[0]
				t2               *ExpressionNode = Expressions[1]
				t3               *ExpressionNode = Expressions[2]
			)

			input.Reset()
			if tmp1, err = t1.Result(input); err != nil {
				return nil, err
			}

			input.Reset()
			if tmp2, err = t2.Result(input); err != nil {
				return nil, err
			}

			input.Reset()
			if tmp3, err = t3.Result(input); err != nil {
				return nil, err
			}

			if Type.TypeOf(tmp1) != Type.STRING || Type.TypeOf(tmp2) != Type.STRING || Type.TypeOf(tmp3) != Type.STRING {
				return nil, fmt.Errorf("type error in REPLACE")
			}

			return strings.Replace(tmp1.(string), tmp2.(string), tmp3.(string), -1), nil

		},
	}
	return res
}
