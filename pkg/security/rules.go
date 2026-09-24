package security

import (
	"regexp"
	"strings"
)

// SecretRule defines pattern matching logic for a credential type.
type SecretRule struct {
	ID          string
	Name        string
	Severity    Severity
	Description string
	Regex       *regexp.Regexp
	PathRegex   *regexp.Regexp
}

// BuiltinSecretRules returns a curated set of high-confidence credential detection rules.
func BuiltinSecretRules() []SecretRule {
	return []SecretRule{
		{
			ID:          "aws-access-key-id",
			Name:        "AWS Access Key ID",
			Severity:    SeverityCritical,
			Description: "Identifies AWS standard IAM access key identifiers",
			Regex:       regexp.MustCompile(`\b(AKIA[0-9A-Z]{16})\b`),
		},
		{
			ID:          "aws-secret-access-key",
			Name:        "AWS Secret Access Key",
			Severity:    SeverityCritical,
			Description: "Identifies AWS Secret Access Keys in key-value pairs",
			Regex:       regexp.MustCompile(`(?i)(?:aws_secret_access_key|aws_sec_key|secret_key)\s*[:=]\s*["']?([a-zA-Z0-9/+=]{40})["']?`),
		},
		{
			ID:          "github-pat",
			Name:        "GitHub Personal Access Token",
			Severity:    SeverityCritical,
			Description: "Identifies GitHub personal access tokens",
			Regex:       regexp.MustCompile(`\b(ghp_[a-zA-Z0-9]{36,40}|github_pat_[a-zA-Z0-9_]{82})\b`),
		},
		{
			ID:          "gitlab-pat",
			Name:        "GitLab Personal Access Token",
			Severity:    SeverityCritical,
			Description: "Identifies GitLab personal access tokens",
			Regex:       regexp.MustCompile(`\b(glpat-[a-zA-Z0-9\-_]{20,30})\b`),
		},
		{
			ID:          "private-key-pem",
			Name:        "Private Cryptographic Key",
			Severity:    SeverityCritical,
			Description: "Identifies unencrypted SSH/RSA/EC private key blocks",
			Regex:       regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH |PGP )?PRIVATE KEY-----`),
		},
		{
			ID:          "database-uri-credentials",
			Name:        "Database Connection URI with Password",
			Severity:    SeverityHigh,
			Description: "Identifies database connection strings with embedded passwords",
			Regex:       regexp.MustCompile(`(?i)(?:postgres|mysql|mongodb|redis|amqp):\/\/[^:\/\s]+:([^@\/\s]+)@[^\/\s]+`),
		},
		{
			ID:          "slack-token",
			Name:        "Slack Bot or User Token",
			Severity:    SeverityHigh,
			Description: "Identifies Slack workspace API tokens",
			Regex:       regexp.MustCompile(`\b(xox[baprs]-[0-9]{10,13}-[0-9]{10,13}-[a-zA-Z0-9]{24,32})\b`),
		},
		{
			ID:          "env-file-credentials",
			Name:        "Committed Environment Secrets File",
			Severity:    SeverityMedium,
			Description: "Identifies committed .env files containing tokens or secrets",
			PathRegex:   regexp.MustCompile(`(?i)(?:^|\/)\.env(?:\.(?:local|prod|production|staging|dev))?$`),
			Regex:       regexp.MustCompile(`(?i)(?:password|secret|token|api_key|private_key)\s*=\s*([^\s]{8,})`),
		},
		{
			ID:          "npmrc-auth-token",
			Name:        "NPM Registry Auth Token",
			Severity:    SeverityHigh,
			Description: "Identifies NPM registry authentication tokens in .npmrc",
			PathRegex:   regexp.MustCompile(`(?i)(?:^|\/)\.npmrc$`),
			Regex:       regexp.MustCompile(`_authToken\s*=\s*([a-f0-9\-]{36}|npm_[a-zA-Z0-9]{36})`),
		},
		{
			ID:          "docker-config-auth",
			Name:        "Docker Registry Config Auth",
			Severity:    SeverityCritical,
			Description: "Identifies Docker registry credentials in config.json",
			PathRegex:   regexp.MustCompile(`(?i)\.docker\/config\.json$`),
			Regex:       regexp.MustCompile(`"auth"\s*:\s*"([a-zA-Z0-9+/=]{20,})"`),
		},
	}
}

// MaskSecret redacts the sensitive portion of a secret while preserving prefix for identification.
func MaskSecret(raw string) string {
	raw = strings.TrimSpace(raw)
	n := len(raw)
	if n <= 6 {
		return "******"
	}
	if n <= 12 {
		return raw[:3] + strings.Repeat("*", n-3)
	}
	prefixLen := 4
	if strings.HasPrefix(raw, "AKIA") {
		prefixLen = 4
	} else if strings.HasPrefix(raw, "ghp_") {
		prefixLen = 8
	} else if strings.HasPrefix(raw, "glpat-") {
		prefixLen = 6
	}

	if prefixLen >= n {
		prefixLen = 3
	}

	return raw[:prefixLen] + strings.Repeat("*", n-prefixLen)
}
