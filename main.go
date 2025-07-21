package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func echoTool(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	var input any = "No message provided"

	if args, ok := params.Arguments.(map[string]any); ok {
		if msg, exists := args["message"]; exists {
			input = msg
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Echo: %v", input),
			},
		},
	}, nil
}

func createLoggingMiddleware(logger *slog.Logger) mcp.Middleware[*mcp.ServerSession] {
	return func(next mcp.MethodHandler[*mcp.ServerSession]) mcp.MethodHandler[*mcp.ServerSession] {
		return func(ctx context.Context, session *mcp.ServerSession, method string, params mcp.Params) (mcp.Result, error) {
			logger.Info(
				"MCP method started",
				"method", method,
				"session_id", session.ID(),
				"has_params", params != nil,
			)

			start := time.Now()
			result, err := next(ctx, session, method, params)
			duration := time.Since(start)

			if err != nil {
				logger.Error(
					"MCP method failed",
					"method", method,
					"session_id", session.ID(),
					"duration_ms", duration.Milliseconds(),
					"err", err,
				)
			} else {
				logger.Info(
					"MCP method completed",
					"method", method,
					"session_id", session.ID(),
					"duration_ms", duration.Milliseconds(),
					"has_result", result != nil,
				)
			}

			return result, err
		}
	}
}

func createMCPServer(logger *slog.Logger) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "echo-server",
			Version: "1.0.0",
		},
		nil,
	)

	server.AddReceivingMiddleware(createLoggingMiddleware(logger))

	inputSchema := &jsonschema.Schema{}
	inputSchema.UnmarshalJSON([]byte(`{
		"type": "object",
		"properties": {
			"message": {
				"type": "string",
				"description": "The message to echo"
			}
		}
	}`))

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "echo",
			Description: "Echoes the input message",
			InputSchema: inputSchema,
		},
		echoTool,
	)

	return server
}

// createStatelessMCPHandler creates a truly stateless MCP handler using the new transport
func createStatelessMCPHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return NewStatelessStreamableHTTPHandler(func(request *http.Request) *mcp.Server {
		logger.Info("Creating new MCP server instance for stateless request", "method", request.Method, "path", request.URL.Path)
		return createMCPServer(logger)
	})
}

// Stateful implementation - reuses a single server instance across requests
func createStatefulMCPHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create server once and reuse it
	sharedServer := createMCPServer(logger)

	return mcp.NewStreamableHTTPHandler(
		func(request *http.Request) *mcp.Server {
			logger.Info("Reusing existing MCP server instance (stateful)", "request_id", request.Header.Get("X-Request-ID"))
			return sharedServer
		},
		nil,
	)
}

func main() {
	port := flag.String("port", "8080", "port to run the server on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/readiness", readinessHandler)
	mux.Handle("/mcp-stateful", createStatefulMCPHandler())
	mux.Handle("/mcp-stateless", createStatelessMCPHandler())

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	fmt.Printf("Server starting on port %s\n", *port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  /readiness - Health check\n")
	fmt.Printf("  /mcp-stateful - Stateful MCP server (shared instance)\n")
	fmt.Printf("  /mcp-stateless - Truly stateless MCP (custom transport, no sessions)\n")
	log.Fatal(server.ListenAndServe())
}
