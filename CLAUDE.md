# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go HTTP server that implements a Model Context Protocol (MCP) server with an echo tool. The server provides both a health check endpoint and MCP protocol support via HTTP.

## Development Commands

The project uses `mise` for task management. Key commands:

- `mise run test` - Run tests with gotestsum
- `mise run build` - Build the server binary to `bin/server`
- `mise run dev` - Run the server in development mode
- `go test -v ./...` - Run tests directly with go test
- `go run .` - Run the server directly
- `go run . -port 3000` - Run on a specific port

## Architecture

The server implements MCP (Model Context Protocol) using the official Go SDK:

### Core Components

- **HTTP Server**: Standard Go HTTP server with three endpoints:
  - `/readiness` - Health check endpoint returning "OK"
  - `/mcp` - Stateless MCP protocol endpoint (new server instance per request)
  - `/mcp-stateful` - Stateful MCP protocol endpoint (shared server instance)

- **MCP Integration**: Uses `github.com/modelcontextprotocol/go-sdk` v0.2.0
  - Server implements an "echo" tool that returns input messages
  - Includes structured logging middleware for all MCP method calls
  - Tool input validation via JSON schema

### Key Files

- `main.go` - Main server implementation with MCP setup
- `main_test.go` - Basic HTTP endpoint tests
- `mise.toml` - Task definitions and tool versions
- `go.mod` - Go module with MCP SDK dependency

### MCP Server Structure

The MCP server is created with:
- Implementation name: "echo-server"
- Version: "1.0.0"
- Single tool: "echo" with string message parameter
- Logging middleware that tracks method execution time and results

### Tool Implementation

The echo tool accepts a JSON object with an optional "message" field and returns it prefixed with "Echo: ". Input validation is handled via JSON schema definition inline in the code.

### StreamableHTTP Implementation Patterns

The server demonstrates two different approaches to implementing MCP over HTTP:

#### Stateless Implementation (`/mcp`)
- **Pattern**: Creates a new `mcp.Server` instance for each HTTP request
- **Function**: Uses `mcp.NewStreamableHTTPHandler` with a factory function that returns `createMCPServer(logger)` on each call
- **Benefits**: 
  - Complete isolation between requests
  - No shared state concerns
  - Memory gets garbage collected after each request
  - Safer for concurrent access
- **Trade-offs**: 
  - Higher memory allocation per request
  - Tool initialization overhead on each request
  - Stateless by design (cannot maintain session data)

#### Stateful Implementation (`/mcp-stateful`)
- **Pattern**: Reuses a single shared `mcp.Server` instance across all HTTP requests
- **Function**: Uses `mcp.NewStreamableHTTPHandler` with a factory function that returns the same pre-created server instance
- **Benefits**:
  - Lower memory allocation overhead
  - Server initialization only happens once
  - Can potentially maintain session state across requests
  - Better performance for high-traffic scenarios
- **Trade-offs**:
  - Shared state between concurrent requests
  - Requires careful consideration of thread safety
  - Memory persists for the application lifetime

Both implementations use the same `mcp.NewStreamableHTTPHandler` function but differ in how they provide the server instance to the factory function.