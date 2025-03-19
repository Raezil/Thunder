package main

import (
	"db"
	. "helpers"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"middlewares"
	. "routes"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
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

// initJaeger initializes a Jaeger tracer.
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

	// Load TLS credentials for gRPC.
	certFile := "../../certs/server.crt"
	keyFile := "../../certs/server.key"
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		sugar.Fatalf("Failed to load TLS credentials: %v", err)
	}

	// Listen on the gRPC port.
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

	// Initialize rate limiter.
	rateLimiter := middlewares.NewRateLimiter(5, 10)

	// Create the gRPC server with TLS and middleware.
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(
			middlewares.ChainUnaryInterceptors(
				rateLimiter.RateLimiterInterceptor,
				middlewares.AuthUnaryInterceptor,
			),
		),
	)
	RegisterServers(grpcServer, client, sugar)

	// Register gRPC Health service.
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	sugar.Infof("Serving gRPC with TLS on 0.0.0.0%s", grpcPort)
	// Run gRPC server in a separate goroutine.
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			sugar.Errorf("gRPC server stopped: %v", err)
		}
	}()

	// Setup secure connection for gRPC-Gateway.
	clientCreds, err := credentials.NewClientTLSFromFile(certFile, "localhost")
	if err != nil {
		sugar.Fatalf("Failed to load client TLS credentials: %v", err)
	}
	conn, err := grpc.Dial("localhost"+grpcPort, grpc.WithTransportCredentials(clientCreds))
	if err != nil {
		log.Fatalln("Failed to dial gRPC server:", err)
	}

	// In your main.go, ensure both handlers use the same header matcher
	headerMatcher := func(key string) (string, bool) {
		switch strings.ToLower(key) {
		case "authorization":
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}

	// For gRPC gateway
	gwmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(headerMatcher),
	)

	// For GraphQL
	gwmuxGraphql := NewGraphqlServeMux()
	gwmuxGraphql.SetIncomingHeaderMatcher(headerMatcher)

	RegisterGraphQLHandlers(gwmuxGraphql.ServeMux, conn)
	RegisterHandlers(gwmux, conn)

	// Convert the gRPC-Gateway mux to work with fasthttp.
	fasthttpHandler := fasthttpadaptor.NewFastHTTPHandler(gwmux)

	// Define FastHTTP handlers.
	healthCheckHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBody([]byte("OK"))
	}

	readyCheckHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBody([]byte("Ready"))
	}

	// Create a FastHTTP router.
	fastMux := middlewares.CORSMiddleware(middlewares.LoggingMiddleware(func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/health":
			healthCheckHandler(ctx)
		case "/ready":
			readyCheckHandler(ctx)
		case "/graphql":
			graphqlHandler := middlewares.HeaderForwarderMiddleware(fasthttpadaptor.NewFastHTTPHandler(gwmuxGraphql))
			graphqlHandler(ctx)
		default:
			fasthttpHandler(ctx) // Pass other requests to gRPC-Gateway
		}
	}))

	// Setup FastHTTP server.
	httpPort := viper.GetString("http.port")
	sugar.Infof("Serving gRPC-Gateway with FastHTTP on https://0.0.0.0%s", httpPort)
	httpServer := &fasthttp.Server{
		Handler: fastMux,
	}

	// Run FastHTTP server in a separate goroutine.
	go func() {
		if err := httpServer.ListenAndServeTLS(httpPort, certFile, keyFile); err != nil {
			sugar.Errorf("FastHTTP server stopped: %v", err)
		}
	}()

	// Listen for interrupt or termination signals.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	sugar.Info("Shutdown signal received. Initiating graceful shutdown...")

	// Gracefully stop the gRPC server.
	grpcServer.GracefulStop()
	sugar.Info("gRPC server gracefully stopped.")

	// Gracefully shutdown the FastHTTP server.
	if err := httpServer.Shutdown(); err != nil {
		sugar.Errorf("Error shutting down FastHTTP server: %v", err)
	} else {
		sugar.Info("Thunder server gracefully stopped.")
	}
}
