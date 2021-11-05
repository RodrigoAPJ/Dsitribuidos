package main

import (
    "context"
    "log"
    "math/rand"
    "time"
    //"errors"
    //"strings"
    "sync"
    "strconv"
    "fmt"
    "google.golang.org/grpc"
    leader_proto "my_packages/grpc_leader"
)

const (
    address     = "10.6.43.107:50051"
    defaultName = "0"
    maxNumPlayers = 16
    minG1    = 1
    maxG1    = 10
    roundsG1 = 4
)

var gameOver bool = false

type Player struct {
    id string
    alive bool
    winner bool
    team string
    c leader_proto.SquidGameClient
}

func PrintPlayers(players []Player, size int) {
    for i := 0; i < size; i++ {
        var state string
        if players[i].alive {
            state = "alive"
        } else {
            state = "dead"
        }
        log.Printf("%v is %v", players[i].id, state)
    }
}

func Choose4Nums(smartBots bool) []int64 {
    givenNumbers := 0
    number := 0
    sum := 0
    var numbers []int64

    if smartBots {
        pos := rand.Intn(4)

        numbers = append(numbers, 5)
        numbers = append(numbers, 5)
        numbers = append(numbers, 5)
        numbers = append(numbers, 5)

        numbers[pos] = 6
        return numbers
    }


    for givenNumbers < roundsG1 && sum < 21{
        number = rand.Intn(maxG1+1-minG1) + minG1
        if number >= minG1 && number <= maxG1 {
            if number + sum <= 21 {
                sum += number
                givenNumbers += 1
            } else {
                number = 21 - sum
                givenNumbers = roundsG1 + 1
                sum = 21
            }
            numbers = append(numbers, int64(number))
        }  
    }

    return numbers
}

func PlayG1(player *Player, player_index int, players *[]Player) {
    wg := new(sync.WaitGroup)
    wg.Add(16)

    var smartBots bool = false
    var response int = 0
    log.Printf("¿Quiere que los bots hagan jugadas inteligentes? 1 = SI, cualquier otro input = NO")
    fmt.Scanf("%d", &response)
    if response == 1 {
        smartBots = true
    }


    PlayerPlayGame1(player, wg)

    //Bots juegan juego 1
    for i := 0; i < len(*players); i++{
        if i != player_index {
            go BotPlayGame1(&((*players)[i]), smartBots, wg)
        }
    }            
    //Esperamos a las jugadas del jugador y los bots reciban respuesta del servidor
    wg.Wait()
}

func PlayerPlayGame1(player *Player, wg *sync.WaitGroup) {
    defer wg.Done()

    var plays leader_proto.PlayG1;
    var choosenNumbers []int64
    givenNumbers := 0
    var number int64
    var sum int64

    log.Printf("Ingrese máximo 4 números para que sumen 21") 
    for givenNumbers < roundsG1 && sum < 21 {
        log.Printf("Ingrese un número entre 1 y 10")
        fmt.Scanf("%d", &number)

        if number >= minG1 && number <= maxG1 {
            if number + sum <= 21 {
                sum += number
                givenNumbers += 1
            } else {
                number = 21 - sum
                givenNumbers = roundsG1 + 1
                sum = 21
            }
            choosenNumbers = append(choosenNumbers, number)
        }  
    }
    plays.Numbers = choosenNumbers
    plays.PlayerId = (*player).id 
    
    SendPlayGame1(player, plays)
}

func BotPlayGame1(player *Player, smartBots bool, wg *sync.WaitGroup) {
    defer wg.Done()
    var plays leader_proto.PlayG1;
    choosenNumbers := Choose4Nums(smartBots)
    plays.Numbers = choosenNumbers
    plays.PlayerId = (*player).id
    SendPlayGame1(player, plays)
}

func SendPlayGame1(player *Player, plays leader_proto.PlayG1) {
    c := (*player).c
    ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
    defer cancel()
    r, err := c.SendPlaysG1(ctx, &plays)
        
    if err != nil {
        log.Fatal(err)
    }
    
    if r.GetPlayProcessed() {
        (*player).alive = r.GetAlive()
    } else {
        log.Printf("Jugada del jugador "+ (*player).id +" no procesada")
    }
}

