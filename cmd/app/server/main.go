package main

import (
	"db"
	"fmt"
	"io"
	"log"
	"middlewares"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	. "routes"

	. "helpers"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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

type App struct {
	certFile   string
	keyFile    string
	db         *db.PrismaClient
	grpcServer *grpc.Server
	logger     *zap.SugaredLogger
	gwmux      *runtime.ServeMux
	graphqlmux *GraphqlServeMux
}

func NewApp() (*App, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	certFile := "../../certs/server.crt"
	keyFile := "../../certs/server.key"
	sugar := logger.Sugar()
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		sugar.Fatalf("Failed to load TLS credentials: %v", err)
		return nil, err
	}

	// Initialize rate limiter with default trusted proxies
	trustedProxies := middlewares.DefaultTrustedProxies()
	sugar.Infof("Initializing rate limiter with trusted proxies: %v", trustedProxies)
	rateLimiter := middlewares.NewRateLimiter(5, 10, trustedProxies)

	// Create the gRPC server with TLS and middleware.
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(
			middlewares.ChainUnaryInterceptors(
				rateLimiter.RateLimiterInterceptor,
				middlewares.AuthUnaryInterceptor,
			),
		),
		grpc.StreamInterceptor(middlewares.AuthStreamInterceptor),
	)

	headerMatcher := func(key string) (string, bool) {
		key = strings.ToLower(key)
		if key == "authorization" {
			return "authorization", true // Return lowercase for consistency
		}
		return runtime.DefaultHeaderMatcher(key)
	}
	// For gRPC gateway
	gwmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(headerMatcher),
	)

	gwmuxGraphql := NewGraphqlServeMux()
	gwmuxGraphql.SetIncomingHeaderMatcher(headerMatcher)
	return &App{
		certFile:   certFile,
		keyFile:    keyFile,
		db:         db.NewClient(),
		grpcServer: grpcServer,
		logger:     sugar,
		gwmux:      gwmux,
		graphqlmux: gwmuxGraphql,
	}, nil
}

func (app *App) RegisterMux() fasthttp.RequestHandler {
	// fasthttp handler
	fasthttpHandler := fasthttpadaptor.NewFastHTTPHandler(wsproxy.WebsocketProxy(app.gwmux))

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
			graphqlHandler := middlewares.HeaderForwarderMiddleware(fasthttpadaptor.NewFastHTTPHandler(app.graphqlmux))
			graphqlHandler(ctx)
		default:
			fasthttpHandler(ctx) // Pass other requests to gRPC-Gateway
		}
	}))
	return fastMux
}

// running
func (app *App) Run() error {
	grpcPort := viper.GetString("grpc.port")
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}
	if err := app.db.Prisma.Connect(); err != nil {
		panic(err)
	}
	defer func() {
		if err := app.db.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()

	// Register gRPC services before starting the server.
	RegisterServers(app.grpcServer, app.db, app.logger)

	log.Println(fmt.Sprintf("Starting gRPC server on port %s", grpcPort))
	// Run gRPC server in a separate goroutine.
	go func() {
		if err := app.grpcServer.Serve(lis); err != nil {
			app.logger.Errorf("gRPC server stopped: %v", err)
		}
	}()

	clientCreds, err := credentials.NewClientTLSFromFile(app.certFile, "localhost")
	if err != nil {
		app.logger.Fatalf("Failed to load client TLS credentials: %v", err)
	}
	conn, err := grpc.Dial("localhost"+grpcPort, grpc.WithTransportCredentials(clientCreds))
	if err != nil {
		log.Fatalln("Failed to dial gRPC server:", err)
		return err
	}

	// Register gRPC-Gateway handlers.
	RegisterHandlers(app.gwmux, conn)
	RegisterGraphQLHandlers(app.graphqlmux.ServeMux, conn)
	// Convert the gRPC-Gateway mux to work with fasthttp.

	// Setup FastHTTP server.
	httpPort := viper.GetString("http.port")
	log.Println(fmt.Sprintf("Starting Thunder on port %s", httpPort))
	httpServer := &fasthttp.Server{
		Handler:      app.RegisterMux(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Logger:       &SilentLogger{}, // Use a silent logger to suppress output
	}
	log.Println("\033[32m✓ Server is running!\033[0m")

	// Run FastHTTP server in a separate goroutine.
	go func() {
		if err := httpServer.ListenAndServeTLS(httpPort, app.certFile, app.keyFile); err != nil {
			app.logger.Errorf("FastHTTP server stopped: %v", err)
		}
	}()

	// Listen for interrupt or termination signals.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	app.logger.Info("Shutdown signal received. Initiating graceful shutdown...")

	// Gracefully stop the gRPC server.
	app.grpcServer.GracefulStop()
	app.logger.Info("gRPC server gracefully stopped.")
	// Gracefully shutdown the FastHTTP server.
	if err := httpServer.Shutdown(); err != nil {
		app.logger.Errorf("Error shutting down FastHTTP server: %v", err)
	} else {
		app.logger.Info("FastHTTP server gracefully stopped.")
	}
	return nil
}

// main program
func main() {
	initConfig()
	initJaeger("grpc-gateway")
	app, err := NewApp()
	if err != nil {
		panic(err)
	}
	app.Run()
}
