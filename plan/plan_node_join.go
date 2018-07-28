package Plan

import (
	"fmt"

	"github.com/xitongsys/guery/Config"
	"github.com/xitongsys/guery/Metadata"
)

type JoinType int32

const (
	_ JoinType = iota
	LEFTJOIN
	RIGHTJOIN
	INNERJOIN
)

type PlanJoinNode struct {
	Metadata              *Metadata.Metadata
	LeftInput, RightInput PlanNode
	Output                PlanNode
	JoinType              JoinType
	JoinCriteria          *JoinCriteriaNode
}

func NewPlanJoinNode(runtime *Config.ConfigRuntime, leftInput PlanNode, rightInput PlanNode, joinType JoinType, joinCriteria *JoinCriteriaNode) *PlanJoinNode {
	res := &PlanJoinNode{
		Metadata:     Metadata.NewMetadata(),
		LeftInput:    leftInput,
		RightInput:   rightInput,
		JoinType:     joinType,
		JoinCriteria: joinCriteria,
	}
	return res
}

func (self *PlanJoinNode) GetInputs() []PlanNode {
	return []PlanNode{self.LeftInput, self.RightInput}
}

func (self *PlanJoinNode) SetInputs(inputs []PlanNode) {
	self.LeftInput, self.RightInput = inputs[0], inputs[1]
}

func (self *PlanJoinNode) GetOutput() PlanNode {
	return self.Output
}

func (self *PlanJoinNode) SetOutput(output PlanNode) {
	self.Output = output
}

func (self *PlanJoinNode) GetNodeType() PlanNodeType {
	return JOINNODE
}

func (self *PlanJoinNode) GetMetadata() *Metadata.Metadata {
	return self.Metadata
}

func (self *PlanJoinNode) SetMetadata() (err error) {
	if err = self.LeftInput.SetMetadata(); err != nil {
		return err
	}
	if err = self.RightInput.SetMetadata(); err != nil {
		return err
	}

	mdl, mdr := self.LeftInput.GetMetadata(), self.RightInput.GetMetadata()
	self.Metadata = Metadata.JoinMetadata(mdl, mdr)
	return nil
}

func (self *PlanJoinNode) String() string {
	res := "PlanJoinNode {\n"
	res += "LeftInput: " + self.LeftInput.String() + "\n"
	res += "RightInput: " + self.RightInput.String() + "\n"
	res += "JoinType: " + fmt.Sprint(self.JoinType) + "\n"
	res += "JoinCriteria: " + fmt.Sprint(self.JoinCriteria) + "\n"
	res += "}\n"
	return res
}
