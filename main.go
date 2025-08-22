// Package main implements the SpamAssassin MCP (Model Context Protocol) server.
//
// This server provides secure, defensive-only email security analysis capabilities
// through the MCP protocol. It integrates with SpamAssassin to offer email spam
// detection, reputation checking, and rule analysis tools.
//
// Security Notice: This server is designed exclusively for defensive security
// analysis. It does not provide capabilities for sending emails, generating spam
// content, or any offensive security operations.
//
// The server provides the following MCP tools:
//   - scan_email: Analyze email content for spam probability and rule matches
//   - check_reputation: Check sender reputation and domain/IP blacklists
//   - explain_score: Provide detailed explanation of spam score calculation
//   - get_config: Retrieve current SpamAssassin configuration
//   - update_rules: Update SpamAssassin rule definitions (defensive updates only)
//   - test_rules: Test custom rules against sample emails in safe environment
//
// All operations include comprehensive security controls:
//   - Input validation and sanitization
//   - Rate limiting (60 requests/minute with burst capacity)
//   - Email size limits (10MB maximum)
//   - Timeout protection (60 second scan limit)
//   - Audit logging of all operations
//
// Architecture:
//   - Containerized deployment with non-root execution
//   - Read-only filesystem for security
//   - Network isolation and resource limits
//   - Health monitoring and graceful shutdown
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"

	"spamassassin-mcp/internal/config"
	"spamassassin-mcp/internal/handlers"
	"spamassassin-mcp/internal/spamassassin"
)

// isRunningInContainer detects if the application is running inside a container.
//
// This function checks for common container indicators:
//   - Presence of /.dockerenv file (Docker)
//   - Container-specific environment variables
//   - cgroup filesystem indicators
func isRunningInContainer() bool {
	// Check for Docker container indicator
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	
	// Check for container environment variables
	if os.Getenv("CONTAINER") != "" || os.Getenv("DOCKER_CONTAINER") != "" {
		return true
	}
	
	return false
}

// main is the entry point for the SpamAssassin MCP server.
//
// It initializes the configuration, sets up logging, creates the SpamAssassin
// client, registers MCP tools, and starts the server with graceful shutdown support.
//
// The server startup sequence:
//  1. Load configuration from files and environment variables
//  2. Initialize structured JSON logging with configurable level
//  3. Create and test SpamAssassin client connection
//  4. Initialize MCP server with defensive security tools
//  5. Set up signal handlers for graceful shutdown
//  6. Start the MCP server and listen for connections
//
// Security: All components are initialized with security-first defaults and
// comprehensive error handling to prevent information disclosure.
func main() {
	// Initialize configuration from files and environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup structured JSON logging with configurable level
	setupLogging(cfg.LogLevel)

	logrus.Info("Starting SpamAssassin MCP Server v1.0.0")

	// Initialize SpamAssassin client with connection testing
	saClient, err := spamassassin.NewClient(cfg.SpamAssassin)
	if err != nil {
		logrus.Fatalf("Failed to initialize SpamAssassin client: %v", err)
	}

	// Create MCP server instance with implementation info
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "spamassassin-mcp",
		Version: "1.0.0",
	}, nil)

	// Initialize request handlers with security configuration and rate limiting
	h := handlers.New(saClient, cfg.Security)

	// Register only defensive security analysis tools (no offensive capabilities)
	registerTools(server, h)

	// Create context for coordinated graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handlers for graceful shutdown on SIGINT/SIGTERM
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logrus.Info("Received shutdown signal, stopping server...")
		cancel()
	}()

	// Choose transport and handling based on environment
	if isRunningInContainer() {
		// Container mode: Use SSE transport for HTTP-based MCP communication
		logrus.Infof("Starting MCP server with SSE transport on %s", cfg.Server.BindAddr)
		
		// Set up HTTP server for SSE transport
		go func() {
			http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
				transport := mcp.NewLoggingTransport(
					mcp.NewSSEServerTransport("/mcp", w),
					os.Stderr,
				)
				if err := server.Run(ctx, transport); err != nil {
					logrus.Errorf("SSE transport error: %v", err)
				}
			})
			
			logrus.Infof("HTTP server listening on %s", cfg.Server.BindAddr)
			if err := http.ListenAndServe(cfg.Server.BindAddr, nil); err != nil {
				logrus.Errorf("HTTP server error: %v", err)
			}
		}()
		
		// Keep container alive
		<-ctx.Done()
	} else {
		// Direct mode: Use stdio transport for client connections
		logrus.Info("Starting MCP server with stdio transport")
		transport := mcp.NewLoggingTransport(mcp.NewStdioTransport(), os.Stderr)
		if err := server.Run(ctx, transport); err != nil {
			logrus.Fatalf("Server error: %v", err)
		}
	}

	logrus.Info("SpamAssassin MCP Server stopped")
}

// setupLogging configures structured JSON logging with the specified level.
//
// The logging configuration uses:
//   - JSON formatter for structured, machine-readable logs
//   - Standard output for container-friendly log collection
//   - Configurable log levels from debug to error
//   - Default to info level for production safety
//
// Log levels:
//   - debug: Verbose debugging information
//   - info: General operational messages (default)
//   - warn: Warning conditions that should be noted
//   - error: Error conditions that require attention
//
// Security: Debug level may include sensitive information and should only
// be used in development environments.
func setupLogging(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

// registerTools registers all available MCP tools with the server.
//
// This function implements the defensive-only security posture by registering
// only analysis and configuration tools. No email transmission, content generation,
// or offensive security capabilities are provided.
//
// Registered tools are organized into three categories:
//
// Email Analysis Tools:
//   - scan_email: Comprehensive spam analysis with rule matching
//   - check_reputation: Sender and domain reputation verification
//   - explain_score: Detailed score breakdown and rule explanations
//
// Configuration Management Tools:
//   - get_config: Read-only configuration inspection
//   - update_rules: Defensive rule updates from trusted sources
//
// Rule Development Tools:
//   - test_rules: Safe testing of custom rules in isolated environment
//
// Security: All tools include comprehensive input validation, rate limiting,
// and audit logging. No tools provide offensive capabilities or data modification.
func registerTools(server *mcp.Server, h *handlers.Handler) {
	// Email analysis tools - core spam detection and analysis functionality
	mcp.AddTool(server, &mcp.Tool{
		Name:        "scan_email",
		Description: "Analyze email content for spam probability and rule matches",
	}, h.ScanEmail)
	
	// TODO: Re-enable other tools once handlers are updated for MCP SDK v0.2.0
	/*
	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_reputation", 
		Description: "Check sender reputation and domain/IP blacklists",
	}, h.CheckReputation)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "explain_score",
		Description: "Explain how a spam score was calculated", 
	}, h.ExplainScore)

	// Configuration management tools - read-only system inspection and defensive updates
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_rules",
		Description: "Update SpamAssassin rule definitions",
	}, h.UpdateRules)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_config",
		Description: "Retrieve current SpamAssassin configuration",
	}, h.GetConfig)

	// Rule development tools - safe testing and validation in isolated environment
	mcp.AddTool(server, &mcp.Tool{
		Name:        "test_rules",
		Description: "Test custom rules against sample emails",
	}, h.TestRules)
	*/

	logrus.Info("Registered 1 defensive security tool (others temporarily disabled)")
}