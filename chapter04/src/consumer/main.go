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
	RabbitMQURL    string
	QueueName      string
	BatchSize      int
	MaxSleepTime   int
	TimeoutDuration time.Duration
}

func main() {
	config := loadConfig()
	conn, ch := setupRabbitMQ(config)
	defer conn.Close()
	defer ch.Close()

	q := declareQueue(ch, config.QueueName)
	processMessages(ch, q, config)
}

func loadConfig() Config {
	return Config{
		RabbitMQURL:    getEnvOrFail("RABBITMQ_URL"),
		QueueName:      getEnvOrFail("QUEUE_NAME"),
		BatchSize:      getEnvAsIntOrFail("BATCH_SIZE"),
		MaxSleepTime:   getEnvAsIntOrDefault("MAX_SLEEP_TIME", 5),
		TimeoutDuration: time.Duration(getEnvAsIntOrDefault("TIMEOUT_DURATION", 30)) * time.Second,
	}
}

func getEnvOrFail(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is not set", key)
	}
	return value
}

func getEnvAsIntOrFail(key string) int {
	value := getEnvOrFail(key)
	return parseIntOrFail(key, value)
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return parseIntOrFail(key, value)
}

func parseIntOrFail(key, value string) int {
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

func processMessages(ch *amqp.Channel, q amqp.Queue, config Config) {
	err := ch.Qos(
		config.BatchSize, // prefetch count
		0,                // prefetch size
		false,            // global
	)
	failOnError(err, "Failed to set QoS")

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

	processedCount := 0
	lastMessageTime := time.Now()

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Channel closed")
				return
			}
			log.Printf("Received a message: %s", msg.Body)
			
			// Sleep for a random time between 3 and MaxSleepTime seconds
			sleepTime := time.Duration(rand.Intn(config.MaxSleepTime-2)+3) * time.Second
			log.Printf("Sleeping for %v", sleepTime)
			time.Sleep(sleepTime)
			
			err := msg.Ack(false)
			if err != nil {
				log.Printf("Error acknowledging message: %s", err)
			}
			
			processedCount++
			lastMessageTime = time.Now()
			log.Printf("Processed message %d", processedCount)

		default:
			// Check if we've exceeded the timeout duration
			if time.Since(lastMessageTime) > config.TimeoutDuration {
				log.Printf("No new messages for %v. Processed %d messages. Exiting.", config.TimeoutDuration, processedCount)
				return
			}
			// Sleep briefly to prevent tight looping
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
}