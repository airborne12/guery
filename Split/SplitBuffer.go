package Split

import (
	"bytes"
	"io"
	"sync"

	"github.com/xitongsys/guery/Metadata"
	"github.com/xitongsys/guery/Type"
	"github.com/xitongsys/guery/Util"
)

type SplitBuffer struct {
	sync.Mutex
	Metadata *Metadata.Metadata
	SP       *Split

	Reader io.Reader
	Writer io.Writer
}

func NewSplitBuffer(md *Metadata.Metadata, reader io.Reader, writer io.Writer) *SplitBuffer {
	res := &SplitBuffer{
		Metadata: md,
		SP:       NewSplit(md),
		Reader:   reader,
		Writer:   writer,
	}
	return res
}

func (self *SplitBuffer) Flush() error {
	err := self.FlushSplit(self.SP)
	self.SP = NewSplit(self.Metadata)
	return err
}

func (self *SplitBuffer) FlushSplit(sp *Split) error {
	colNum := sp.GetColumnNumber()
	//for 0 cols, just need send the number of rows
	if colNum <= 0 {
		buf := Type.EncodeValues([]interface{}{int64(sp.GetRowsNumber())}, Type.INT64)
		return Util.WriteMessage(self.Writer, buf)
	}

	//for several cols
	for i := 0; i < colNum; i++ {
		col := sp.ValueFlags[i]
		buf := Type.EncodeBool(col)
		if err := Util.WriteMessage(self.Writer, buf); err != nil {
			return err
		}

		col = sp.Values[i]
		t, err := sp.Metadata.GetTypeByIndex(i)
		if err != nil {
			return err
		}
		buf = Type.EncodeValues(col, t)
		if err := Util.WriteMessage(self.Writer, buf); err != nil {
			return err
		}
	}

	colNum = sp.GetKeyColumnNumber()
	for i := 0; i < colNum; i++ {
		col := sp.KeyFlags[i]
		buf := Type.EncodeBool(col)
		if err := Util.WriteMessage(self.Writer, buf); err != nil {
			return err
		}

		col = sp.Keys[i]
		t, err := sp.Metadata.GetKeyTypeByIndex(i)
		if err != nil {
			return err
		}
		buf = Type.EncodeValues(col, t)
		if err := Util.WriteMessage(self.Writer, buf); err != nil {
			return err
		}
	}
	return nil
}

func (self *SplitBuffer) ReadSplit() (*Split, error) {
	sp := Split.NewSplit(self.Metadata)
	colNum := self.Metadata.GetColumnNumber()
	//for 0 cols
	if colNum <= 0 {
		buf, err := Util.ReadMessage(self.Reader)
		if err != nil {
			return sp, err
		}
		vals, err := Type.DecodeINT64(bytes.NewReader(buf))
		if err != nil || len(vals) <= 0 {
			return sp, err
		}
		sp.RowsNumber = int(vals[0].(int64))
	}

	//for cols
	for i := 0; i < colNum; i++ {
		buf, err := Util.ReadMessage(self.Reader)
		if err != nil {
			return sp, err
		}

		sp.ValueFlags[i], err = Type.DecodeBOOL(bytes.NewReader(buf))
		if err != nil {
			return sp, err
		}

		buf, err = Util.ReadMessage(self.Reader)
		t, err := self.Metadata.GetTypeByIndex(i)
		if err != nil {
			return sp, err
		}
		values, err := Type.DecodeValue(bytes.NewReader(buf), t)
		if err != nil {
			return sp, err
		}

		//log.Println("=======", buf, values, self.ValueNilFlags)

		sp.Values[i] = make([]interface{}, len(self.ValueNilFlags[i]))
		k := 0
		for j := 0; j < len(sp.ValueFlags[i]) && k < len(values); j++ {
			if sp.ValueFlags[i][j].(bool) {
				sp.Values[i][j] = values[k]
				k++
			} else {
				sp.Values[i][j] = nil
			}
		}
		self.RowsNumber = len(sp.ValueFlags[i])
	}

	keyNum := self.MD.GetKeyNumber()
	for i := 0; i < keyNum; i++ {
		buf, err := Util.ReadMessage(self.Reader)
		if err != nil {
			return sp, err
		}
		self.KeyFlags[i], err = Type.DecodeBOOL(bytes.NewReader(buf))
		if err != nil {
			return sp, err
		}

		buf, err = Util.ReadMessage(self.Reader)
		t, err := self.MD.GetKeyTypeByIndex(i)
		if err != nil {
			return sp, err
		}
		keys, err := Type.DecodeValue(bytes.NewReader(buf), t)
		if err != nil {
			return sp, err
		}

		sp.Keys[i] = make([]interface{}, len(sp.KeyFlags[i]))
		k := 0
		for j := 0; j < len(sp.KeyFlags[i]) && k < len(keys); j++ {
			if sp.KeyFlags[i][j].(bool) {
				sp.Keys[i][j] = keys[k]
				k++
			} else {
				sp.Keys[i][j] = nil
			}
		}
	}
	return sp, nil
}

func (self *SplitBuffer) Write(sp *Split.Split, index ...int) error {
	self.Lock()
	defer self.Unlock()
	self.SP.Append(sp, index...)
	if self.SP.GetRowsNumber() >= Split.MAX_SPLIT_SIZE {
		return self.Flush()
	}
	return nil
}

func (self *SplitBuffer) WriteValues(vals []interface{}) error {
	self.Lock()
	defer self.Unlock()
	self.SP.AppendValues(vals)
	if self.SP.GetRowsNumber() >= Split.MAX_SPLIT_SIZE {
		return self.Flush()
	}
	return nil
}
