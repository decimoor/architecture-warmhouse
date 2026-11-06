package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type temperatureResponse struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

type temperatureGenerator struct {
	mu       sync.Mutex
	random   *rand.Rand
	previous map[string]float64
}

func newTemperatureGenerator() *temperatureGenerator {
	return &temperatureGenerator{
		random:   rand.New(rand.NewSource(time.Now().UnixNano())),
		previous: make(map[string]float64),
	}
}

func (g *temperatureGenerator) next(key string) float64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	value := math.Round((18+g.random.Float64()*10)*10) / 10
	if previous, ok := g.previous[key]; ok && value == previous {
		value += 0.1
		if value > 28 {
			value = 18
		}
		value = math.Round(value*10) / 10
	}
	g.previous[key] = value
	return value
}

type server struct {
	temperatures *temperatureGenerator
}

func (s *server) handleTemperature(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	s.writeTemperature(w, r.URL.Query().Get("location"), "")
}

func (s *server) handleTemperatureByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	sensorID := strings.TrimPrefix(r.URL.Path, "/temperature/")
	if sensorID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sensor ID is required"})
		return
	}
	s.writeTemperature(w, "", sensorID)
}

func (s *server) writeTemperature(w http.ResponseWriter, location, sensorID string) {
	if location == "" {
		switch sensorID {
		case "1":
			location = "Living Room"
		case "2":
			location = "Bedroom"
		case "3":
			location = "Kitchen"
		default:
			location = "Unknown"
		}
	}

	if sensorID == "" {
		switch location {
		case "Living Room":
			sensorID = "1"
		case "Bedroom":
			sensorID = "2"
		case "Kitchen":
			sensorID = "3"
		default:
			sensorID = "0"
		}
	}

	response := temperatureResponse{
		Value:       s.temperatures.next(sensorID + "|" + location),
		Unit:        "°C",
		Timestamp:   time.Now().UTC(),
		Location:    location,
		Status:      "active",
		SensorID:    sensorID,
		SensorType:  "temperature",
		Description: "Simulated temperature reading",
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	temperatureServer := &server{temperatures: newTemperatureGenerator()}
	mux := http.NewServeMux()
	mux.HandleFunc("/temperature", temperatureServer.handleTemperature)
	mux.HandleFunc("/temperature/", temperatureServer.handleTemperatureByID)

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("temperature-api listening on port %s", port)
	log.Fatal(httpServer.ListenAndServe())
}
