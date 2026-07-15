package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sfn"

	"application-management-service/internal/application"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	tableName := requireEnv(logger, "APPLICATION_TABLE_NAME")
	offerLogURL := requireEnv(logger, "OFFERLOG_BASE_URL")
	stateMachineARN := requireEnv(logger, "PRICING_STATE_MACHINE_ARN")
	port := envOrDefault("PORT", "8081")

	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("loading AWS config", "error", err)
		os.Exit(1)
	}

	var dynamoOpts []func(*dynamodb.Options)
	if endpoint := os.Getenv("DYNAMODB_ENDPOINT_URL"); endpoint != "" {
		dynamoOpts = append(dynamoOpts, func(o *dynamodb.Options) { o.BaseEndpoint = &endpoint })
	}
	dynamoClient := dynamodb.NewFromConfig(awsCfg, dynamoOpts...)

	var sfnOpts []func(*sfn.Options)
	if endpoint := os.Getenv("STEPFUNCTIONS_ENDPOINT_URL"); endpoint != "" {
		sfnOpts = append(sfnOpts, func(o *sfn.Options) { o.BaseEndpoint = &endpoint })
	}
	sfnClient := sfn.NewFromConfig(awsCfg, sfnOpts...)

	repo := application.NewDynamoRepository(dynamoClient, tableName)
	offerClient := application.NewOfferLogClient(offerLogURL, nil)
	sessions := application.NewDynamoSessionStore(dynamoClient, tableName)
	workflow := application.NewSFNWorkflowStarter(sfnClient, stateMachineARN)
	service := application.NewService(repo, offerClient, sessions, workflow)
	handler := application.NewHandler(service, logger)

	mux := http.NewServeMux()
	handler.Register(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           application.WithRequestID(withCORS(mux, envOrDefault("CORS_ALLOWED_ORIGIN", "*"))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting application-management-service", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}

func requireEnv(logger *slog.Logger, key string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.Error("missing required environment variable", "key", key)
		os.Exit(1)
	}
	return value
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// withCORS lets the UI dev server (a different origin in local
// development — different port on localhost) call this API. In
// production the UI and API share an origin (see nginx.conf), so this
// only matters for local dev, but it's harmless either way.
func withCORS(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Channel-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
