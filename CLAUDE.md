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

- **HTTP Server**: Standard Go HTTP server with two endpoints:
  - `/readiness` - Health check endpoint returning "OK"
  - `/mcp` - MCP protocol endpoint for tool interactions

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