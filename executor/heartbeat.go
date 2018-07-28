package Executor

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/xitongsys/guery/Logger"
	"github.com/xitongsys/guery/pb"
	"google.golang.org/grpc"
)

func (self *Executor) Heartbeat() {
	for {
		if err := self.DoHeartbeat(); err != nil {
			time.Sleep(3 * time.Second)
		}
	}
}

func (self *Executor) DoHeartbeat() error {
	grpcConn, err := grpc.Dial(self.AgentAddress, grpc.WithInsecure())
	if err != nil {
		Logger.Errorf("DoHeartBeat failed: %v", err)
		return err
	}
	defer grpcConn.Close()

	client := pb.NewGueryAgentClient(grpcConn)
	stream, err := client.SendHeartbeat(context.Background())
	if err != nil {
		return err
	}

	ticker := time.NewTicker(1 * time.Second)
	quickTicker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-quickTicker.C:
			if self.IsStatusChanged {
				self.IsStatusChanged = false
				if err := self.SendOneHeartbeat(stream); err != nil {
					return err
				}
			}
		case <-ticker.C:
			if err := self.SendOneHeartbeat(stream); err != nil {
				return err
			}
		}
	}
}

func (self *Executor) SendOneHeartbeat(stream pb.GueryAgent_SendHeartbeatClient) error {
	address, ports, err := net.SplitHostPort(self.Address)
	if err != nil {
		return err
	}
	var port int32
	fmt.Sscanf(ports, "%d", &port)

	hb := &pb.ExecutorHeartbeat{
		Location: &pb.Location{
			Name:    self.Name,
			Address: address,
			Port:    port,
		},
		Status: self.Status,
		Infos:  self.Infos,
	}

	if self.Instruction != nil {
		hb.TaskId = self.Instruction.TaskId
	}

	if err := stream.Send(hb); err != nil {
		Logger.Errorf("failed to SendOneHeartbeat: %v, %v", err, self.AgentAddress)
		return err
	}
	return nil
}
