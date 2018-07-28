package Plan

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/xitongsys/guery/Config"
	"github.com/xitongsys/guery/Metadata"
)

type PlanLimitNode struct {
	Input       PlanNode
	Output      PlanNode
	Metadata    *Metadata.Metadata
	LimitNumber *int64
}

func NewPlanLimitNode(runtime *Config.ConfigRuntime, input PlanNode, t antlr.TerminalNode) *PlanLimitNode {
	res := &PlanLimitNode{
		Input:    input,
		Metadata: Metadata.NewMetadata(),
	}
	if ns := t.GetText(); ns != "ALL" {
		var num int64
		fmt.Sscanf(ns, "%d", &num)
		res.LimitNumber = &num
	}
	return res
}

func (self *PlanLimitNode) GetInputs() []PlanNode {
	return []PlanNode{self.Input}
}

func (self *PlanLimitNode) SetInputs(inputs []PlanNode) {
	self.Input = inputs[0]
}

func (self *PlanLimitNode) GetOutput() PlanNode {
	return self.Output
}

func (self *PlanLimitNode) SetOutput(output PlanNode) {
	self.Output = output
}

func (self *PlanLimitNode) GetNodeType() PlanNodeType {
	return LIMITNODE
}

func (self *PlanLimitNode) GetMetadata() *Metadata.Metadata {
	return self.Metadata
}

func (self *PlanLimitNode) SetMetadata() error {
	if err := self.Input.SetMetadata(); err != nil {
		return err
	}
	self.Metadata = self.Input.GetMetadata().Copy()
	return nil
}

func (self *PlanLimitNode) String() string {
	res := "PlanLimitNode {\n"
	res += "Input: " + self.Input.String() + "\n"
	res += "LimitNubmer: " + fmt.Sprint(*self.LimitNumber) + "\n"
	res += "}\n"
	return res
}
