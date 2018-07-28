package Plan

import (
	"github.com/xitongsys/guery/Config"
	"github.com/xitongsys/guery/Metadata"
)

type PlanCombineNode struct {
	Inputs   []PlanNode
	Output   PlanNode
	Metadata *Metadata.Metadata
}

func NewPlanCombineNode(runtime *Config.ConfigRuntime, inputs []PlanNode) *PlanCombineNode {
	return &PlanCombineNode{
		Inputs:   inputs,
		Metadata: Metadata.NewMetadata(),
	}
}

func (self *PlanCombineNode) GetInputs() []PlanNode {
	return self.Inputs
}

func (self *PlanCombineNode) SetInputs(inputs []PlanNode) {
	self.Inputs = inputs
}

func (self *PlanCombineNode) GetOutput() PlanNode {
	return self.Output
}

func (self *PlanCombineNode) SetOutput(output PlanNode) {
	self.Output = output
}

func (self *PlanCombineNode) GetNodeType() PlanNodeType {
	return COMBINENODE
}

func (self *PlanCombineNode) GetMetadata() *Metadata.Metadata {
	return self.Metadata
}

func (self *PlanCombineNode) SetMetadata() (err error) {
	self.Metadata = Metadata.NewMetadata()
	for _, input := range self.Inputs {
		if err = input.SetMetadata(); err != nil {
			return err
		}
		self.Metadata = Metadata.JoinMetadata(self.Metadata, input.GetMetadata())
	}
	return nil
}

func (self *PlanCombineNode) String() string {
	res := "PlanCombineNode {\n"
	for _, n := range self.Inputs {
		res += n.String()
	}
	res += "}\n"
	return res
}
