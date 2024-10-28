package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	RabbitMQURL string
	QueueName   string
	BatchSize   int
}

func main() {
	config := loadConfig()
	conn, ch := setupRabbitMQ(config)
	defer conn.Close()
	defer ch.Close()

	q := declareQueue(ch, config.QueueName)
	setQoS(ch, config.BatchSize)

	for {
		processMessages(ch, q, config.BatchSize)
	}
}

func loadConfig() Config {
	rabbitMQURL := getEnvOrFail("RABBITMQ_URL")
	queueName := getEnvOrFail("QUEUE_NAME")
	batchSize := getEnvOrDefault("BATCH_SIZE", "10")

	return Config{
		RabbitMQURL: rabbitMQURL,
		QueueName:   queueName,
		BatchSize:   parseBatchSize(batchSize),
	}
}

func getEnvOrFail(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is not set", key)
	}
	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func parseBatchSize(batchSizeStr string) int {
	batchSize, err := strconv.Atoi(batchSizeStr)
	if err != nil {
		log.Fatalf("Invalid BATCH_SIZE value: %s", err)
	}
	if batchSize <= 0 {
		log.Fatal("BATCH_SIZE must be a positive integer")
	}
	return batchSize
}

func setupRabbitMQ(config Config) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(config.RabbitMQURL)
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	return conn, ch
}

func declareQueue(ch *amqp.Channel, queueName string) amqp.Queue {
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")
	return q
}

func setQoS(ch *amqp.Channel, prefetchCount int) {
	err := ch.Qos(
		prefetchCount, // prefetch count
		0,             // prefetch size
		false,         // global
	)
	failOnError(err, "Failed to set QoS")
}

func processMessages(ch *amqp.Channel, q amqp.Queue, batchSize int) {
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	messageCount := 0
	timeout := time.After(10 * time.Second)

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Channel closed")
				return
			}
			processMessage(msg)
			messageCount++
			if messageCount == batchSize {
				log.Printf("Processed %d messages", messageCount)
				return
			}
		case <-timeout:
			if messageCount == 0 {
				log.Println("No messages available. Waiting for 10 seconds before terminating.")
				time.Sleep(10 * time.Second)
				os.Exit(0)
			}
			log.Printf("Timeout: Processed %d messages", messageCount)
			return
		}
	}
}

func processMessage(msg amqp.Delivery) {
	log.Printf("Received a message: %s", msg.Body)
	sleepTime := time.Duration(rand.Intn(8)+3) * time.Second
	time.Sleep(sleepTime)
	msg.Ack(false)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}