package main

import (
	"encoding/json"
	//	"fmt"
	"github.com/gorilla/mux"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"
	"github.com/openzipkin/zipkin-go/model"
	reporterhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

var (
	logger    = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	port      = getEnv("PORT", "8888")                                 //flag.String("p", "8888", "server port")
	zipkinUrl = getEnv("ZIPKIN", "http://localhost:9411/api/v2/spans") //flag.String("s", "127.0.0.1", "server address")
)

type Discount struct {
	Category string
	Value    int
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func newTracer() (*zipkin.Tracer, error) {
	// The reporter sends traces to zipkin server
	reporter := reporterhttp.NewReporter(zipkinUrl)
	port64, err := strconv.ParseUint(port, 16, 16)
	// Local endpoint represent the local service information
	localEndpoint := &model.Endpoint{ServiceName: "discountService", Port: uint16(port64)}
	// Sampler tells you which traces are going to be sampled or not. In this case we will record 100% (1.00) of traces.
	sampler, err := zipkin.NewCountingSampler(1)
	if err != nil {
		return nil, err
	}
	t, err := zipkin.NewTracer(
		reporter,
		zipkin.WithSampler(sampler),
		zipkin.WithLocalEndpoint(localEndpoint),
	)
	if err != nil {
		return nil, err
	}

	return t, err
}

func discountHandler(w http.ResponseWriter, r *http.Request) {
	//random test fault
	if rand.Intn(100) <= 10 {
		w.WriteHeader(500)
		return
	}

	params := mux.Vars(r)
	cat := params["cat"]
	discount := Discount{Category: cat, Value: categoryToDiscount(cat)}
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(discount); err != nil {
		w.WriteHeader(500)
	}
}

func categoryToDiscount(category string) int {
	switch category {
	case "platinum":
		return 20
	case "gold":
		return 10
	case "silver":
		return 5
	default:
		return 0
	}
}

func main() {
	tracer, err := newTracer()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/category/{cat}", discountHandler)

	router.Use(zipkinhttp.NewServerMiddleware(
		tracer,
		zipkinhttp.SpanName("request")), // name for request span
	)

	log.Fatal(http.ListenAndServe(":"+port, router))

}
