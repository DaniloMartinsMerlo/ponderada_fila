package main

import (
    "database/sql"
    "encoding/json"
    "log"

    _ "github.com/lib/pq"
    "github.com/streadway/amqp"
)

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
		"data_request", 
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
		"data_request", 
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

func db_query(msg amqp.Delivery) {
    var request map[string]interface{}
    json.Unmarshal(msg.Body, &request)

    queryType := request["query_type"].(string)

    var rows *sql.Rows
    var err error

    switch queryType {
    case "all":
        rows, err = db.Query("SELECT * FROM sensor_data")
    case "discrete":
        rows, err = db.Query("SELECT * FROM sensor_data WHERE discrete_value != ''")
    case "analog":
        rows, err = db.Query("SELECT * FROM sensor_data WHERE numeric_value != 0")
    default:
        errMsg, _ := json.Marshal(map[string]string{"error": "query_type inválido"})
        ch.Publish("", msg.ReplyTo, false, false, amqp.Publishing{
            ContentType:   "application/json",
            CorrelationId: msg.CorrelationId,
            Body:          errMsg,
        })
        return
    }

    if err != nil {
        log.Printf("Erro na query: %v", err)
        return
    }
    defer rows.Close()

    cols, _ := rows.Columns()
    var resultado []map[string]interface{}

    for rows.Next() {
        row := make([]interface{}, len(cols))
        rowPtrs := make([]interface{}, len(cols))
        for i := range row {
            rowPtrs[i] = &row[i]
        }
        rows.Scan(rowPtrs...)

        m := make(map[string]interface{})
        for i, col := range cols {
            m[col] = row[i]
        }
        resultado = append(resultado, m)
    }

    body_answer, _ := json.Marshal(resultado)

    ch.Publish("", msg.ReplyTo, false, false, amqp.Publishing{
        ContentType:   "application/json",
        CorrelationId: msg.CorrelationId,
        Body:          body_answer,
    })
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
            db_query(d)
        }
    }()
    <-forever
}