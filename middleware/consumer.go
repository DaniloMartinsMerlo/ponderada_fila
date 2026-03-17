package main

import (
	"log"
	"github.com/streadway/amqp"
)

var ch *amqp.Channel

type RabbitTemplate struct {
	DeviceID      string  `json:"device_id"`
	Timestamp     string  `json:"timestamp"`
	SensorType    string  `json:"sensor_type"`
	ReadingType   string  `json:"reading_type"`
	DiscreteValue string  `json:"discrete_value"`
	NumericValue  float64 `json:"numeric_value"`
}

func initRabbitMQ() *amqp.Connection {
	conn, err := amqp.Dial("amqp://admin:admin@rabbitmq:5672/")
	if err != nil {
		log.Fatal("Falha ao conectar no RabbitMQ:", err)
	}

	ch, err = conn.Channel()
	if err != nil {
		log.Fatal("Falha ao abrir canal:", err)
	}

	_, err = ch.QueueDeclare(
		"sensor_data", 
		true,       
		false,      
		false,      
		false,      
		nil,
	)
	if err != nil {
		log.Fatal("Falha ao declarar fila:", err)
	}
	return conn
}


func queueConsumer() <-chan amqp.Delivery {
	msgs, err := ch.Consume(
		"sensor_data", 
		"",     
		true,   
		false,  
		false,  
		false,  
		nil,
	)	
	if err != nil {
		log.Fatal("Falha ao consumir mensagens:", err)
	}
	return msgs
}

func main() {
	conn := initRabbitMQ()
	defer conn.Close()	

	msgs := queueConsumer()
	var forever chan struct{}
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}