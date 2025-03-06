package main

import (
	"db"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"middlewares"
	. "routes"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// initConfig sets default values and loads environment variables.
func initConfig() {
	viper.SetDefault("grpc.port", ":50051")
	viper.SetDefault("http.port", ":8080")
	viper.AutomaticEnv()
}

// initJaeger initializes a Jaeger tracer based on environment configuration.
func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatalf("Failed to read Jaeger env vars: %v", err)
	}
	cfg.ServiceName = service
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		log.Fatalf("Could not initialize Jaeger tracer: %v", err)
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer
}

func main() {
	// Initialize configuration.
	initConfig()

	// Setup structured logging.
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	sugar := logger.Sugar()
	defer logger.Sync()

	// Initialize Jaeger tracer.
	_, closer := initJaeger("thunder-grpc")
	defer closer.Close()

	// Load TLS credentials for the gRPC server.
	certFile := "../certs/server.crt"
	keyFile := "../certs/server.key"
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		sugar.Fatalf("Failed to load TLS credentials: %v", err)
	}

	// Listen on the configured gRPC port.
	grpcPort := viper.GetString("grpc.port")
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Connect to the database.
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()

	// Initialize rate limiter (e.g., 5 requests per second, burst of 10).
	rateLimiter := middlewares.NewRateLimiter(5, 10)

	// Create the gRPC server with TLS and custom interceptors.
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(
			middlewares.ChainUnaryInterceptors(
				rateLimiter.RateLimiterInterceptor, // Rate limiting
				middlewares.AuthUnaryInterceptor,   // Authentication
			),
		),
	)
	RegisterServers(grpcServer, client, sugar)

	// Register gRPC Health service.
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	sugar.Infof("Serving gRPC with TLS on 0.0.0.0%s", grpcPort)
	go func() {
		log.Fatalln(grpcServer.Serve(lis))
	}()

	// Setup secure connection for gRPC-Gateway.
	clientCreds, err := credentials.NewClientTLSFromFile(certFile, "localhost")
	if err != nil {
		sugar.Fatalf("Failed to load client TLS credentials: %v", err)
	}
	conn, err := grpc.Dial(
		"localhost"+grpcPort,
		grpc.WithTransportCredentials(clientCreds),
	)
	if err != nil {
		log.Fatalln("Failed to dial gRPC server:", err)
	}

	// Register gRPC-Gateway handlers.
	gwmux := runtime.NewServeMux()
	RegisterHandlers(gwmux, conn)

	// Create a new HTTP mux and add health and readiness endpoints.
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// You can add more logic here if needed.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// You might include checks (e.g., database connectivity) before reporting readiness.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})
	mux.Handle("/", gwmux)

	httpPort := viper.GetString("http.port")
	gwServer := &http.Server{
		Addr:              httpPort,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // Timeout for reading request headers
	}

	sugar.Infof("Serving gRPC-Gateway on https://0.0.0.0%s", httpPort)
	log.Fatalln(gwServer.ListenAndServeTLS(certFile, keyFile))
}
