package main

import (
	"log"
	"sync"
	"time"

	"github.com/NeowayLabs/wabbit"
	reConnect "github.com/doniantoro/go-mq-reconnect"
	"github.com/rabbitmq/amqp091-go"

	// "github.com/rabbitmq/amqp091-go"
	"github.com/streadway/amqp"
)

func main() {

	publishWabitv2()
	// publishStreadway()
	// consumeStreadway()
	// publishWabbit()
	// consumeWabbit()
	wg := sync.WaitGroup{}
	wg.Add(1)

	time.Sleep(3 * time.Second)
	wg.Wait()
}
func publishWabitv2() {
	conn, err := reConnect.NewRabbitMqConfig(reConnect.Wabbit).Rabbitmq("amqp://127.0.0.1:5674")
	if err != nil {
		log.Panic(err)
	}

	// sendCh := conn.Channel().Wabbitv2()

}
func publishStreadway() {
	conn, err := reConnect.NewRabbitMqConfig(reConnect.Streadway).Rabbitmq("amqp://127.0.0.1:5674")
	if err != nil {
		log.Panic(err)
	}

	sendCh := conn.Channel()

	if sendCh == nil {
		log.Println("sendCh", sendCh)
	}

	exchangeName := "test-exchange-streadway"
	route := "test-route"

	err = sendCh.ExchangeDeclare(exchangeName, amqp.ExchangeTopic, true, false, false, false, amqp.Table{"alternate-exchange": "my-ae"})
	if err != nil {
		log.Panic(err)
	}

	go func() {
		for {

			err := sendCh.Publish(exchangeName, route, false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(time.Now().String()),
			})

			log.Printf("publish from streadway, err: %v", err)
			time.Sleep(5 * time.Second)
		}
	}()

}

func consumeStreadway() {
	exchangeName := "test-exchange-streadway"
	route := "test-route"
	queueName := "test-queue-streadway"
	conn, err := reConnect.NewRabbitMqConfig(reConnect.Streadway).Rabbitmq("amqp://127.0.0.1:5674")
	if err != nil {
		log.Panic(err)
	}
	consumeCh := conn.Channel()
	if consumeCh.Err != nil {
		log.Panic(err)
	}
	_, err = consumeCh.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Panic(err)
	}

	if err := consumeCh.QueueBind(queueName, route, exchangeName, false, nil); err != nil {
		log.Panic(err)
	}

	go func() {
		d, err := consumeCh.Consume(queueName, "", false, false, false, false, nil)
		if err != nil {
			log.Panic(err)
		}

		for msg := range d {
			log.Printf("msg from streadway: %s", string(msg.Body))
			msg.Ack(true)
		}
	}()
}

func publishWabbit() {
	conn, err := reConnect.NewRabbitMqConfig(reConnect.Wabbit).Rabbitmq("amqp://127.0.0.1:5674")
	if err != nil {
		log.Panic(err)
	}

	sendCh := conn.Channel()

	// fmt.Println("sendCh", sendCh)
	if sendCh == nil {
		log.Println("sendCh", sendCh)
	}

	exchangeName := "test-exchange-wabbit"
	route := "test-route"
	// a := sendCh.Wabbit
	err = sendCh.Wabbit.ExchangeDeclare(exchangeName,
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
		log.Panic(err)
	}

	go func() {
		for {

			err = sendCh.Wabbit.Publish(
				exchangeName,
				route,
				[]byte(time.Now().String()),
				wabbit.Option{
					"contentType": "application/json",
				})

			log.Printf("publish wabit, err: %v", err)
			time.Sleep(5 * time.Second)
		}
	}()

}

func consumeWabbit() {
	exchangeName := "test-exchange-wabbit"
	route := "test-route"
	queueName := "test-queue-wabbit"
	conn, err := reConnect.NewRabbitMqConfig("wabit").Rabbitmq("amqp://127.0.0.1:5674")
	if err != nil {
		log.Panic(err)
	}
	consumeCh := conn.Channel()
	if consumeCh.Err != nil {
		log.Panic(err)
	}
	_, err = consumeCh.Wabbit.QueueDeclare(
		queueName, // name
		wabbit.Option{
			"delete":    false,
			"exclusive": false,
			"noWait":    false,
		}, // arguments
	)

	if err := consumeCh.Wabbit.QueueBind(queueName, route, exchangeName, wabbit.Option{
		"noWait": false,
	}); err != nil {
		log.Panic(err)
	}

	go func() {
		d, err := consumeCh.Wabbit.Consume(queueName, "", nil)
		if err != nil {
			log.Panic(err)
		}

		for msg := range d {
			log.Printf("msg wabit : %s", string(msg.Body()))
			msg.Ack(true)
		}
	}()
}
