package main

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

type SensorData struct {
    DeviceID      string  `json:"device_id"`
    Timestamp     string  `json:"timestamp"`
    SensorType    string  `json:"sensor_type"`
    ReadingType   string  `json:"reading_type"`
    DiscreteValue string  `json:"discrete_value"`
    NumericValue  float64 `json:"numeric_value"`
}

var ch *amqp.Channel
var db *sql.DB

func initPostgres() {
	var err error
	connStr := "host=db user=admin password=1234 dbname=pond_db sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Falha ao conectar no Postgres:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Postgres não respondeu:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sensor_data (
			id             SERIAL PRIMARY KEY,
			device_id      TEXT NOT NULL,
			timestamp      TIMESTAMPTZ NOT NULL,
			sensor_type    TEXT NOT NULL,
			reading_type   TEXT NOT NULL,
			discrete_value TEXT,
			numeric_value  DOUBLE PRECISION,
			created_at     TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatal("Falha ao criar tabela:", err)
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

func db_insert(body []byte) {
    var data SensorData
    if err := json.Unmarshal(body, &data); 
	err != nil {
        log.Printf("Erro ao deserializar mensagem: %v", err)
        return
    }

    _, err := db.Exec(`
        INSERT INTO sensor_data (device_id, timestamp, sensor_type, reading_type, discrete_value, numeric_value)
        VALUES ($1, $2, $3, $4, $5, $6)`,
        data.DeviceID,
        data.Timestamp,
        data.SensorType,
        data.ReadingType,
        data.DiscreteValue,
        data.NumericValue,
    )
    if err != nil {
        log.Printf("Erro ao inserir no banco: %v", err)
        return
    }

    log.Printf("Inserido no banco: %s", data.DeviceID)
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
    initPostgres()
	conn := initRabbitMQ()
	defer conn.Close()	

	msgs := queueConsumer()
	var forever chan struct{}
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			db_insert(d.Body)
		}
	}()
	<-forever
}