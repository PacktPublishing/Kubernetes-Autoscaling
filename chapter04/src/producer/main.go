package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	RabbitMQURL   string
	QueueName     string
	MessageCount  int
	MessagePrefix string
}

func main() {
	config := loadConfig()
	conn, ch := setupRabbitMQ(config)
	defer conn.Close()
	defer ch.Close()

	q := declareQueue(ch, config.QueueName)
	sendMessages(ch, q, config)
}

func loadConfig() Config {
	return Config{
		RabbitMQURL:   getEnvOrFail("RABBITMQ_URL"),
		QueueName:     getEnvOrFail("QUEUE_NAME"),
		MessageCount:  getEnvAsIntOrFail("MESSAGE_COUNT"),
		MessagePrefix: getEnvOrDefault("MESSAGE_PREFIX", "Message"),
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

func getEnvAsIntOrFail(key string) int {
	value := getEnvOrFail(key)
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Invalid %s value: %s", key, err)
	}
	return intValue
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

func sendMessages(ch *amqp.Channel, q amqp.Queue, config Config) {
	for i := 1; i <= config.MessageCount; i++ {
		body := fmt.Sprintf("%s %d", config.MessagePrefix, i)
		err := ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf("Sent %s", body)
	}
	log.Printf("Sent %d messages", config.MessageCount)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}