package Master

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/soheilhy/cmux"
	"github.com/xitongsys/guery/Logger"
	"github.com/xitongsys/guery/Scheduler"
	"github.com/xitongsys/guery/Topology"
	"github.com/xitongsys/guery/pb"
	"google.golang.org/grpc"
)

var masterServer *Master

type Master struct {
	Topology  *Topology.Topology
	Scheduler *Scheduler.Scheduler
	StartTime time.Time
}

func NewMaster() *Master {
	m := &Master{
		Topology:  Topology.NewTopology(),
		StartTime: time.Now(),
	}
	m.Scheduler = Scheduler.NewScheduler(m.Topology)
	return m
}

func (self *Master) SendHeartbeat(stream pb.GueryMaster_SendHeartbeatServer) error {
	var location *pb.Location
	for {
		hb, err := stream.Recv()
		if err == nil {
			if location == nil {
				location = hb.Location
				Logger.Infof("Add executor %v", location)
			}

		} else {
			if location != nil {
				self.Topology.DropExecutorInfo(location)
				Logger.Infof("Lost executor %v: %v", location, err)
			}
			if err == io.EOF {
				Logger.Infof("Lost executor %v: %v", location, err)
				return nil
			}
			if err != nil {
				Logger.Infof("Lost executor %v: %v", location, err)
				return err
			}
		}
		self.Topology.UpdateExecutorInfo(hb)

	}
}

func (self *Master) JobHandler(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("hello,world"))
	Logger.Infof("JobHandler")
}

///////////////////////////
func RunMaster(address string) {
	masterServer = NewMaster()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Master failed to start on %v: %v", address, err)
	}
	defer listener.Close()

	m := cmux.New(listener)

	grpcL := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpL := m.Match(cmux.Any())

	grpcS := grpc.NewServer()
	pb.RegisterGueryMasterServer(grpcS, masterServer)

	r := mux.NewRouter()
	r.HandleFunc("/", masterServer.UIHandler)
	r.HandleFunc("/query", masterServer.QueryHandler)
	r.HandleFunc("/job/{id:[0-9]+}", masterServer.JobHandler)
	httpS := &http.Server{Handler: r}

	go grpcS.Serve(grpcL)
	go httpS.Serve(httpL)

	masterServer.Scheduler.AutoFresh()

	if err := m.Serve(); err != nil {
		log.Fatalf("Master failed to serve: %v", err)

	}

}
