package eplan

import (
	"github.com/xitongsys/guery/metadata"
	"github.com/xitongsys/guery/pb"
	. "github.com/xitongsys/guery/plan"
)

type EPlanOrderByNode struct {
	Location  pb.Location
	Inputs    []pb.Location
	Output    pb.Location
	SortItems []*SortItemNode
	Metadata  *metadata.Metadata
}

func (self *EPlanOrderByNode) GetNodeType() EPlanNodeType {
	return EORDERBYNODE
}

func (self *EPlanOrderByNode) GetInputs() []pb.Location {
	return self.Inputs
}

func (self *EPlanOrderByNode) GetOutputs() []pb.Location {
	return []pb.Location{self.Output}
}

func (self *EPlanOrderByNode) GetLocation() pb.Location {
	return self.Location
}

func NewEPlanOrderByNode(node *PlanOrderByNode, inputs []pb.Location, output pb.Location) *EPlanOrderByNode {
	return &EPlanOrderByNode{
		Location:  output,
		Inputs:    inputs,
		Output:    output,
		SortItems: node.SortItems,
		Metadata:  node.GetMetadata(),
	}
}
