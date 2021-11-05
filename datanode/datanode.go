package main

import (
	"context"
	//"math/rand"
	//"time"
	"log"
	"net"
    "strconv"
    //"errors"
    //"fmt"
	"sync"
	"os"
	"google.golang.org/grpc"
	dataNode_proto "my_packages/grpc_datanode"
)

const (
    dataNode1 = "10.6.43.105:50053"
    dataNode2 = "10.6.43.105:50054"
    dataNode3 = "10.6.43.105:50055" 
    path1 = "DN1/"
    path2 = "DN2/"
    path3 = "DN3/"
)

var mu sync.Mutex

type server struct {
	dataNode_proto.UnimplementedDNSquidGameServer
}

func (s *server) DNSendPlayG2(ctx context.Context, in *dataNode_proto.DNPlayG2) (*dataNode_proto.DNConfirmation, error) {
	
	var arr []int64
	arr = append(arr, in.Number)

	WriteOnFile(arr, in.PlayerId, "2", in.Node)
	
	return &dataNode_proto.DNConfirmation{Processed: true}, nil
}

func (s *server) DNSendPlaysG1(ctx context.Context, in *dataNode_proto.DNPlayG1) (*dataNode_proto.DNConfirmation, error) {
	
	WriteOnFile(in.Numbers, in.PlayerId, "1", in.Node)
	
	return &dataNode_proto.DNConfirmation{Processed: true}, nil
}

/*func (s *server) ReceivePlaysG1(ctx context.Context, in *dataNode_proto.Player) (*dataNode_proto.PlayG1, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReceivePlaysG1 not implemented")
}*/

func WriteOnFile(numbers []int64, playerId string, phase string, node string) {
	mu.Lock()
	var auxPath string = "./"

	if node == "1" {
		auxPath = path1
	} else if node == "2" {
		auxPath = path2
	} else {
		auxPath = path3
	}

	//file, err := os.Create("jugador_"+playerId+"__ronda_1.txt")
	file, err := os.OpenFile(auxPath+"jugador_"+playerId+"__ronda_"+phase+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
        log.Fatal(err)
    }
    
    defer file.Close()
    
	for i := 0; i < len(numbers); i++{
		num := int(numbers[i])
		_num := strconv.Itoa(num)
	    if _, err := file.WriteString(_num + "\n"); err != nil {
	    	log.Fatalf("Error writing on file "+ "jugador_"+playerId+"__ronda_1.txt")
	    	mu.Unlock()
	    	
	    	return
	    }	
	}

    mu.Unlock()
}

func OpenDataNodeServer(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	defer lis.Close()

	s := grpc.NewServer()
	dataNode_proto.RegisterDNSquidGameServer(s, &server{})

	log.Printf("Server " + port + " opened and listening")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}


func main() {
	lis, err := net.Listen("tcp", dataNode1)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	defer lis.Close()

	s := grpc.NewServer()
	dataNode_proto.RegisterDNSquidGameServer(s, &server{})

	go OpenDataNodeServer(dataNode2)
	go OpenDataNodeServer(dataNode3)

	log.Printf("Server "+dataNode1+ "opened and listening")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
