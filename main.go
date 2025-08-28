package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Configuration structure
type Config struct {
	JiraBaseURL   string `json:"jira_base_url"`
	Email         string `json:"email"`
	APIToken      string `json:"api_token"`
	BoardID       string `json:"board_id"`
	ProjectKey    string `json:"project_key"`
	CheckInterval int    `json:"check_interval_minutes"`
	JQL           string `json:"jql"`
}

// Jira API response structures
type JiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
		Labels []string `json:"labels"`
	} `json:"fields"`
}

type JiraSearchResponse struct {
	Issues []JiraIssue `json:"issues"`
	Total  int         `json:"total"`
}

type JiraMonitor struct {
	config     Config
	client     *http.Client
	lastIssues []JiraIssue
	isFirstRun bool
}

func main() {
	fmt.Println("🚀 Starting Jira New Ticket Notifier...")

	monitor, err := NewJiraMonitor()
	if err != nil {
		log.Fatal("Failed to initialize monitor:", err)
	}

	// No test notifications

	// Run initial check
	monitor.checkAndNotify()

	// Set up periodic checking
	ticker := time.NewTicker(time.Duration(monitor.config.CheckInterval) * time.Minute)
	defer ticker.Stop()

	fmt.Printf("✅ Monitor started! Checking every %d minutes...\n", monitor.config.CheckInterval)
	fmt.Println("Press Ctrl+C to stop")

	for range ticker.C {
		monitor.checkAndNotify()
	}
}

func NewJiraMonitor() (*JiraMonitor, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	monitor := &JiraMonitor{
		config:     config,
		client:     &http.Client{Timeout: 30 * time.Second},
		isFirstRun: true,
	}

	return monitor, nil
}

func loadConfig() (Config, error) {
	config := Config{
		CheckInterval: 5, // default fallback
	}

	// Load environment variables from .env file
	if err := loadEnvFile(".env"); err != nil {
		fmt.Println("⚠️  No .env file found. Creating one...")
		if err := createEnvFile(); err != nil {
			return config, fmt.Errorf("failed to create .env file: %v", err)
		}
		fmt.Println("✅ Created .env file. Please fill it with your credentials and run again.")
		os.Exit(0)
	}

	// Load from environment variables
	config.JiraBaseURL = os.Getenv("JIRA_BASE_URL")
	config.Email = os.Getenv("JIRA_EMAIL")
	config.APIToken = os.Getenv("JIRA_API_TOKEN")
	config.ProjectKey = os.Getenv("JIRA_PROJECT_KEY")
	config.BoardID = os.Getenv("JIRA_BOARD_ID")
	config.JQL = os.Getenv("JIRA_JQL")

	// Load check interval from env or use default
	if intervalStr := os.Getenv("CHECK_INTERVAL_MINUTES"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			config.CheckInterval = interval
		}
	}

	// Validate required fields
	if config.JiraBaseURL == "" || config.Email == "" || config.APIToken == "" || config.JQL == "" {
		return config, fmt.Errorf("missing required environment variables. Please check your .env file")
	}

	fmt.Println("📋 Configuration loaded from .env file")
	return config, nil
}

func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func createEnvFile() error {
	envContent := `# Jira Configuration
# Fill in your details below and save the file

# Your Jira instance URL (e.g., https://your-company.atlassian.net)
JIRA_BASE_URL=https://c24-hotel.atlassian.net

# Your Jira email address
JIRA_EMAIL=your-email@example.com

# Your Jira API token (get from: https://id.atlassian.com/manage-profile/security/api-tokens)
JIRA_API_TOKEN=your-api-token-here

# Your project key (optional - only needed if not using custom JQL)
JIRA_PROJECT_KEY=HOAPI

# Your board ID (optional - extracted from URL if not specified)
JIRA_BOARD_ID=14

# Custom JQL query for filtering tickets
JIRA_JQL=project = "Hotel - API" AND assignee is EMPTY AND status = "To Do" AND (labels is EMPTY OR labels not IN (HO-Provider))

# Check interval in minutes (default: 5)
CHECK_INTERVAL_MINUTES=5
`

	return os.WriteFile(".env", []byte(envContent), 0644)
}

func (m *JiraMonitor) checkAndNotify() {
	fmt.Printf("\n🔍 Checking at %s...\n", time.Now().Format("15:04:05"))

	currentIssues, err := m.getTodoIssues()
	if err != nil {
		log.Printf("❌ Error checking Jira: %v", err)
		return
	}

	fmt.Printf("📊 Current tickets: %d\n", len(currentIssues))

	if m.isFirstRun {
		m.lastIssues = currentIssues
		m.isFirstRun = false
		fmt.Println("🎯 Initial tickets recorded:")
		for _, issue := range currentIssues {
			fmt.Printf("  • %s: %s%s\n", issue.Key, issue.Fields.Summary, m.formatLabels(issue.Fields.Labels))
		}
		return
	}

	// Compare current issues with previous
	added, removed := m.compareIssues(m.lastIssues, currentIssues)

	if len(added) > 0 {
		fmt.Printf("✨ %d new tickets:\n", len(added))
		for _, issue := range added {
			fmt.Printf("  • %s: %s%s\n", issue.Key, issue.Fields.Summary, m.formatLabels(issue.Fields.Labels))
		}

		// Log new tickets to file and send notification
		m.logNewTickets(added)
		m.sendNotification("🔔 NEW JIRA TICKETS!", m.formatNewTicketsMessage(added))

		m.lastIssues = currentIssues
	} else {
		fmt.Println("✅ No changes detected")

		// Still update if removed tickets (but don't show them)
		if len(removed) > 0 {
			m.lastIssues = currentIssues
		}
	}
}

