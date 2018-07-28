package EPlan

import (
	"github.com/xitongsys/guery/Metadata"
	. "github.com/xitongsys/guery/Plan"
	"github.com/xitongsys/guery/pb"
)

type EPlanSelectNode struct {
	Location      pb.Location
	Input, Output pb.Location
	SelectItems   []*SelectItemNode
	Having        *BooleanExpressionNode
	Metadata      *Metadata.Metadata
	IsAggregate   bool
}

func (self *EPlanSelectNode) GetNodeType() EPlanNodeType {
	return ESELECTNODE
}

func (self *EPlanSelectNode) GetInputs() []pb.Location {
	return []pb.Location{self.Input}
}

func (self *EPlanSelectNode) GetOutputs() []pb.Location {
	return []pb.Location{self.Output}
}

func (self *EPlanSelectNode) GetLocation() pb.Location {
	return self.Location
}

func NewEPlanSelectNode(node *PlanSelectNode, input, output pb.Location) *EPlanSelectNode {
	return &EPlanSelectNode{
		Location:    output,
		Input:       input,
		Output:      output,
		SelectItems: node.SelectItems,
		Having:      node.Having,
		Metadata:    node.GetMetadata(),
		IsAggregate: node.IsAggregate,
	}
}
