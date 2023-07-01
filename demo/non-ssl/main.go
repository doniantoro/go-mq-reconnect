package main

import (
	"log"
	"sync"
	"time"

	"github.com/NeowayLabs/wabbit"
	reConnect "github.com/doniantoro/go-mq-reconnect/v3"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	publish()
	consume()
	wg := sync.WaitGroup{}
	wg.Add(1)

	wg.Wait()
}
func publish() {
	conn, err := reConnect.NewRabbitMqConfig().Rabbitmq("amqp://127.0.0.1:5674")
	if err != nil {
		log.Panic(err)
	}

	sendCh, err := conn.Channel()
	if err != nil {
		log.Panic(err)
	}

	exchangeName := "test-exchange"
	route := "test-route"

	go func() {
		for {

			err = sendCh.ExchangeDeclare(exchangeName,
				"topic", wabbit.Option{
					"durable":  true,
					"delete":   false,
					"internal": false,
					"noWait":   false,
					"args": amqp091.Table{
						"alternate-exchange": "my-ae",
					},
				})
			if err != nil {
				log.Println(err)
			}
			err = sendCh.Publish(
				exchangeName,
				route,
				[]byte(time.Now().String()),
				wabbit.Option{
					"contentType": "application/json",
				})

			// fmt.Println("sendCh222", sendCh.IsClosed())
			log.Printf("publish, err: %v", err)
			time.Sleep(5 * time.Second)
		}
	}()

}
func consume() {
	exchangeName := "test-exchange"
	route := "test-route"
	queueName := "test-queue"
	conn, err := reConnect.NewRabbitMqConfig().Rabbitmq("amqp://127.0.0.1:5674")
	if err != nil {
		log.Panic(err)
	}
	consumeCh, err := conn.Channel()
	if err != nil {
		log.Panic(err)
	}
	_, err = consumeCh.QueueDeclare(
		queueName, // name
		wabbit.Option{
			"delete":    false,
			"exclusive": false,
			"noWait":    false,
			"args": amqp091.Table{
				"alternate-exchange": "my-ae",
			},
		}, // arguments
	)

	if err := consumeCh.QueueBind(queueName, route, exchangeName, wabbit.Option{
		"noWait": false,
	}); err != nil {
		log.Panic(err)
	}

	go func() {
		d, err := consumeCh.Consume(queueName, "", nil)
		if err != nil {
			log.Panic(err)
		}

		for msg := range d {
			log.Printf("msg: %s", string(msg.Body()))
			msg.Ack(true)
		}
	}()
}
