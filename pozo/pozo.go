/* https://www.rabbitmq.com/#getstarted */

package main

import (
	"log"
	"context"
	"net"
	"google.golang.org/grpc"
	"fmt"
	pozo_proto "my_packages/grpc_pozo"
	//amqp "github.com/rabbitmq/amqp091-go"
)

const (
	port = "10.6.43.105:50056"
)

var monto int

type server struct {
	pozo_proto.UnimplementedPozoServer
}

func (s *server) GetMonto(ctx context.Context, in *pozo_proto.RequestMonto) (*pozo_proto.Monto, error) {
	return &pozo_proto.Monto{MontoPozo:"9999999"}, nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func CreateGRPCServer(){
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	defer lis.Close()

	s := grpc.NewServer()
	pozo_proto.RegisterPozoServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	
	/*conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"pozo", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	*/

	go CreateGRPCServer()

	var option int = 0
	fmt.Scanf("%d", &option)

}