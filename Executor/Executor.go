package Executor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/xitongsys/guery/EPlan"
	"github.com/xitongsys/guery/pb"
	"google.golang.org/grpc"
)

type Executor struct {
	MasterAddress string

	Address string
	Name    string

	InputLocations, OutputLocations               []*pb.Location
	InputChannelLocations, OutputChannelLocations []*pb.Location
	Readers                                       []io.Reader
	Writers                                       []io.Writer

	Status          int32
	IsStatusChanged bool
}

var executorServer *Executor

func NewExecutor(masterAddress string, address, name string) *Executor {
	res := &Executor{
		MasterAddress: masterAddress,
		Address:       address,
		Name:          name,
	}
	return res
}

func (self *Executor) Duplicate(ctx context.Context, em *pb.Empty) (*pb.Empty, error) {
	exeFullName, _ := osext.Executable()
	command := exec.Command(exeFullName,
		fmt.Sprintf("executor"),
		fmt.Sprintf("--master %v:%v", self.MasterAddress),
		fmt.Sprintf("--address %v", strings.Split(self.Address, ":")[0]+":0"),
		fmt.Sprintf("--name %v", self.Name),
	)
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	err := command.Start()
	return nil, err
}

func (self *Executor) Quit(ctx context.Context, em *pb.Empty) (*pb.Empty, error) {
	self.Duplicate(context.Background(), nil)
	time.Sleep(10 * time.Second)
	os.Exit(0)
	return nil, nil
}

func (self *Executor) SendInstruction(ctx context.Context, instruction *pb.Instruction) (*pb.Empty, error) {
	nodeType := EPlan.EPlanNodeType(instruction.TaskType)
	switch nodeType {
	case EPlan.ESCANNODE:
		return nil, self.RunScan(instruction)
	case EPlan.ESELECTNODE:
		return nil, self.RunSelect(instruction)
	case EPlan.EGROUPBYNODE:
		return nil, self.RunGroupBy(instruction)
	case EPlan.EJOINNODE:
		return nil, self.RunJoin(instruction)
	case EPlan.EDUPLICATENODE:
		return nil, self.RunDuplicate(instruction)
	default:
		return nil, fmt.Errorf("Unknown node type")
	}
	return nil, nil
}

func (self *Executor) GetOutputChannelLocation(ctx context.Context, location *pb.Location) (*pb.Location, error) {
	if int(location.ChannelIndex) >= len(self.OutputChannelLocations) {
		return nil, fmt.Errorf("ChannelLocation not found: %v", location)
	}
	return self.OutputChannelLocations[location.ChannelIndex], nil
}

///////////////////////////////
func RunExecutor(masterAddress string, address, name string) {
	executorServer = NewExecutor(masterAddress, address, name)
	listener, err := net.Listen("tcp", executorServer.Address)
	if err != nil {
		log.Fatalf("Executor failed to run: %v", err)
	}
	defer listener.Close()
	executorServer.Address = listener.Addr().String()

	go executorServer.Heartbeat()

	grpcS := grpc.NewServer()
	pb.RegisterGueryExecutorServer(grpcS, executorServer)
	grpcS.Serve(listener)
}
