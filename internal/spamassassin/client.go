package spamassassin

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"spamassassin-mcp/internal/config"
)

type Client struct {
	host      string
	port      int
	timeout   time.Duration
	threshold float64
}

type ScanResult struct {
	Score     float64
	Threshold float64
	IsSpam    bool
	RulesHit  []RuleMatch
	Summary   string
	Headers   map[string]string
}

type RuleMatch struct {
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
}

type ConfigInfo struct {
	Version      string         `json:"version"`
	Threshold    float64        `json:"threshold"`
	BayesEnabled bool           `json:"bayes_enabled"`
	RuleCount    int            `json:"rule_count"`
	Settings     map[string]any `json:"settings"`
}

var (
	scoreRegex = regexp.MustCompile(`(-?\d+\.?\d*)/(-?\d+\.?\d*)`)
	ruleRegex  = regexp.MustCompile(`\s*(-?\d+\.?\d*)\s+(\w+)\s+(.*)`)
)

func NewClient(cfg config.SpamAssassinConfig) (*Client, error) {
	client := &Client{
		host:      cfg.Host,
		port:      cfg.Port,
		timeout:   cfg.Timeout,
		threshold: cfg.Threshold,
	}

	// Test connection
	if err := client.ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to SpamAssassin: %w", err)
	}

	logrus.Infof("Connected to SpamAssassin at %s:%d", client.host, client.port)
	return client, nil
}

func (c *Client) ping() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.host, c.port), c.timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send PING command
	_, err = conn.Write([]byte("PING SPAMC/1.2\r\n\r\n"))
	if err != nil {
		return err
	}

	// Read response
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		response := scanner.Text()
		if strings.Contains(response, "PONG") {
			return nil
		}
		return fmt.Errorf("unexpected response: %s", response)
	}

	return fmt.Errorf("no response from SpamAssassin")
}

func (c *Client) ScanEmail(content string, options ScanOptions) (*ScanResult, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.host, c.port), c.timeout)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	// Build command
	cmd := "CHECK"
	if options.Verbose {
		cmd = "REPORT"
	}

	// Send headers
	headers := fmt.Sprintf("%s SPAMC/1.2\r\nContent-length: %d\r\n", cmd, len(content))
	if options.CheckBayes {
		headers += "User: bayes\r\n"
	}
	headers += "\r\n"

	// Send request
	_, err = conn.Write([]byte(headers + content))
	if err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}

	// Read response
	return c.parseResponse(conn, options.Verbose)
}

func (c *Client) parseResponse(conn net.Conn, verbose bool) (*ScanResult, error) {
	scanner := bufio.NewScanner(conn)
	result := &ScanResult{
		Threshold: c.threshold,
		Headers:   make(map[string]string),
		RulesHit:  make([]RuleMatch, 0),
	}

	// Parse response headers
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break // End of headers
		}

		if strings.HasPrefix(line, "Spam:") {
			// Parse spam status line
			if err := c.parseSpamLine(line, result); err != nil {
				return nil, err
			}
		} else {
			// Store other headers
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Parse message body if verbose
	if verbose {
		var body strings.Builder
		for scanner.Scan() {
			body.WriteString(scanner.Text() + "\n")
		}
		result.Summary = body.String()
		c.parseRules(result.Summary, result)
	}

	result.IsSpam = result.Score >= result.Threshold

	return result, scanner.Err()
}

func (c *Client) parseSpamLine(line string, result *ScanResult) error {
	// Example: "Spam: True ; 15.3 / 5.0"
	matches := scoreRegex.FindStringSubmatch(line)
	if len(matches) != 3 {
		return fmt.Errorf("invalid spam line format: %s", line)
	}

	score, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return fmt.Errorf("invalid score: %s", matches[1])
	}

	threshold, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return fmt.Errorf("invalid threshold: %s", matches[2])
	}

	result.Score = score
	result.Threshold = threshold

	return nil
}

func (c *Client) parseRules(content string, result *ScanResult) {
	lines := strings.Split(content, "\n")
	inRulesSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "pts rule name") {
			inRulesSection = true
			continue
		}

		if inRulesSection && line != "" {
			matches := ruleRegex.FindStringSubmatch(line)
			if len(matches) == 4 {
				score, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					continue
				}

				rule := RuleMatch{
					Name:        matches[2],
					Score:       score,
					Description: matches[3],
				}
				result.RulesHit = append(result.RulesHit, rule)
			}
		}
	}
}

func (c *Client) GetConfig() (*ConfigInfo, error) {
	// This would require additional SpamAssassin integration
	// For now, return basic info
	return &ConfigInfo{
		Version:      "3.4.x",
		Threshold:    c.threshold,
		BayesEnabled: true,
		RuleCount:    1000, // Approximate
		Settings: map[string]any{
			"host":    c.host,
			"port":    c.port,
			"timeout": c.timeout.String(),
		},
	}, nil
}

func (c *Client) UpdateRules() error {
	// In a real implementation, this would trigger rule updates
	logrus.Info("Rule update requested (not implemented in basic client)")
	return nil
}

type ScanOptions struct {
	CheckBayes bool
	Verbose    bool
}