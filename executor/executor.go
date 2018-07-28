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
	"sync"
	"time"

	"github.com/kardianos/osext"
	"github.com/xitongsys/guery/Config"
	"github.com/xitongsys/guery/EPlan"
	"github.com/xitongsys/guery/Logger"
	"github.com/xitongsys/guery/pb"
	"google.golang.org/grpc"
)

type Executor struct {
	sync.Mutex
	AgentAddress string

	Address string
	Name    string

	Instruction                                   *pb.Instruction
	EPlanNode                                     EPlan.ENode
	InputLocations, OutputLocations               []*pb.Location
	InputChannelLocations, OutputChannelLocations []*pb.Location
	Readers                                       []io.Reader
	Writers                                       []io.Writer

	Status          pb.TaskStatus
	IsStatusChanged bool
	Infos           []*pb.LogInfo

	DoneChan chan int
}

var executorServer *Executor

func NewExecutor(agentAddress string, address, name string) *Executor {
	res := &Executor{
		AgentAddress: agentAddress,
		Address:      address,
		Name:         name,
		DoneChan:     make(chan int),
		Infos:        []*pb.LogInfo{},
		Status:       pb.TaskStatus_TODO,
	}
	return res
}

func (self *Executor) AddLogInfo(info interface{}, level pb.LogLevel) {
	if info == nil {
		return
	}
	logInfo := &pb.LogInfo{
		Level: level,
		Info:  []byte(fmt.Sprintf("%v", info)),
	}
	self.Lock()
	defer self.Unlock()
	self.Infos = append(self.Infos, logInfo)
	if level == pb.LogLevel_ERR {
		self.Status = pb.TaskStatus_ERROR
	}
}

func (self *Executor) Clear() {
	//self.Instruction = nil
	//self.EPlanNode = nil
	//self.InputLocations, self.OutputLocations = []*pb.Location{}, []*pb.Location{}
	//self.InputChannelLocations, self.OutputChannelLocations = []*pb.Location{}, []*pb.Location{}
	for _, writer := range self.Writers {
		writer.(io.WriteCloser).Close()
	}
	//self.Readers, self.Writers = []io.Reader{}, []io.Writer{}
	self.IsStatusChanged = true
	if self.Status != pb.TaskStatus_ERROR {
		self.Status = pb.TaskStatus_SUCCEED
	}

	select {
	case <-self.DoneChan:
	default:
		close(self.DoneChan)
	}

}

func (self *Executor) Duplicate(ctx context.Context, em *pb.Empty) (*pb.Empty, error) {
	res := &pb.Empty{}
	exeFullName, _ := osext.Executable()

	command := exec.Command(exeFullName,
		fmt.Sprintf("executor"),
		"--agent",
		fmt.Sprintf("%v", self.AgentAddress),
		"--address",
		fmt.Sprintf("%v", strings.Split(self.Address, ":")[0]+":0"),
		"--config",
		fmt.Sprintf("%v", Config.Conf.File),
	)

	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	err := command.Start()
	return res, err
}

func (self *Executor) Quit(ctx context.Context, em *pb.Empty) (*pb.Empty, error) {
	res := &pb.Empty{}
	os.Exit(0)
	return res, nil
}

func (self *Executor) Restart(ctx context.Context, em *pb.Empty) (*pb.Empty, error) {
	res := &pb.Empty{}
	self.Duplicate(context.Background(), em)
	time.Sleep(time.Second)
	self.Quit(ctx, em)
	return res, nil
}

