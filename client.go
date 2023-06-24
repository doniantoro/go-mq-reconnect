package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/NeowayLabs/wabbit"
	"github.com/NeowayLabs/wabbit/amqp"

	streadway "github.com/streadway/amqp"
)

var (
	Wabbit    = "wabit"
	Streadway = "streadway"
	NotMatch  = "error : conntype is not match"
)

func NewRabbitMqConfig(connType string) *Connection {
	return &Connection{connType: connType}
}

// Connection amqp.Connection wrapper

// Channel wabbit.Channel wapper
type Channel struct {
	streadway.Channel
	Channel2 interface{}
	closed   int32
	Err      error
	connType string
}

type Connection struct {
	connType  string
	wabit     wabbit.Conn
	streadway *streadway.Connection
}

// This function is function to reconnect rabbitmq non ssl , this function need parameter rabbit mq host
// This function owned by  Connection,and will return Connection it self. before call this function , need call NewRabbitMqConfig
// This function will retry to reconnect every 3 seconds
func (c *Connection) Rabbitmq(uri string) (*Connection, error) {
	var connection *Connection
	if c.connType == Wabbit {

		var conn wabbit.Conn
		var err error
		conn, err = amqp.Dial(uri)
		if err != nil {

			log.Println("connection closed")
			return nil, err
		}
		connection = &Connection{wabit: conn, connType: c.connType}

		go func() {
			for {
				reason, ok := <-connection.wabit.NotifyClose(make(chan wabbit.Error))
				// exit this goroutine if closed by developer
				if !ok {
					log.Println("connection closed")
					break
				}
				log.Printf("connection closed, reason: %v", reason)

				// reconnect if not closed by developer
				for {
					// wait 3s for reconnect
					time.Sleep(3 * time.Second)

					conn, err = amqp.Dial(uri)
					if err == nil {
						connection.wabit = conn
						log.Println("reconnect success")
						break
					}

					log.Println("reconnect failed, err: ", err)
				}
			}
		}()
	} else {
		var conn *streadway.Connection
		var err error
		conn, err = streadway.Dial(uri)
		if err != nil {

			log.Println("connection closed")
			return nil, err
		}
		connection = &Connection{streadway: conn, connType: c.connType}

		go func() {
			for {
				reason, ok := <-connection.streadway.NotifyClose(make(chan *streadway.Error))
				// exit this goroutine if closed by developer
				if !ok {
					log.Println("connection closed")
					break
				}
				log.Printf("connection closed, reason: %v", reason)

				// reconnect if not closed by developer
				for {
					// wait 3s for reconnect
					time.Sleep(3 * time.Second)

					conn, err = streadway.Dial(uri)
					if err == nil {
						connection.streadway = conn
						log.Println("reconnect success")
						break
					}

					log.Println("reconnect failed, err: ", err)
				}
			}
		}()
	}
	return connection, nil

}

// This function is function to reconnect rabbitmq ssl , this function need some parameter :
// - Uri = rabbit mq host

// - certFile = public key of ssl certificate

// - keyFile = private key of ssl certificate

// - caCert = caCert of ssl certificate

// - serverName= server name that will config in tls

// This function owned by  Connection,and will return Connection it self. before call this function , need call NewRabbitMqConfig
// This function will retry to reconnect every 3 seconds
func (c *Connection) RabbitmqSsl(uri, certFile, keyFile, caCert, serverName string) (*Connection, error) {

	caCertByte, err := os.ReadFile(caCert)

	if err != nil {
		log.Println("failed to read file ", caCert, ": with error err", err)
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Println("failed to load key pair , with error", err)
		return nil, err
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(caCertByte)

	tlsConf := &tls.Config{
		RootCAs:      rootCAs,
		Certificates: []tls.Certificate{cert},
		ServerName:   serverName, // Optional
	}

	conn, err := amqp.DialTLS(uri, tlsConf)
	if err != nil {

		log.Println("failed dial tls , with error", err)
		return nil, err
	}
	connection := &Connection{
		wabit:    conn,
		connType: c.connType,
	}

	go func() {
		for {
			reason, ok := <-connection.wabit.NotifyClose(make(chan wabbit.Error))
			// exit this goroutine if closed by developer
			if !ok {
				log.Println("connection tls closed")
				break
			}
			log.Printf("connection tls closed, reason: %v", reason)

			// reconnect if not closed by developer
			for {
				// wait 3s for reconnect
				time.Sleep(3 * time.Second)

				conn, err := amqp.DialTLS(uri, tlsConf)
				if err == nil {
					connection.wabit = conn
					log.Println("reconnect tls success")
					break
				} else {
					log.Println("reconnect tls failed, err: ", err)

				}
			}
		}
	}()

	log.Println("RabbitMQ accepted connection using SSL")
	return connection, nil

}

// This function is function to re-create channel , after reconnect connection to rabbitmq
// This function need to be called to re-create channel
// This function owned by  Connection,and will return Connection it self. before call this function , need call NewRabbitMqConfig
// This function will retry to reconnect every 3 seconds
func (c *Connection) Channel() *Channel {
	var channel Channel
	if c.connType == Wabbit {

		ch, err := c.wabit.Channel()
		if err != nil {

			log.Println("failed to close channel")
			return &Channel{
				Err: err,
			}
		}

		channel.Channel2 = ch

		go func() {
			for {
				// ch2 := channel.Channel2.(wabbit.Channel)
				reason, ok := <-channel.Channel2.(wabbit.Channel).NotifyClose(make(chan wabbit.Error))
				// exit this goroutine if closed by developer
				if !ok || channel.IsClosed() {
					log.Println("channel closed")
					channel.Channel2.(wabbit.Channel).Close() // close again, ensure closed flag set when connection closed
					break
				}
				log.Println("channel closed, reason: ", reason)

				// reconnect if not closed by developer
				for {
					// wait 3s for connection reconnect
					time.Sleep(time.Duration(3) * time.Second)

					ch, err := c.wabit.Channel()
					if err == nil {
						log.Println("channel recreate success")
						channel.Channel2 = ch
						break
					}

					log.Println("channel recreate failed, err:", err)
				}
			}

		}()
	} else {
		ch, err := c.streadway.Channel()
		if err != nil {

			log.Println("failed to close channel")
			return &Channel{
				Err: err,
			}
		}

		channel.Channel2 = *ch

		go func() {
			for {
				reason, ok := <-channel.Channel.NotifyClose(make(chan *streadway.Error))
				// exit this goroutine if closed by developer
				if !ok || channel.IsClosed() {
					log.Println("channel closed")
					channel.Channel.Close() // close again, ensure closed flag set when connection closed
					break
				}
				log.Println("channel closed, reason: ", reason)

				// reconnect if not closed by developer
				for {
					// wait 3s for connection reconnect
					time.Sleep(time.Duration(3) * time.Second)

					ch, err := c.streadway.Channel()
					if err == nil {
						log.Println("channel recreate success")
						channel.Channel2 = *ch
						break
					}

					log.Println("channel recreate failed, err:", err)
				}
			}

		}()
	}
	channel.connType = c.connType
	return &channel
}

// This function is function to Close connection
func (ch *Channel) Wabbitv2() (wabbit.Channel, error) {
	if ch.connType != Wabbit {
		return nil, errors.New(NotMatch)
	}
	return ch.Channel2.(wabbit.Channel), nil
}

// This function is function to Close connection
func (ch *Channel) IsClosed() bool {
	return (atomic.LoadInt32(&ch.closed) == 1)
}
