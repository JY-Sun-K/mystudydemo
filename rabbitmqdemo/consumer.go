package main

import (
	"grpcdemo/rabbitmqdemo/models"
	RabbitMQ "grpcdemo/rabbitmqdemo/rabbitmq"
)

func main() {
	messages:=models.NewMessageSum()

	rabbitmqConsumeSimple := RabbitMQ.NewRabbitMQSimple("penguin")
	rabbitmqConsumeSimple.ConsumeSimple(messages)


}
