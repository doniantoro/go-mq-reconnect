package main

import (
	"log"
	"sync"
	"time"

	"github.com/NeowayLabs/wabbit"
	reConnect "github.com/doniantoro/go-mq-reconnect"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	publish()
	consume()
	wg := sync.WaitGroup{}
	wg.Add(1)

	wg.Wait()
	// main3()
}
func publish() {
	conn, err := reConnect.NewRabbitMqConfig().Rabbitmq("amqp://127.0.0.1:5674")
	if err != nil {
		log.Println(err)
	}
	sendCh, err := conn.Channel()
	if err != nil {
		log.Println(err)
	}

	exchangeName := "test-exchange"
	route := "test-route"

	queueName := "test-queue"

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

			_, err = sendCh.QueueDeclare(
				queueName, // name
				wabbit.Option{
					"delete":    false,
					"exclusive": false,
					"noWait":    false,
					"noAct":     true,
					"args": amqp091.Table{
						"alternate-exchange": "my-ae",
					},
				}, // arguments
			)
			if err != nil {
				log.Println(err)
			}
			if err := sendCh.QueueBind(queueName, route, exchangeName, wabbit.Option{
				"noWait": false,
				"noAct":  true,
				"args": amqp091.Table{
					"alternate-exchange": "my-ae",
				},
			}); err != nil {
				log.Println(err)
			}

			err = sendCh.Publish(
				exchangeName,
				route,
				[]byte(time.Now().String()),
				wabbit.Option{
					"contentType": "application/json",
				})

			log.Printf("publish, err: %v", err)
			time.Sleep(5 * time.Second)
		}
	}()

}
func consume() {
	queueName := "test-queue"
	conn, err := reConnect.NewRabbitMqConfig().Rabbitmq("amqp://127.0.0.1:5674")
	if err != nil {
		log.Println(err)
	}
	consumeCh, err := conn.Channel()
	if err != nil {
		log.Println(err)
	}

	go func() {

		d, err := consumeCh.Consume(queueName, "", wabbit.Option{
			"args": amqp091.Table{
				"alternate-exchange": "my-ae",
			},
		})
		if err != nil {
			log.Println(err)
		}

		for msg := range d {
			log.Printf("msg: %s", string(msg.Body()))
			msg.Ack(true)
		}
	}()
}
