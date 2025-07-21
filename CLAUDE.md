# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go HTTP server that implements a Model Context Protocol (MCP) server with an echo tool. The server provides both a health check endpoint and MCP protocol support via HTTP with both stateful and stateless implementations.

## Development Commands

The project uses `mise` for task management. Key commands:

- `mise run test` - Run tests with gotestsum
- `mise run build` - Build the server binary to `bin/server`
- `mise run dev` - Run the server in development mode
- `go test -v ./...` - Run tests directly with go test
- `go run .` - Run the server directly
- `go run . -port 3000` - Run on a specific port

## Architecture

The server implements MCP (Model Context Protocol) using the official Go SDK with custom transport implementations:

### Core Components

- **HTTP Server**: Standard Go HTTP server with three endpoints:
  - `/readiness` - Health check endpoint returning "OK"
  - `/mcp-stateful` - Stateful MCP endpoint using StreamableHTTP with sessions
  - `/mcp-stateless` - Stateless MCP endpoint using custom transport without sessions

- **MCP Integration**: Uses `github.com/modelcontextprotocol/go-sdk` v0.2.0
  - Server implements an "echo" tool that returns input messages
  - Includes structured logging middleware for all MCP method calls
  - Tool input validation via JSON schema

### Key Files

- `main.go` - Main server implementation with MCP setup
- `stateless_streamable_http_transport.go` - Custom stateless HTTP transport implementation
- `main_test.go` - Basic HTTP endpoint tests
- `mise.toml` - Task definitions and tool versions
- `go.mod` - Go module with MCP SDK dependency

### MCP Server Structure

The MCP server is created with:
- Implementation name: "echo-server"
- Version: "1.0.0"
- Single tool: "echo" with string message parameter
- Logging middleware that tracks method execution time and results

### Transport Implementations

#### Stateful Transport (`/mcp-stateful`)
- Uses the standard `mcp.NewStreamableHTTPHandler`
- Creates sessions with `Mcp-Session-Id` headers
- Supports Server-Sent Events (SSE) for streaming responses
- Maintains connection state and session management

#### Stateless Transport (`/mcp-stateless`)
- Uses custom `NewStatelessStreamableHTTPHandler` from `stateless_streamable_http_transport.go`
- No session management - each request is independent
- Returns plain JSON responses instead of SSE
- No `Mcp-Session-Id` headers
- Simplified request/response cycle without connection persistence

### Tool Implementation

The echo tool accepts a JSON object with an optional "message" field and returns it prefixed with "Echo: ". Input validation is handled via JSON schema definition inline in the code.

## Reference Documentation

### MCP Protocol Specification
- **Full Specification**: https://modelcontextprotocol.io/llms-full.txt
- Complete MCP protocol documentation including message types, capabilities, tools, resources, and implementation requirements

### Go SDK Documentation
- **Design Document**: https://raw.githubusercontent.com/modelcontextprotocol/go-sdk/refs/heads/main/design/design.md
- Comprehensive design principles, architecture, type safety features, middleware system, and implementation philosophy for the Go SDK

### Key Design Insights
- The stateless implementation eliminates ~70% of the complexity from the standard StreamableHTTP transport
- Custom transports can leverage the MCP SDK's transport interface while bypassing session management
- The `SessionID()` method returning an empty string prevents session header generation