func (self *Executor) SendInstruction(ctx context.Context, instruction *pb.Instruction) (*pb.Empty, error) {
	res := &pb.Empty{}

	if err := self.SetRuntime(instruction); err != nil {
		return res, err
	}

	nodeType := EPlan.EPlanNodeType(instruction.TaskType)
	Logger.Infof("Instruction: %v", instruction.TaskType)
	self.Status = pb.TaskStatus_RUNNING
	self.IsStatusChanged = true

	self.DoneChan = make(chan int)
	switch nodeType {
	case EPlan.ESCANNODE:
		return res, self.SetInstructionScan(instruction)
	case EPlan.ESELECTNODE:
		return res, self.SetInstructionSelect(instruction)
	case EPlan.EGROUPBYNODE:
		return res, self.SetInstructionGroupBy(instruction)
	case EPlan.EJOINNODE:
		return res, self.SetInstructionJoin(instruction)
	case EPlan.EHASHJOINNODE:
		return res, self.SetInstructionHashJoin(instruction)
	case EPlan.EHASHJOINSHUFFLENODE:
		return res, self.SetInstructionHashJoinShuffle(instruction)
	case EPlan.EDUPLICATENODE:
		return res, self.SetInstructionDuplicate(instruction)
	case EPlan.EAGGREGATENODE:
		return res, self.SetInstructionAggregate(instruction)
	case EPlan.EAGGREGATEFUNCGLOBALNODE:
		return res, self.SetInstructionAggregateFuncGlobal(instruction)
	case EPlan.EAGGREGATEFUNCLOCALNODE:
		return res, self.SetInstructionAggregateFuncLocal(instruction)
	case EPlan.ELIMITNODE:
		return res, self.SetInstructionLimit(instruction)
	case EPlan.EFILTERNODE:
		return res, self.SetInstructionFilter(instruction)
	case EPlan.EUNIONNODE:
		return res, self.SetInstructionUnion(instruction)
	case EPlan.EORDERBYLOCALNODE:
		return res, self.SetInstructionOrderByLocal(instruction)
	case EPlan.EORDERBYNODE:
		return res, self.SetInstructionOrderBy(instruction)
	case EPlan.ESHOWNODE:
		return res, self.SetInstructionShow(instruction)
	case EPlan.EBALANCENODE:
		return res, self.SetInstructionBalance(instruction)
	default:
		self.Status = pb.TaskStatus_TODO
		return res, fmt.Errorf("Unknown node type")
	}
	return res, nil
}

func (self *Executor) Run(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	res := &pb.Empty{}
	nodeType := EPlan.EPlanNodeType(self.Instruction.TaskType)

	switch nodeType {
	case EPlan.ESCANNODE:
		go self.RunScan()
	case EPlan.ESELECTNODE:
		go self.RunSelect()
	case EPlan.EGROUPBYNODE:
		go self.RunGroupBy()
	case EPlan.EJOINNODE:
		go self.RunJoin()
	case EPlan.EHASHJOINNODE:
		go self.RunHashJoin()
	case EPlan.EHASHJOINSHUFFLENODE:
		go self.RunHashJoinShuffle()
	case EPlan.EDUPLICATENODE:
		go self.RunDuplicate()
	case EPlan.EAGGREGATENODE:
		go self.RunAggregate()
	case EPlan.EAGGREGATEFUNCGLOBALNODE:
		go self.RunAggregateFuncGlobal()
	case EPlan.EAGGREGATEFUNCLOCALNODE:
		go self.RunAggregateFuncLocal()
	case EPlan.ELIMITNODE:
		go self.RunLimit()
	case EPlan.EFILTERNODE:
		go self.RunFilter()
	case EPlan.EORDERBYLOCALNODE:
		go self.RunOrderByLocal()
	case EPlan.EORDERBYNODE:
		go self.RunOrderBy()
	case EPlan.EUNIONNODE:
		go self.RunUnion()
	case EPlan.ESHOWNODE:
		go self.RunShow()
	case EPlan.EBALANCENODE:
		go self.RunBalance()
	default:
		return res, fmt.Errorf("Unknown node type")
	}
	return res, nil
}

func (self *Executor) GetOutputChannelLocation(ctx context.Context, location *pb.Location) (*pb.Location, error) {
	if int(location.ChannelIndex) >= len(self.OutputChannelLocations) {
		return nil, fmt.Errorf("ChannelLocation %v not found: %v", location.ChannelIndex, location)
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
	Logger.Infof("Executor: %v", executorServer.Address)

	go executorServer.Heartbeat()

	grpcS := grpc.NewServer()
	pb.RegisterGueryExecutorServer(grpcS, executorServer)
	grpcS.Serve(listener)
}
