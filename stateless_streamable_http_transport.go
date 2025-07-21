package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewStatelessStreamableHTTPHandler creates an HTTP handler for stateless MCP interactions.
// Unlike the standard StreamableHTTP handler, this version:
// - Does not create or manage sessions
// - Responds with plain JSON instead of SSE streams
// - Processes each request independently
// - Does not include Mcp-Session-Id headers
func NewStatelessStreamableHTTPHandler(getServer func(*http.Request) *mcp.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests for stateless operation
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed - only POST supported for stateless MCP", http.StatusMethodNotAllowed)
			return
		}

		// Validate Accept header (similar to streamable.go)
		accept := r.Header.Get("Accept")
		if !acceptsContentType(accept, "application/json") {
			http.Error(w, "Accept header must include application/json", http.StatusNotAcceptable)
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse JSON-RPC message
		msg, err := parseJSONRPCMessage(body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON-RPC message: %v", err), http.StatusBadRequest)
			return
		}

		// Get server instance for this request
		server := getServer(r)
		if server == nil {
			http.Error(w, "Failed to create server instance", http.StatusInternalServerError)
			return
		}

		// Create stateless transport and handle the message
		transport := &statelessTransport{
			request:  r,
			response: w,
			message:  msg,
		}

		// Process the single request/response cycle
		ctx := context.Background()
		err = server.Run(ctx, transport)
		if err != nil {
			// If transport hasn't written a response yet, send error
			if !transport.responseWritten {
				http.Error(w, fmt.Sprintf("Server error: %v", err), http.StatusInternalServerError)
			}
		}
	})
}

// parseJSONRPCMessage parses a JSON-RPC message from raw bytes
func parseJSONRPCMessage(data []byte) (jsonrpc.Message, error) {
	// Parse raw JSON to determine message type
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Determine if it's a request/notification or response based on presence of method
	if _, hasMethod := raw["method"]; hasMethod {
		var req jsonrpc.Request
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}
		return &req, nil
	} else {
		var resp jsonrpc.Response
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("invalid response: %w", err)
		}
		return &resp, nil
	}
}

// acceptsContentType checks if the Accept header includes the specified content type
func acceptsContentType(accept, contentType string) bool {
	if accept == "" {
		return true // No Accept header means accept everything
	}

	// Simple check for the content type in the Accept header
	return strings.Contains(accept, contentType) || strings.Contains(accept, "*/*")
}

// statelessTransport implements mcp.Transport for single request/response cycles
type statelessTransport struct {
	request         *http.Request
	response        http.ResponseWriter
	message         jsonrpc.Message
	responseWritten bool
}

// Connect implements mcp.Transport.Connect
func (t *statelessTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	return &statelessConnection{
		request:         t.request,
		response:        t.response,
		message:         t.message,
		transport:       t,
		messageConsumed: false,
	}, nil
}

// statelessConnection implements mcp.Connection for single message handling
type statelessConnection struct {
	request         *http.Request
	response        http.ResponseWriter
	message         jsonrpc.Message
	transport       *statelessTransport
	messageConsumed bool
	responseWritten bool
}

// Read implements mcp.Connection.Read - returns the single HTTP request message
func (c *statelessConnection) Read(ctx context.Context) (jsonrpc.Message, error) {
	if c.messageConsumed {
		// After reading the one message, indicate end of stream
		return nil, io.EOF
	}

	c.messageConsumed = true
	return c.message, nil
}

// Write implements mcp.Connection.Write - writes the JSON-RPC response to HTTP response
func (c *statelessConnection) Write(ctx context.Context, msg jsonrpc.Message) error {
	if c.responseWritten {
		return fmt.Errorf("response already written")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	c.response.Header().Set("Content-Type", "application/json")
	c.response.WriteHeader(http.StatusOK)
	_, err = c.response.Write(data)

	c.responseWritten = true
	c.transport.responseWritten = true

	return err
}

// Close implements mcp.Connection.Close - no-op for stateless connections
func (c *statelessConnection) Close() error {
	return nil
}

// SessionID implements mcp.Connection.SessionID - returns empty string for stateless operation
func (c *statelessConnection) SessionID() string {
	return ""
}
