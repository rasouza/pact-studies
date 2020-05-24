package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/streadway/amqp"
)

var pact = dsl.Pact{
	Consumer: "MyConsumer",
	Provider: "MyProvider",
	LogDir:   logDir,
	PactDir:  pactDir,
	LogLevel: "DEBUG",
}

func TestConsumer(t *testing.T) {
	message := pact.AddMessage()
	message.
		Given("interaction exists").
		ExpectsToReceive("a interaction").
		WithContent(map[string]interface{}{
			"merchant_code": dsl.Like("MCL3HPF3"),
			"referral_code": dsl.Like("ochYN"),
		})
	pact.VerifyMessageConsumer(t, message, consumeHandlerWrapper)
}

var consumeHandlerWrapper = func(m dsl.Message) error {
	return consumeHandler()
}

var consumeHandler = func() error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
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

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	return nil
}

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
