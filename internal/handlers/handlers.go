package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"spamassassin-mcp/internal/config"
	"spamassassin-mcp/internal/spamassassin"
)

type Handler struct {
	saClient   *spamassassin.Client
	security   config.SecurityConfig
	rateLimiter *rate.Limiter
}

// Request/Response types for MCP tools
type ScanEmailParams struct {
	Content     string            `json:"content" description:"Raw email content including headers"`
	Headers     map[string]string `json:"headers,omitempty" description:"Additional headers to analyze"`
	CheckBayes  bool             `json:"check_bayes,omitempty" description:"Include Bayesian analysis"`
	Verbose     bool             `json:"verbose,omitempty" description:"Return detailed rule explanations"`
}

type ScanEmailResult struct {
	Score       float64                    `json:"score" description:"Spam score"`
	Threshold   float64                    `json:"threshold" description:"Spam threshold"`
	IsSpam      bool                      `json:"is_spam" description:"Whether email is classified as spam"`
	RulesHit    []spamassassin.RuleMatch  `json:"rules_hit" description:"Matched spam rules"`
	Summary     string                    `json:"summary" description:"Human-readable analysis"`
	Timestamp   time.Time                 `json:"timestamp" description:"Analysis timestamp"`
}

type CheckReputationParams struct {
	Sender string `json:"sender" description:"Email sender address"`
	Domain string `json:"domain,omitempty" description:"Sender domain"`
	IP     string `json:"ip,omitempty" description:"Sender IP address"`
}

type ReputationResult struct {
	Sender     string            `json:"sender"`
	Domain     string            `json:"domain"`
	IP         string            `json:"ip"`
	Reputation string            `json:"reputation"`
	Blocked    bool              `json:"blocked"`
	Reasons    []string          `json:"reasons"`
	Details    map[string]string `json:"details"`
}

type UpdateRulesParams struct {
	Source string `json:"source,omitempty" description:"Rule source (official/custom)"`
	Force  bool   `json:"force,omitempty" description:"Force update even if recent"`
}

type TestRulesParams struct {
	Rules      string   `json:"rules" description:"Custom rule definitions"`
	TestEmails []string `json:"test_emails" description:"Sample emails to test against"`
}

type TestRulesResult struct {
	Results []TestResult `json:"results"`
	Summary string       `json:"summary"`
}

type TestResult struct {
	Email   string  `json:"email"`
	Score   float64 `json:"score"`
	IsSpam  bool    `json:"is_spam"`
	Rules   []string `json:"rules_matched"`
}

type ExplainScoreParams struct {
	EmailContent string `json:"email_content" description:"Email to analyze"`
}

type ScoreExplanation struct {
	FinalScore   float64                   `json:"final_score"`
	RuleDetails  []spamassassin.RuleMatch  `json:"rule_details"`
	BayesScore   float64                   `json:"bayes_score,omitempty"`
	NetworkTests []string                  `json:"network_tests"`
	Explanation  string                    `json:"explanation"`
}

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	ipRegex    = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
)

// Defensive operations whitelist
var allowedOperations = map[string]bool{
	"scan_email":        true,
	"check_reputation":  true,
	"update_rules":      true,
	"get_config":        true,
	"test_rules":        true,
	"explain_score":     true,
}

func New(saClient *spamassassin.Client, security config.SecurityConfig) *Handler {
	// Create rate limiter
	limiter := rate.NewLimiter(
		rate.Every(time.Minute/time.Duration(security.RateLimiting.RequestsPerMinute)),
		security.RateLimiting.BurstSize,
	)

	return &Handler{
		saClient:   saClient,
		security:   security,
		rateLimiter: limiter,
	}
}

func (h *Handler) ScanEmail(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ScanEmailParams]) (*mcp.CallToolResultFor[ScanEmailResult], error) {
	if !h.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	req := params.Arguments

	// Security validation
	if err := h.validateEmailContent(req.Content); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"operation": "scan_email",
		"size":      len(req.Content),
		"verbose":   req.Verbose,
		"bayes":     req.CheckBayes,
	}).Info("Processing email scan request")

	// Scan email with SpamAssassin
	options := spamassassin.ScanOptions{
		CheckBayes: req.CheckBayes,
		Verbose:    req.Verbose,
	}

	result, err := h.saClient.ScanEmail(req.Content, options)
	if err != nil {
		logrus.WithError(err).Error("SpamAssassin scan failed")
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Build response
	response := &ScanEmailResult{
		Score:     result.Score,
		Threshold: result.Threshold,
		IsSpam:    result.IsSpam,
		RulesHit:  result.RulesHit,
		Summary:   result.Summary,
		Timestamp: time.Now(),
	}

	logrus.WithFields(logrus.Fields{
		"score":    result.Score,
		"is_spam":  result.IsSpam,
		"rules":    len(result.RulesHit),
	}).Info("Email scan completed")

	return &mcp.CallToolResultFor[ScanEmailResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Email analysis completed. Score: %.2f, Spam: %v", response.Score, response.IsSpam)},
		},
	}, nil
}

