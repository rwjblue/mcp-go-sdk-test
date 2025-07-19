package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

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

func mcpHandler(w http.ResponseWriter, r *http.Request) {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "echo-server",
			Version: "1.0.0",
		},
		nil,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "echo",
			Description: "Echoes the input message",
			InputSchema: nil,
		},
		echoTool,
	)

	handler := mcp.NewStreamableHTTPHandler(
		func(request *http.Request) *mcp.Server {
			return server
		},
		nil,
	)

	handler.ServeHTTP(w, r)
}

func main() {
	port := flag.String("port", "8080", "port to run the server on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/readiness", readinessHandler)
	mux.HandleFunc("/mcp", mcpHandler)

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	fmt.Printf("Server starting on port %s\n", *port)
	log.Fatal(server.ListenAndServe())
}

