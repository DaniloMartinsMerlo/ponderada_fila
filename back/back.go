package main

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

var ch *amqp.Channel

type SensorInput struct {
	DeviceID    string `json:"device_id" binding:"required"`
	Timestamp   string `json:"timestamp" binding:"required"`
	SensorType  string `json:"sensor_type" binding:"required"`
	ReadingType string `json:"reading_type" binding:"required"`
	Value       string `json:"value" binding:"required"`
}

type SensorOutput struct {
	DeviceID      string  `json:"device_id"`
	Timestamp     string  `json:"timestamp"`
	SensorType    string  `json:"sensor_type"`
	ReadingType   string  `json:"reading_type"`
	DiscreteValue string  `json:"discrete_value"`
	NumericValue  float64 `json:"numeric_value"`
}

func post_dados(c *gin.Context) {
	var input SensorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := processarSensor(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body, err := json.Marshal(output)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao serializar payload"})
		return
	}

	err = ch.Publish(
		"",
		"sensor_data",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body: body,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao publicar mensagem na fila"})
		return
	}

	c.JSON(http.StatusOK, output)

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

func main() {
	r := gin.Default()
	conn := initRabbitMQ()
	defer conn.Close()

	r.POST("/dados", post_dados)
	r.Run(":8088")
}