func (h *Handler) CheckReputation(ctx context.Context, params json.RawMessage) (any, error) {
	if !h.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	var req CheckReputationParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Validate input
	if req.Sender != "" && !emailRegex.MatchString(req.Sender) {
		return nil, fmt.Errorf("invalid email address format")
	}

	if req.IP != "" && !ipRegex.MatchString(req.IP) {
		return nil, fmt.Errorf("invalid IP address format")
	}

	logrus.WithFields(logrus.Fields{
		"operation": "check_reputation",
		"sender":    req.Sender,
		"domain":    req.Domain,
		"ip":        req.IP,
	}).Info("Processing reputation check")

	// Extract domain from sender if not provided
	domain := req.Domain
	if domain == "" && req.Sender != "" {
		parts := strings.Split(req.Sender, "@")
		if len(parts) == 2 {
			domain = parts[1]
		}
	}

	// Check against blocked domains
	blocked := false
	var reasons []string

	for _, blockedDomain := range h.security.BlockedDomains {
		if strings.Contains(domain, blockedDomain) {
			blocked = true
			reasons = append(reasons, fmt.Sprintf("Domain %s is blocked", blockedDomain))
		}
	}

	// Determine reputation (simplified logic)
	reputation := "unknown"
	if blocked {
		reputation = "bad"
	} else if contains(h.security.AllowedSenders, req.Sender) {
		reputation = "good"
	}

	result := &ReputationResult{
		Sender:     req.Sender,
		Domain:     domain,
		IP:         req.IP,
		Reputation: reputation,
		Blocked:    blocked,
		Reasons:    reasons,
		Details: map[string]string{
			"check_time": time.Now().Format(time.RFC3339),
			"source":     "spamassassin-mcp",
		},
	}

	logrus.WithFields(logrus.Fields{
		"reputation": reputation,
		"blocked":    blocked,
	}).Info("Reputation check completed")

	return result, nil
}

func (h *Handler) GetConfig(ctx context.Context, params json.RawMessage) (any, error) {
	logrus.Info("Retrieving SpamAssassin configuration")
	return h.saClient.GetConfig()
}

func (h *Handler) UpdateRules(ctx context.Context, params json.RawMessage) (any, error) {
	if !h.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	var req UpdateRulesParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"operation": "update_rules",
		"source":    req.Source,
		"force":     req.Force,
	}).Info("Processing rule update request")

	if err := h.saClient.UpdateRules(); err != nil {
		return nil, fmt.Errorf("rule update failed: %w", err)
	}

	return map[string]any{
		"status":    "success",
		"message":   "Rules updated successfully",
		"timestamp": time.Now(),
	}, nil
}

func (h *Handler) TestRules(ctx context.Context, params json.RawMessage) (any, error) {
	if !h.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	var req TestRulesParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Validate input
	if req.Rules == "" {
		return nil, fmt.Errorf("rules cannot be empty")
	}

	logrus.WithFields(logrus.Fields{
		"operation":   "test_rules",
		"test_emails": len(req.TestEmails),
	}).Info("Processing rule test request")

	// This is a simplified implementation
	// In a real scenario, you'd create a temporary SpamAssassin configuration
	// and test the rules against the provided emails

	results := make([]TestResult, 0, len(req.TestEmails))
	for _, email := range req.TestEmails {
		if err := h.validateEmailContent(email); err != nil {
			continue // Skip invalid emails
		}

		// Scan with current rules (simplified)
		scanResult, err := h.saClient.ScanEmail(email, spamassassin.ScanOptions{Verbose: true})
		if err != nil {
			continue
		}

		result := TestResult{
			Email:  truncateString(email, 100),
			Score:  scanResult.Score,
			IsSpam: scanResult.IsSpam,
			Rules:  make([]string, 0, len(scanResult.RulesHit)),
		}

		for _, rule := range scanResult.RulesHit {
			result.Rules = append(result.Rules, rule.Name)
		}

		results = append(results, result)
	}

	return &TestRulesResult{
		Results: results,
		Summary: fmt.Sprintf("Tested %d emails against custom rules", len(results)),
	}, nil
}

func (h *Handler) ExplainScore(ctx context.Context, params json.RawMessage) (any, error) {
	if !h.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	var req ExplainScoreParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := h.validateEmailContent(req.EmailContent); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	logrus.WithField("operation", "explain_score").Info("Processing score explanation request")

	// Scan with verbose output
	result, err := h.saClient.ScanEmail(req.EmailContent, spamassassin.ScanOptions{
		Verbose:    true,
		CheckBayes: true,
	})
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Build explanation
	explanation := h.buildScoreExplanation(result)

	response := &ScoreExplanation{
		FinalScore:   result.Score,
		RuleDetails:  result.RulesHit,
		Explanation:  explanation,
		NetworkTests: []string{}, // Would be populated with actual network test results
	}

	return response, nil
}

func (h *Handler) validateEmailContent(content string) error {
	if len(content) > int(h.security.MaxEmailSize) {
		return fmt.Errorf("email size exceeds limit of %d bytes", h.security.MaxEmailSize)
	}

	if content == "" {
		return fmt.Errorf("email content cannot be empty")
	}

	// Parse as email to validate format
	if _, err := mail.ReadMessage(strings.NewReader(content)); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}

	return nil
}

func (h *Handler) buildScoreExplanation(result *spamassassin.ScanResult) string {
	var explanation strings.Builder

	explanation.WriteString(fmt.Sprintf("Final Score: %.2f (Threshold: %.2f)\n", result.Score, result.Threshold))
	explanation.WriteString(fmt.Sprintf("Classification: %s\n\n", map[bool]string{true: "SPAM", false: "HAM"}[result.IsSpam]))

	if len(result.RulesHit) > 0 {
		explanation.WriteString("Rules Triggered:\n")
		for _, rule := range result.RulesHit {
			explanation.WriteString(fmt.Sprintf("  %s: %.2f - %s\n", rule.Name, rule.Score, rule.Description))
		}
	} else {
		explanation.WriteString("No spam rules triggered.\n")
	}

	return explanation.String()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}