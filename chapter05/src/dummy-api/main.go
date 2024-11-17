package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Response struct {
	Database struct {
		Metrics struct {
			WriteLatency int `json:"write_latency"`
		} `json:"metrics"`
	} `json:"database"`
}

func main() {
	http.HandleFunc("/", handleRequest)
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	dummyValue := getDummyValue()

	response := Response{}
	response.Database.Metrics.WriteLatency = dummyValue

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getDummyValue() int {
	dummyValueStr := os.Getenv("DUMMY_VALUE")
	if dummyValueStr == "" {
		log.Println("DUMMY_VALUE environment variable not set. Using default value 0.")
		return 0
	}

	dummyValue, err := strconv.Atoi(dummyValueStr)
	if err != nil {
		log.Printf("Error converting DUMMY_VALUE to integer: %v. Using default value 0.", err)
		return 0
	}

	return dummyValue
}