package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func get_dados(c *gin.Context) {
	queryType := c.DefaultQuery("query_type", "all")

	replyQueue, err := ch.QueueDeclare(
		"",    
		false, 
		true,  
		true,  
		false,
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao criar fila de resposta"})
		return
	}

	replies, err := ch.Consume(
		replyQueue.Name,
		"",
		true,  
		true,  
		false,
		false,
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao consumir fila de resposta"})
		return
	}

	correlationID := uuid.New().String()

	payload, _ := json.Marshal(gin.H{"query_type": queryType})

	err = ch.Publish("", "data_request", false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: correlationID,
		ReplyTo:       replyQueue.Name,
		Body:          payload,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao publicar requisição"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case msg := <-replies:
			if msg.CorrelationId == correlationID {
				var result interface{}
				json.Unmarshal(msg.Body, &result)
				c.JSON(http.StatusOK, result)
				return
			}
		case <-ctx.Done():
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "timeout aguardando resposta"})
			return
		}
	}
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

	queues := []string{"sensor_data", "data_request"}
	for _, queueName := range queues {
		_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			log.Fatalf("Falha ao declarar fila %s: %v", queueName, err)
		}
	}

	return conn
}

func main() {
	r := gin.Default()

	conn := initRabbitMQ()
	defer conn.Close()

	r.POST("/dados", post_dados)
	r.GET("/dados", get_dados)

	r.Run(":8088")
}