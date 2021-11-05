package main

import (
	"context"
	"math/rand"
	"time"
	"log"
	"net"
	//"reflect"
    "strconv"
    "errors"
    "fmt"
	"sync"
    //"io"
	"google.golang.org/grpc"
	leader_proto "my_packages/grpc_leader"
	nameNode_proto "my_packages/grpc_nameNode"
	pozo_proto "my_packages/grpc_pozo"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	port = "10.6.43.107:50051"
	NameNode = "10.6.43.106:50052"
	minG1    = 6
	maxG1    = 10
	roundsG1 = 4
	minG2 = 1
	maxG2 = 4
	maxNumPlayers = 16
	addressPozo = "10.6.43.105:50056"
)

type Player struct{
    id string
    alive bool
}

var numOfCurrPlayers int = 0
var asignedToTeams int = 0
var gameStarted bool = false
var players []Player
var TeamNumbers = [2]int{0, 0}
var TeamWinner string
var gameBeingPlayed int = 0

var numbersGame1 []int
var numberGame2 int

var waitGroup *sync.WaitGroup
var playedGame int = 0

var c_NameNode nameNode_proto.NNSquidGameClient
var cPozo pozo_proto.PozoClient

var channel *amqp.Channel 
var queue amqp.Queue


type server struct {
	leader_proto.UnimplementedSquidGameServer
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (s *server) JoinGame(ctx context.Context, in *leader_proto.JoinRequest) (*leader_proto.JoinReply, error) {
	if !gameStarted {
		if numOfCurrPlayers < maxNumPlayers {
			
			id := strconv.Itoa(numOfCurrPlayers)
			numOfCurrPlayers += 1

			player := Player{id:id, alive:true}
			players = append(players, player)

			if numOfCurrPlayers == maxNumPlayers {
				gameStarted = true
				log.Printf("16 player have joined The Squid Game")
			}

			return &leader_proto.JoinReply{Message: id}, nil
		}
	}
	return &leader_proto.JoinReply{Message:"-1"}, errors.New("Game already started")
}

func (s *server) SendPlaysG1(ctx context.Context, in *leader_proto.PlayG1) (*leader_proto.State, error) { 
	if gameBeingPlayed == 1 {
		defer waitGroup.Done()

		var sum int = 0
		var alive bool = true
		for i:=0; i < len(in.Numbers); i++ {
			sum += int(in.Numbers[i])

			if int(in.Numbers[i]) >= numbersGame1[i] {
				alive = false
			}
		}

		if(sum != 21) {
			alive = false
		}
		
		if !alive {
			numOfCurrPlayers -= 1
		}


    	_id , _ := strconv.Atoi(in.PlayerId)
    	players[_id].alive = alive

    	if alive {
    		log.Printf("Jugador "+in.PlayerId+" murió en el juego 1")
    	}

    	//SEND TO NAME NODE
	    c_NameNode.NNSendPlaysG1(ctx, &nameNode_proto.NNPlayG1{Numbers:in.Numbers, PlayerId:in.PlayerId})
	    

		return &leader_proto.State{Alive: alive, Winner: false, PlayProcessed: true}, nil
	}

	return &leader_proto.State{PlayProcessed: false}, nil
} 

func (s *server) SendPlayG2(ctx context.Context, in *leader_proto.PlayG2) (*leader_proto.State, error) {

	//Procesar jugadas
	if gameBeingPlayed == 2 {
		defer waitGroup.Done()

		//TODO: ENVIAR MSJ A NAMENODE SOBRE JUGADA DEL JUGADOR

		c_NameNode.NNSendPlaysG2(ctx, &nameNode_proto.NNPlayG2{Number:in.Play, PlayerId:in.PlayerId})
		
		if in.Team == "1" {
			TeamNumbers[0] += int(in.Play)
		} else {
			TeamNumbers[1] += int(in.Play)
		}

		return &leader_proto.State{PlayProcessed: true}, nil
	}

	return &leader_proto.State{PlayProcessed: false}, nil
}

func (s *server) GetTeamG2(ctx context.Context, in *leader_proto.PlayerInfo) (*leader_proto.TeamInfo, error) {
	if gameBeingPlayed == 2 { 
		defer waitGroup.Done()

		//TODO: ENVIAR MSJ A NAMENODE SOBRE TEAM ASIGNADO AL JUGADOR... O SI MURIO

		if numOfCurrPlayers % 2 != 0 {
			numOfCurrPlayers -= 1

			return &leader_proto.TeamInfo{Team:"KILLED"}, nil
		}

		if numOfCurrPlayers / 2 < asignedToTeams {
			asignedToTeams += 1
			return &leader_proto.TeamInfo{Team:"1"}, nil
		} 

		asignedToTeams += 1
		return &leader_proto.TeamInfo{Team:"2"}, nil
	}
	
	return &leader_proto.TeamInfo{Team:"None"}, nil
}

func (s *server) GetResultsG2(ctx context.Context, in *leader_proto.TeamInfo) (*leader_proto.State, error) {
	if playedGame == 2 {
		defer waitGroup.Done()

		if in.Team == TeamWinner || TeamWinner == "both" {
			// Retornar estado vivo y procesado
			return &leader_proto.State{Alive:true, PlayProcessed: true}, nil 
		}

		// Actualizamos info jugador
		i, _ := strconv.Atoi(in.PlayerId)
		players[i].alive = false

		log.Printf("Jugador: "+in.PlayerId+" murió en el juego 2")

		// Retornar estado muerto y procesado
		return &leader_proto.State{Alive:false, PlayProcessed: true}, nil
	}

	// Retornar msj no procesado
	return &leader_proto.State{PlayProcessed: false}, nil
}

func CheckParity(n1 int, n2 int) bool {
	if (n1 + n2) % 2 == 0 {
		return true
	}

	return false
}

func GenerateRandomNumbers(slicer *[]int, size int, min int, max int) {
	for i := 0; i < size+1; i++ {
		*slicer = append(*slicer, rand.Intn(max-min) + min)
	}
}

func Game1() {
    GenerateRandomNumbers(&numbersGame1, roundsG1, minG1, maxG1+1)
	gameBeingPlayed = 1
}

func Game2() {
	numberGame2 = rand.Intn(1+maxG2-minG2) + minG2
	gameBeingPlayed = 2
}

func Game3() {
	gameBeingPlayed = 3
}

func CountAlivePlayers() int {
	cont := 0
	for i := 0; i < len(players); i++ {
		if players[i].alive {
			cont+=1
		}
	}

	return cont
}

func SetTeamWinnerG2() {
	T1Win := CheckParity(TeamNumbers[0], numberGame2)
	T2Win := CheckParity(TeamNumbers[1], numberGame2)

	if T1Win && T2Win {
		TeamWinner = "both"
		return
	}

	if T1Win {
		TeamWinner = "1"
		return
	} else {
		TeamWinner = "2"
		return
	}

	//Se elije al azar equipo ganador
	if rand.Intn(2) == 0 {
		TeamWinner = "1"
		return
	}

	TeamWinner = "2"
	return
}

func PlayGame(game int) {
	wg := new(sync.WaitGroup)
	wg.Add(CountAlivePlayers())
	waitGroup = wg

	if game == 1 {
		Game1()
		wg.Wait()
		playedGame = 1
		gameBeingPlayed = -1
	} else if game == 2 {
		Game2()
		
		// Esperamos a que los jugadores obtengan su equipo
		wg.Wait()

		SetTeamWinnerG2()

		wg := new(sync.WaitGroup)
		wg.Add(asignedToTeams)
		waitGroup = wg

		// Esperamos a que jugadores manden jugada
		wg.Wait()
		playedGame = 2

		wg = new(sync.WaitGroup)
		wg.Add(asignedToTeams)
		waitGroup = wg
		// Esperamos a que jugadores Revisen su jugada
		wg.Wait()
		gameBeingPlayed = -1
	} else if game == 3 {
		Game3()
		wg.Wait()
		playedGame = 3
		gameBeingPlayed = -1
	} else {
		log.Printf("There's no more games to play :(")
	}
}

func CheckWinners() bool {
	cont := 0
	for i:= 0; i < len(players); i++ {
		if players[i].alive {
			cont += 1
		}
	}

	if playedGame == 1 || playedGame == 2 {
		if cont == 1 {
			return true
		}
		return false
	}

	if playedGame == 3 {
		if cont == 1 {
			return true
		}
		return false
	} 

	return false
}

func PrintWinners() {
	for i := 0; i < len(players); i++ {
		if players[i].alive {
			log.Printf("Congratulations! Player " + players[i].id + " winned the Squid Game!")
		}
	}
}

/* Funcion para mandar mensajes a un queue por RabbitMQ */
func Send(msg string){

	err := channel.Publish(
		"",     // exchange
		queue.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
		
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", msg)
}
/* Fin */

func GetMonto() {
    ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
    defer cancel()
    r, err := cPozo.GetMonto(ctx, &pozo_proto.RequestMonto{Request:""})
        
    if err != nil {
        log.Fatal(err)
    }

	log.Printf("Monto acumulado: " + r.MontoPozo)

}

func PrintAlivePlayers() {
	vivos = ""
	for i := 0; i < len(players); i++{
		if players[i].alive == true {
			vivos += (strconv.Itoa(i) + " ")
		}
	}
	if len(vivos) > 0 {
		log.Printf("Jugadores vivos: " + vivos)
	}
}

func Menu() {
	option := 0
	juego  := 1

	flagWinners := false

	for option != 4 {
		log.Printf("Apriete uno de los números para continuar: ")
		log.Printf("1 Dar Comienzo al juego n° " + strconv.Itoa(juego))
		log.Printf("2 Obtener jugadas de un jugador")
		log.Printf("3 Pedir monto pozo")
		log.Printf("4 Salir")
		fmt.Scanf("%d", &option)

		if option == 1 {
			PlayGame(juego)
			juego += 1
			PrintAlivePlayers()
		} else if option == 2 {	

		} else if option == 3 {
			GetMonto()
		} else if option != 4{
			log.Printf("Opcion inválida, elija una opcion entre 1 y 4")
		}

		if CheckWinners() {
			option = 4
			flagWinners = true
		}
	}

	if flagWinners {
		PrintWinners()
	} else {
		log.Printf("There are no winners.")
	}

}

func ConectToNameNode(){
	conn, err := grpc.Dial(NameNode, grpc.WithInsecure(), grpc.WithBlock())
        
    if err != nil {
        log.Printf("did not connect: %v", err)
        defer conn.Close()
        return
    }

    c_NameNode = nameNode_proto.NewNNSquidGameClient(conn)

    log.Printf("ME CONECTE AL NAMENODE: " + NameNode)
}

func ConnectGRPPozo() {
	_Conn, err := grpc.Dial(addressPozo, grpc.WithInsecure(), grpc.WithBlock())
        
	if err != nil {
		log.Printf("did not connect: %v", err)
		defer _Conn.Close()
		return
	}

	cPozo = pozo_proto.NewPozoClient(_Conn)
	log.Printf("ME CONECTE AL POZO: " + addressPozo)
}

func CreateRabbit() {
	/* Conexion a RabbitMQ */
	rabbit, err := amqp.Dial("amqp://guest:guest@localhost:15672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer rabbit.Close()
	
	/* canal de comunicacion */
	ch, err := rabbit.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	
	channel = ch

	/* queue por donde mandar info al pozo */
	q, err := ch.QueueDeclare(
		"pozo", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	failOnError(err, "Failed to declare a queue")

	queue = q
	/* Fin de las conexiones */
}

func main() {
	
	go ConnectGRPPozo()
	
	//CreateRabbit()
	//go Send("JAJAJAJAJAJAJA")

	go ConectToNameNode()

	rand.Seed(time.Now().UnixNano())

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	defer lis.Close()

	s := grpc.NewServer()
	leader_proto.RegisterSquidGameServer(s, &server{})
	
	go Menu() 


	//ABAJO DE ESTE IF NADA
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}