func RequestTeam(player *Player, waitGroup *sync.WaitGroup) {
    defer waitGroup.Done()

    c := (*player).c
    ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
    defer cancel()
    r, err := c.GetTeamG2(ctx, &leader_proto.PlayerInfo{PlayerId:(*player).id})
        
    if err != nil {
        log.Fatal(err)
    }

    if r.Team == "None" {
        log.Printf("Player " + (*player).id + " couldn't play Game 2")
        return 
    }

    if r.Team == "KILLED" {
        log.Printf("Player "+ (*player).id +" get KILLED by leader, given the odd number of participants")
        (*player).alive = false      
    } else if r.Team == "1" {
        (*player).team = "1"
        log.Printf("Player " + (*player).id + " is on Team 1")
    } else if r.Team == "2" {
        (*player).team = "2"
        log.Printf("Player " + (*player).id + " is on Team 2")
    }

    return
}

func CountAlivePlayers(players []Player) int {
    cont := 0
    for i := 0; i < len(players); i++ {
        if players[i].alive {
            cont+=1
        }
    }

    return cont
}

func SendPlayG2(player *Player, num int , playsG2Processed *bool, waitGroup *sync.WaitGroup) {
    defer waitGroup.Done()

    c := (*player).c
    ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
    defer cancel()
    r, err := c.SendPlayG2(ctx, &leader_proto.PlayG2{Play:int64(num), PlayerId: (*player).id, Team: (*player).team})
        
    if err != nil {
        log.Fatal(err)
    }

    if r.GetPlayProcessed() {
        log.Printf("Player " + (*player).id + " played " + strconv.Itoa(num) + " for Team " + (*player).team)
        *playsG2Processed = true
        return
    }
    log.Printf("Play of Player " + (*player).id + " did not get processed by Leader...")
}

func GetResultsG2(player *Player, waitGroup *sync.WaitGroup) {
    defer waitGroup.Done()

    c := (*player).c
    ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
    defer cancel()
    r, err := c.GetResultsG2(ctx, &leader_proto.TeamInfo{Team:(*player).team, PlayerId: (*player).id})
        
    if err != nil {
        log.Fatal(err)
    }

    if r.GetPlayProcessed() {
        if !r.Alive {
            log.Printf("Player " + (*player).id + " died")
            (*player).alive = false
            return
        }
        log.Printf("Player " + (*player).id + " DIDN'T die")
    }



    return 
}

func PlayG2(player *Player, players *[]Player) {

    waitGroup := new(sync.WaitGroup)
    waitGroup.Add(CountAlivePlayers(*players))

    // Obtener equipo para jugadores vivos
    for i := 0; i < len(*players); i++ {
        if (*players)[i].alive {
            go RequestTeam(&((*players)[i]), waitGroup)
        }
    }

    waitGroup.Wait()


    teamsExist := false
    for i := 0; i < len(*players); i++ {
        if (*players)[i].team != "" {
            teamsExist = true
        }
    }

    if !teamsExist {
        return
    }

    var playsG2Processed bool = false
    waitGroup = new(sync.WaitGroup)
    waitGroup.Add(CountAlivePlayers(*players))

    //jej := strconv.Itoa(CountAlivePlayers(*players))
    //log.Printf(jej)

    // Si seguimos vivos pedir numero entre 1 y 4
    if (*player).alive {
        // Pedir input y mandar jugada
        number := -1

        for number < 0 || number >= 4 {
            log.Printf("Inserte número entre 1 y 4: ")
            fmt.Scanf("%d", &number)
            if number < 1 || number > 4 {
                log.Printf("Inserte un número entre 1 y 4 por favor")
            }
        }
        go SendPlayG2(player, number , &playsG2Processed, waitGroup)
    }

    // Generar numeros para cada bot vivo entre 1 y 4
    for i := 0; i < len(*players); i++ {
        if (*players)[i].alive && (*players)[i].id != (*player).id {
            go SendPlayG2(&((*players)[i]), rand.Intn(4) + 1, &playsG2Processed, waitGroup)
        }
    }

    // Esperar la cantidad de respuestas correspondientes
    waitGroup.Wait()

    if !playsG2Processed {
        return
    }

    // Preguntar por Resultados
    waitGroup = new(sync.WaitGroup)
    waitGroup.Add(CountAlivePlayers(*players))
    for i := 0; i < len(*players); i++ {
        if (*players)[i].alive {
            // Preguntar por Resultados con GRPC
            go GetResultsG2(&((*players)[i]), waitGroup)
        }
    }
    // Procesar y esperar respuestas
    waitGroup.Wait()
}

