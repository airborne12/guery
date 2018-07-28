package Executor

import (
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack"
	"github.com/xitongsys/guery/EPlan"
	"github.com/xitongsys/guery/Logger"
	"github.com/xitongsys/guery/Metadata"
	"github.com/xitongsys/guery/Plan"
	"github.com/xitongsys/guery/Row"
	"github.com/xitongsys/guery/Type"
	"github.com/xitongsys/guery/Util"
	"github.com/xitongsys/guery/pb"
)

func (self *Executor) SetInstructionHashJoin(instruction *pb.Instruction) (err error) {
	var enode EPlan.EPlanHashJoinNode
	if err = msgpack.Unmarshal(instruction.EncodedEPlanNodeBytes, &enode); err != nil {
		return err
	}
	self.Instruction = instruction
	self.EPlanNode = &enode
	self.InputLocations = []*pb.Location{}
	for i, _ := range enode.LeftInputs {
		self.InputLocations = append(self.InputLocations, &enode.LeftInputs[i])
	}
	for i, _ := range enode.RightInputs {
		self.InputLocations = append(self.InputLocations, &enode.RightInputs[i])
	}
	self.OutputLocations = []*pb.Location{&enode.Output}
	return nil
}

func CalHashKey(es []*Plan.ValueExpressionNode, rg *Row.RowsGroup) (string, error) {
	res := ""
	for _, e := range es {
		r, err := e.Result(rg)
		if err != nil {
			return res, err
		}
		res += Type.ToKeyString(r.([]interface{})[0]) + ":"
	}
	return res, nil
}

func (self *Executor) RunHashJoin() (err error) {
	fname := fmt.Sprintf("executor_%v_hashjoin_%v_cpu.pprof", self.Name, time.Now().Format("20060102150405"))
	f, _ := os.Create(fname)
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	defer func() {
		if err != nil {
			self.AddLogInfo(err, pb.LogLevel_ERR)
		}
		self.Clear()
	}()
	writer := self.Writers[0]
	enode := self.EPlanNode.(*EPlan.EPlanHashJoinNode)

	//read md
	mds := make([]*Metadata.Metadata, len(self.Readers))

	for i, reader := range self.Readers {
		mds[i] = &Metadata.Metadata{}
		if err = Util.ReadObject(reader, mds[i]); err != nil {
			return err
		}
	}
	leftNum := len(enode.LeftInputs)
	leftReaders, rightReaders := self.Readers[:leftNum], self.Readers[leftNum:]
	leftMd, rightMd := mds[0], mds[leftNum]

	//write md
	if err = Util.WriteObject(writer, enode.Metadata); err != nil {
		return err
	}

	rbWriter := Row.NewRowsBuffer(enode.Metadata, nil, writer)

	defer func() {
		rbWriter.Flush()
	}()

	//init
	if err := enode.JoinCriteria.Init(enode.Metadata); err != nil {
		return err
	}
	for _, k := range enode.LeftKeys {
		if err := k.Init(leftMd); err != nil {
			return err
		}
	}
	for _, k := range enode.RightKeys {
		if err := k.Init(rightMd); err != nil {
			return err
		}
	}

	//write rows
	rightRg := Row.NewRowsGroup(rightMd)
	rowsMap := make(map[string][]int)

	switch enode.JoinType {
	case Plan.INNERJOIN:
		fallthrough
	case Plan.LEFTJOIN:
		//read right
		var wg sync.WaitGroup
		var mutex sync.Mutex
		for i, _ := range rightReaders {
			wg.Add(1)
			go func(index int) {
				defer func() {
					wg.Done()
				}()

				rightReader := rightReaders[index]
				rightRbReader := Row.NewRowsBuffer(rightMd, rightReader, nil)
				for {
					rg, err := rightRbReader.Read()
					if err == io.EOF {
						err = nil
						break
					}
					if err != nil {
						self.AddLogInfo(err, pb.LogLevel_ERR)
						return
					}
					mutex.Lock()
					rn := rightRg.GetRowsNumber()
					for i := 0; i < rg.GetRowsNumber(); i++ {
						key := rg.GetKeyString(i)
						if _, ok := rowsMap[key]; ok {
							rowsMap[key] = append(rowsMap[key], rn+i)
						} else {
							rowsMap[key] = []int{rn + i}
						}
					}
					rightRg.AppendRowGroupRows(rg)
					mutex.Unlock()
				}
			}(i)
		}
		wg.Wait()

		//read left
		for i, _ := range leftReaders {
			wg.Add(1)
			go func(index int) {
				defer func() {
					wg.Done()
				}()
				leftReader := leftReaders[index]
				leftRbReader := Row.NewRowsBuffer(leftMd, leftReader, nil)
				for {
					rg, err := leftRbReader.Read()
					if err == io.EOF {
						err = nil
						break
					}
					if err != nil {
						self.AddLogInfo(err, pb.LogLevel_ERR)
						return
					}

					for i := 0; i < rg.GetRowsNumber(); i++ {
						row := rg.GetRow(i)
						leftKey := row.GetKeyString()
						joinNum := 0
						if _, ok := rowsMap[leftKey]; ok {
							for _, i := range rowsMap[leftKey] {
								rightRow := rightRg.GetRow(i)
								joinRow := Row.RowPool.Get().(*Row.Row)
								joinRow.Clear()
								joinRow.AppendVals(row.Vals...)
								joinRow.AppendVals(rightRow.Vals...)
								rg := Row.NewRowsGroup(enode.Metadata)
								rg.Write(joinRow)
								if ok, err := enode.JoinCriteria.Result(rg); ok && err == nil {
									if err = rbWriter.WriteRow(joinRow); err != nil {
										self.AddLogInfo(err, pb.LogLevel_ERR)
										return
									}
									joinNum++
								} else if err != nil {
									self.AddLogInfo(err, pb.LogLevel_ERR)
									return
								}
								Row.RowPool.Put(rightRow)
								Row.RowPool.Put(joinRow)
							}
						}

						if enode.JoinType == Plan.LEFTJOIN && joinNum == 0 {
							joinRow := Row.NewRow(row.Vals...)
							joinRow.AppendVals(make([]interface{}, len(mds[1].GetColumnNames()))...)
							if err = rbWriter.WriteRow(joinRow); err != nil {
								self.AddLogInfo(err, pb.LogLevel_ERR)
								return
							}
						}

						Row.RowPool.Put(row)
					}
				}
			}(i)
		}

		wg.Wait()

	case Plan.RIGHTJOIN:

	}

	Logger.Infof("RunJoin finished")
	return err
}
