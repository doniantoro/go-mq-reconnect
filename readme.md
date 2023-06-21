# Go-mq-reconnect - Library to reconnect rabbit mq

- [Description](#description)
- [Installation](#installation)
- [Usage](#usage)
  - [Making a simple reconnection rabbitmq nonssl](#making-a-simple-reconnection-rabbitmq-nonssl)
  - [Making a simple reconnection rabbitmq ssl](#making-a-simple-reconnection-rabbitmq-ssl)
  - [How to run example](#how-to-run-example)
- [Help and docs](#help-and-docs)
- [License](#license)

## Description

The problem face when using rabbitmq in golang is when rabbitmq down, with some reason it will happen.if that happen , we must restart our services that use rabbitmq , and also make our business stop temporary after the service restarted. Go-Mq-Reconnect is a Rabbitmq Recoonect for Golang that will make connection after rabbitmq down .whit this library , no worry rabbit mq restarted,because our service will re-connect to rabbit mq . This Library is based on [Wabbit](https://github.com/NeowayLabs/wabbit) , so after connect the rabbit-mq, the syntax will be same with wabbit . Go-Mq-Reconnect will can handle :

- Create reconect rabbitmq non ssl
- Create reconect rabbitmq ssl

## Installation

```
go get -u github.com/platogo/go-amqp-reconnect
```

Then Import it :

```
import reConnect "github.com/doniantoro/go-mq-reconnect"
```

## Usage

### Making a simple Reconnection rabbitmq nonssl

The below example will Connect to rabbitmq non-ssl:

```
// Create a new connection of non-ssl connection rabbit mq
conn, err := reConnect.NewRabbitMqConfig().Rabbitmq("amqp://127.0.0.1:5672")
if err != nil{
	panic(err)
}

```

### Making a simple Reconnection rabbitmq ssl

The below example will Connect to rabbitmq ssl,ssl key can generate on [Openssl](https://www.openssl.org/):

```
// Create a new connection of ssl connection rabbit mq
conn, err := reConnect.NewRabbitMqConfig().RabbitmqSsl("amqp://127.0.0.1:5673", "server.crt", "server.key", "ca.crt")
	if err != nil {
		log.Panic(err)
	}

```

### How to run example

to run example,you can run rabbitmq first, if dont have rabbit mq , can use my docker compose , with command :

- non-ssl

  ```
  //go to directory
  cd demo/non-ssl/

  //optional if you dont have rabbitmq
  docker-compose.yml up -d

  //if you run with your own rabbitmq , you need change port on demo/non-ssl/main.go ,
  // default usually 5672,then run the program with command
  go run demo/non-ssl/main.go
  ```

- ssl

  ```
  //go to directory
  cd demo/ssl/

  //optional if you dont have rabbitmq
  docker-compose.yml up -d

  //if you run with your own rabbitmq , you need change port on demo/ssl/main.go,
  // default usually 5672,then run the program with command
  go run demo/ssl/main.go
  ```

## Help and docs

We use GitHub issues only to discuss bugs and new features. For support please refer to:

- [Documentation](https://pkg.go.dev/github.com/doniantoro/go-mq-reconnect/v2)
- [Medium ( soon) ](https://)

## License

```

Copyright 2023, Doni Antoro

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

```

```

```
