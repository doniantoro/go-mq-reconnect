package config

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/NeowayLabs/wabbit"
	"github.com/NeowayLabs/wabbit/amqp"
)

func NewRabbitMqConfig() *Connection {
	return &Connection{}
}

type Config struct {
	Connection Connection
	Channel    Channel
}

// Connection amqp.Connection wrapper
type Connection struct {
	wabbit.Conn
}

// Channel amqp.Channel wapper
type Channel struct {
	wabbit.Channel
	closed int32
}

// This function is function to reconnect rabbitmq non ssl , this function need parameter rabbit mq host
// This function owned by  Connection,and will return Connection it self. before call this function , need call NewRabbitMqConfig
// This function will retry to reconnect every 3 seconds
func (c *Connection) Rabbitmq(uri string) (*Connection, error) {

	var conn wabbit.Conn
	var err error
	conn, err = amqp.Dial(uri)
	if err != nil {

		log.Println("connection closed")
		return nil, err
	}
	connection := &Connection{Conn: conn}

	go func() {
		for {
			reason, ok := <-connection.NotifyClose(make(chan wabbit.Error))
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
					connection.Conn = conn
					log.Println("reconnect success")
					break
				}

				log.Println("reconnect failed, err: ", err)
			}
		}
	}()

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
		Conn: conn,
	}

	go func() {
		for {
			reason, ok := <-connection.Conn.NotifyClose(make(chan wabbit.Error))
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
					connection.Conn = conn
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
func (c *Connection) Channel() (*Channel, error) {
	ch, err := c.Conn.Channel()
	if err != nil {

		log.Println("failed to close channel")
		return nil, err
	}

	channel := &Channel{
		Channel: ch,
	}

	go func() {
		for {
			reason, ok := <-channel.Channel.NotifyClose(make(chan wabbit.Error))
			// exit this goroutine if closed by developer
			if !ok || channel.IsClosed() {
				log.Println("channel closed")
				channel.Close() // close again, ensure closed flag set when connection closed
				break
			}
			log.Println("channel closed, reason: ", reason)

			// reconnect if not closed by developer
			for {
				// wait 3s for connection reconnect
				time.Sleep(time.Duration(3) * time.Second)

				ch, err := c.Conn.Channel()
				if err == nil {
					log.Println("channel recreate success")
					channel.Channel = ch
					break
				}

				log.Println("channel recreate failed, err:", err)
			}
		}

	}()

	return channel, nil
}

// IsClosed indicate closed by developer
func (ch *Channel) IsClosed() bool {
	return (atomic.LoadInt32(&ch.closed) == 1)
}