func HacerJugadaJuego3(player Player) {

}

func CheckPlayersStates(players []Player) ([]Player, []Player) {
    var winners []Player
    var alives []Player

    for i := 0; i < len(players); i++ {
        if players[i].winner{
            winners = append(winners, players[i])
        }
        if players[i].alive{
            alives = append(alives, players[i])
        }
    }

    return alives, winners
}

func PrintWinners(winners []Player) {
    for i := 0; i < len(winners); i++ {
        log.Printf("Jugador: " + winners[i].id + " a ganado")
    }
}

func PrintSurvivors(survivors []Player) {
    for i := 0; i < len(survivors); i++ {
        log.Printf("Jugador: " + survivors[i].id + " sigue vivo")
    }
}

func Menu(player *Player, players *[]Player){
    option := 0
    player_index, _ := strconv.Atoi((*player).id)

    for option != 3{
        log.Printf("Apriete uno de los números para continuar: ")
        log.Printf("1 Hacer jugada para juego 1")
        log.Printf("2 Hacer jugada para juego 2")
        log.Printf("3 Hacer jugada para juego 3")
        log.Printf("4 Salir")
        fmt.Scanf("%d", &option)

        if option == 1 {
            PlayG1(player, player_index, players)
        } else if option == 2 {
            PlayG2(player, players)
        } else if option == 3 {
            HacerJugadaJuego3(*player)
        } else if option == 4 {
            return
        } else {
            log.Printf("Opcion inválida, elija una opcion entre 1 y 4")
        }

        if option >= 1 || option <= 3 {
            alives, winners := CheckPlayersStates(*players)
            
            if len(winners) > 0 {
                PrintWinners(winners)
                return
            } else {
                if len(alives) > 0 {
                    log.Printf("¿Ver supervivientes?")
                    log.Printf("0 para no verlos")
                    log.Printf("1 para si verlos")
                    fmt.Scanf("%d", &option)
                    if option == 1 {
                        PrintSurvivors(alives)
                    }
                } else {
                    log.Printf("There are no winners or alive players")
                    return
                }   
            }
        }

    }
}

func JoinSquidGameServer(players *[]Player) {

    for i := 0; i < maxNumPlayers; i++ {
        // Set up a connection to the server.
        conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
        
        if err != nil {
            log.Printf("did not connect: %v", err)
            defer conn.Close()
            return
        }

        c := leader_proto.NewSquidGameClient(conn)

        // Contact the server and print out its response.
        id := defaultName
        ctx, cancel := context.WithTimeout(context.Background(), 6 * time.Second)
        defer cancel()
        
        r, err := c.JoinGame(ctx, &leader_proto.JoinRequest{Name: id})
        
        if err != nil {
            log.Printf("Error, could not connect")
            return
        } 

        id = r.GetMessage()

        player := Player{id:id, alive:true, winner:false, team:"", c:c}
        *players = append(*players, player)
        
        if( i + 1 == maxNumPlayers) {
            log.Printf("Soy el jugador %v", id)
        }
    }
    
    if len(*players) == maxNumPlayers {
        log.Printf("16 jugadores han ingresado al juego")
    } else {
        log.Printf("No han logrado ingresar 16 jugadores... reinicie el juego")
    }
}


func main() {
   
    var players []Player
    JoinSquidGameServer(&players)

    Menu(&players[len(players)-1], &players)    

}