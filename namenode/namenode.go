package main

import (
	"context"
	"math/rand"
	"time"
	"log"
	"net"
    //"strconv"
    //"errors"
    //"fmt"
	"sync"
	"os"
	"google.golang.org/grpc"
	nameNode_proto "my_packages/grpc_nameNode"
	dataNode_proto "my_packages/grpc_dataNode"
)

const (
	port = "10.6.43.106:50052"
    dataNode1 = "10.6.43.105:50053"
    dataNode2 = "10.6.43.105:50054"
    dataNode3 = "10.6.43.105:50055" 
)

var mu sync.Mutex

var c_dN1 dataNode_proto.DNSquidGameClient
var c_dN2 dataNode_proto.DNSquidGameClient
var c_dN3 dataNode_proto.DNSquidGameClient

type server struct {
	nameNode_proto.UnimplementedNNSquidGameServer
}

func (s *server) NNSendPlaysG1(ctx context.Context, in *nameNode_proto.NNPlayG1) (*nameNode_proto.NNConfirmation, error) {
	node := rand.Intn(3)
	r := false

	if node == 0 {
		WriteOnRegister(in.PlayerId, dataNode1, "1")
		r = SendToDataNode(ctx, dataNode1, in.GetNumbers(), in.GetPlayerId(), "1")
	} else if node == 1 {
		WriteOnRegister(in.PlayerId, dataNode2, "1")
		r = SendToDataNode(ctx, dataNode2, in.GetNumbers(), in.GetPlayerId(), "2")
	} else {
		WriteOnRegister(in.PlayerId, dataNode3, "1")
		r = SendToDataNode(ctx, dataNode3, in.GetNumbers(), in.GetPlayerId(), "3")
	}

	return &nameNode_proto.NNConfirmation{Processed: r}, nil
}

func (s *server) NNSendPlaysG2(ctx context.Context, in *nameNode_proto.NNPlayG2) (*nameNode_proto.NNConfirmation, error) {
	node := rand.Intn(3)
	r := false

	if node == 0 {
		WriteOnRegister(in.PlayerId, dataNode1, "2")
		r = SendToDataNode2(ctx, dataNode1, in.Number, in.GetPlayerId(), "1")
	} else if node == 1 {
		WriteOnRegister(in.PlayerId, dataNode2, "2")
		r = SendToDataNode2(ctx, dataNode2, in.Number, in.GetPlayerId(), "2")
	} else {
		WriteOnRegister(in.PlayerId, dataNode3, "2")
		r = SendToDataNode2(ctx, dataNode3, in.Number, in.GetPlayerId(), "3")
	}

	return &nameNode_proto.NNConfirmation{Processed: r}, nil
}

func SendToDataNode2 (ctx context.Context, dN string, number int64, player string, node string) bool {
	var c dataNode_proto.DNSquidGameClient
    
    if dN == dataNode1 {
    	c = c_dN1
    } else if dN == dataNode2 {
    	c = c_dN2
    } else {
    	c = c_dN3
    }

	r, err := c.DNSendPlayG2(ctx, &dataNode_proto.DNPlayG2{Number: number, PlayerId: player, Node: node})

    if err != nil {
        log.Printf("Error, could not connect")
        return false
    } 
   	if r.GetProcessed() {
   		return true
   	} 

   	return false
}


func SendToDataNode(ctx context.Context, dN string, numbers []int64, player string, node string) bool{
    var c dataNode_proto.DNSquidGameClient
    
    if dN == dataNode1 {
    	c = c_dN1
    } else if dN == dataNode2 {
    	c = c_dN2
    } else {
    	c = c_dN3
    }

	r, err := c.DNSendPlaysG1(ctx, &dataNode_proto.DNPlayG1{Numbers: numbers, PlayerId: player, Node: node})

    if err != nil {
        log.Printf("Error, could not connect")
        return false
    } 
   	if r.GetProcessed() {
   		return true
   	} 

   	return false
}

/*func (s *server) ReceivePlaysG1(ctx context.Context, in *nameNode_proto.Player) (*nameNode_proto.PlayG1, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReceivePlaysG1 not implemented")
}*/

func WriteOnRegister(playerId string, node string, phase string) {
	mu.Lock()

	//file, err := os.Create("jugador_"+playerId+"__ronda_1.txt")
	//WLS file, err := os.OpenFile("Registro.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	file, err := os.OpenFile("Registro.txt", os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
        log.Fatal(err)
    }
    
    defer file.Close()
    
    if _, err := file.WriteString("Jugador_" + playerId + " Ronda_" + phase + " ip_" + node + "\n"); err != nil {
    	log.Fatalf("Error writing on file "+ "jugador_"+playerId+"__ronda_1.txt")
    	mu.Unlock()
    	
    	return
    }

    mu.Unlock()
}

func ConectToDataNode(dN string) {
	conn, err := grpc.Dial(dN, grpc.WithInsecure(), grpc.WithBlock())
        
    if err != nil {
        log.Printf("did not connect: %v", err)
        defer conn.Close()
        return
    }

    if dN == dataNode1 {
    	c_dN1 = dataNode_proto.NewDNSquidGameClient(conn)
    } else if dN == dataNode2 {
    	c_dN2 = dataNode_proto.NewDNSquidGameClient(conn)
    } else {
    	c_dN3 = dataNode_proto.NewDNSquidGameClient(conn)
    }
    log.Printf("ME CONECTE AL DATANODE "+ dN)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ConectToDataNode(dataNode1)
	ConectToDataNode(dataNode2)
	ConectToDataNode(dataNode3)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	defer lis.Close()

	s := grpc.NewServer()
	nameNode_proto.RegisterNNSquidGameServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
