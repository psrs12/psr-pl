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
	"github.com/aws/aws-sdk-go-v2/service/sfn"

	"workflow-status-service/internal/workflowstatus"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	appManagementBaseURL := requireEnv(logger, "APPLICATION_MANAGEMENT_BASE_URL")
	port := envOrDefault("PORT", "8086")

	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("loading AWS config", "error", err)
		os.Exit(1)
	}

	var sfnOpts []func(*sfn.Options)
	if endpoint := os.Getenv("STEPFUNCTIONS_ENDPOINT_URL"); endpoint != "" {
		sfnOpts = append(sfnOpts, func(o *sfn.Options) { o.BaseEndpoint = &endpoint })
	}
	sfnClient := sfn.NewFromConfig(awsCfg, sfnOpts...)

	applications := workflowstatus.NewApplicationManagementClient(appManagementBaseURL, nil)
	executions := workflowstatus.NewSFNExecutionReader(sfnClient)
	service := workflowstatus.NewService(applications, executions)
	handler := workflowstatus.NewHandler(service, logger)

	mux := http.NewServeMux()
	handler.Register(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           withCORS(mux, envOrDefault("CORS_ALLOWED_ORIGIN", "*")),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting workflow-status-service", "port", port)
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

func withCORS(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