func (m *JiraMonitor) compareIssues(oldIssues, newIssues []JiraIssue) (added, removed []JiraIssue) {
	oldKeys := make(map[string]JiraIssue)
	newKeys := make(map[string]JiraIssue)

	for _, issue := range oldIssues {
		oldKeys[issue.Key] = issue
	}

	for _, issue := range newIssues {
		newKeys[issue.Key] = issue
	}

	// Find added issues
	for key, issue := range newKeys {
		if _, exists := oldKeys[key]; !exists {
			added = append(added, issue)
		}
	}

	// Find removed issues
	for key, issue := range oldKeys {
		if _, exists := newKeys[key]; !exists {
			removed = append(removed, issue)
		}
	}

	return added, removed
}

func (m *JiraMonitor) getTodoIssues() ([]JiraIssue, error) {
	// Use the JQL query from configuration
	jql := m.config.JQL

	// Properly URL encode the JQL
	encodedJQL := strings.ReplaceAll(jql, " ", "%20")
	encodedJQL = strings.ReplaceAll(encodedJQL, "\"", "%22")
	encodedJQL = strings.ReplaceAll(encodedJQL, "=", "%3D")
	encodedJQL = strings.ReplaceAll(encodedJQL, "(", "%28")
	encodedJQL = strings.ReplaceAll(encodedJQL, ")", "%29")

	url := fmt.Sprintf("%s/rest/api/3/search?jql=%s&maxResults=1000&fields=status,summary,labels",
		m.config.JiraBaseURL, encodedJQL)

	fmt.Printf("🔍 Using JQL: %s\n", jql)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(m.config.Email, m.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jira API error: %d - %s", resp.StatusCode, string(body))
	}

	var searchResp JiraSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	return searchResp.Issues, nil
}

func (m *JiraMonitor) formatLabels(labels []string) string {
	if len(labels) == 0 {
		return ""
	}

	var formattedLabels []string
	for _, label := range labels {
		formattedLabels = append(formattedLabels, fmt.Sprintf("[%s]", label))
	}

	return " " + strings.Join(formattedLabels, " ")
}

func (m *JiraMonitor) logNewTickets(newTickets []JiraIssue) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("\n=== %s ===\n", timestamp)

	for _, ticket := range newTickets {
		logEntry += fmt.Sprintf("• %s: %s\n", ticket.Key, ticket.Fields.Summary)
	}

	// Append to log file
	file, err := os.OpenFile("new_tickets.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("⚠️  Could not write to log file: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(logEntry); err != nil {
		fmt.Printf("⚠️  Could not write to log file: %v\n", err)
		return
	}

	fmt.Printf("📝 Logged %d new tickets to new_tickets.log\n", len(newTickets))
}

func (m *JiraMonitor) formatNewTicketsMessage(newTickets []JiraIssue) string {
	if len(newTickets) == 0 {
		return "No new tickets"
	}

	if len(newTickets) == 1 {
		return "Found 1 new ticket"
	}

	return fmt.Sprintf("Found %d new tickets", len(newTickets))
}

func (m *JiraMonitor) sendNotification(title, message string) {
	fmt.Printf("🔔 Notification: %s - %s\n", title, message)

	// Method 1: Force visible dialog
	script := fmt.Sprintf(`display dialog "%s\n\n%s" with title "Jira Monitor" buttons {"OK"} default button "OK" giving up after 5`,
		strings.ReplaceAll(title, `"`, `\"`),
		strings.ReplaceAll(message, `"`, `\"`))

	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err == nil {
		fmt.Println("📬 Dialog notification sent!")
		return
	} else {
		fmt.Printf("⚠️  Dialog failed: %v\n", err)
	}

	// Method 2: Terminal notifier with app bundle
	cmd = exec.Command("terminal-notifier", "-title", title, "-message", message, "-sound", "Glass", "-appIcon", "https://cdn-icons-png.flaticon.com/512/889/889192.png")
	if err := cmd.Run(); err == nil {
		fmt.Println("📬 Terminal notifier sent!")
		return
	} else {
		fmt.Printf("⚠️  Terminal notifier failed: %v\n", err)
	}

	// Method 3: Just play sound and show in terminal
	exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Run()
	fmt.Printf("🔊 %s: %s\n", title, message)
	fmt.Println("📬 Audio + terminal notification sent!")
}
