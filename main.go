package main

import (
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"strconv"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	}))
	// Get some values from local environments
	targetURL := os.Getenv("TARGET_URL")
	requestRate, convRequestRateErr := strconv.Atoi(os.Getenv("REQUEST_RATE"))
	if convRequestRateErr != nil {
		fmt.Fprintf(os.Stderr, "failed to convert request rate to integer %s: %s", os.Getenv("REQUEST_RATE"), convRequestRateErr)
		os.Exit(1)
	}
	requestDuration, convDurationErr := strconv.Atoi(os.Getenv("REQUEST_DURATION"))
	if convDurationErr != nil {
		fmt.Fprintf(os.Stderr, "failed to convert request duration to integer %s: %s", os.Getenv("REQUEST_DURATION"), convDurationErr)
		os.Exit(1)
	}
	requestBodyFilePath := os.Getenv("REQUEST_BODY_FILE_PATH")

	rate := vegeta.Rate{
		Freq: requestRate,
		Per:  time.Second,
	}

	body, err := ioutil.ReadFile(requestBodyFilePath)
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	duration := time.Duration(requestDuration) * time.Second

	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "POST",
		URL:    targetURL,
		Body:   body,
		Header: map[string][]string{
			"Content-Type":  {"application/json"},
			"Authorization": {"Bearer " + os.Getenv("TOKEN")},
		},
	})
	attacker := vegeta.NewAttacker()

	logger.With("rate", requestRate).With("duration", requestDuration).Info("Starting vegeta attacker")

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Attack") {
		metrics.Add(res)
	}

	metrics.Close()

	latencies := map[string]string{
		//"Total": metrics.Latencies.Total.String(),
		"Mean": metrics.Latencies.Mean.String(),
		"P50":  metrics.Latencies.P50.String(),
		"P90":  metrics.Latencies.P90.String(),
		"P95":  metrics.Latencies.P95.String(),
		"P99":  metrics.Latencies.P99.String(),
		"Max":  metrics.Latencies.Max.String(),
		"Min":  metrics.Latencies.Min.String(),
	}

	logger.Info("Finished vegeta attacker",
		slog.Any("latencies", latencies),
		slog.Uint64("requests", metrics.Requests),
		slog.Float64("request_rate", metrics.Rate),
		slog.Float64("request_throughput", metrics.Throughput),
		slog.Float64("success[%]", metrics.Success),
		slog.Any("status_codes with count", metrics.StatusCodes),
		slog.Any("errors", metrics.Errors),
	)
}
