# 🎯 Jira New Ticket Notifier

**Never miss a new ticket again!** 

A lightweight Go application that silently monitors your Jira tickets and instantly alerts you when new ones appear. Perfect for busy developers, project managers, and support teams who need to stay on top of their queue without constantly refreshing Jira.

## ✨ Why You'll Love This

- 🔕 **Silent & Smart**: Only notifies when there's actually something new (no notification spam!)
- 🎨 **Beautiful Output**: Clean terminal display with ticket labels and organized formatting
- ⚡ **Lightning Fast**: Efficient API calls that won't slow down your system
- 🔧 **Highly Configurable**: Use any JQL query to monitor exactly what matters to you
- 📝 **Persistent Logging**: Automatic logging so you never lose track of new tickets
- 🍎 **macOS Native**: Beautiful native notifications that respect your Do Not Disturb settings

## 🚀 Quick Start

**Prerequisites:** Make sure you have [Go installed](https://golang.org/doc/install) (1.19+ required)

```bash
# 1. Clone and setup
git clone https://github.com/vahiiiid/jira-new-ticket-notifier.git
cd jira-new-ticket-notifier

# 2. Create your config (will prompt for setup)
cat > .env << 'EOF'
JIRA_BASE_URL=https://your-company.atlassian.net
JIRA_EMAIL=your-email@company.com
JIRA_API_TOKEN=your-api-token-here
JIRA_JQL=project = "YOUR-PROJECT" AND assignee is EMPTY AND status = "To Do"
CHECK_INTERVAL_MINUTES=5
EOF

# 3. Run and enjoy!
go run main.go
```

## 📺 What You'll See

### Initial Startup
```
🚀 Starting Jira New Ticket Notifier...
📋 Configuration loaded from .env file
🔍 Checking at 14:30:15...
🔍 Using JQL: project = "MYPROJECT" AND assignee is EMPTY AND status = "To Do"
📊 Current tickets: 4
🎯 Initial tickets recorded:
  • PROJ-1234: Fix authentication bug [Critical] [Backend] [Security]
  • PROJ-1235: Update payment gateway [Enhancement] [Payment]
  • PROJ-1236: Improve error handling [Bug-Fix] [Frontend]
  • PROJ-1237: Database optimization [Performance]
✅ Monitor started! Checking every 5 minutes...
Press Ctrl+C to stop
```

### When New Tickets Appear
```
🔍 Checking at 14:35:20...
📊 Current tickets: 6
✨ 2 new tickets:
  • PROJ-1238: API rate limiting [Enhancement] [Backend]
  • PROJ-1239: Mobile app crash on iOS [Critical] [Mobile] [Bug-Fix]
📝 Logged 2 new tickets to new_tickets.log
📬 Notification sent!
```

### When Nothing Changes
```
🔍 Checking at 14:40:25...
📊 Current tickets: 6
✅ No changes detected
```

### macOS Notification (Simple & Clean)
```
🔔 NEW JIRA TICKETS!
Found 2 new tickets
```

## 🛠 Setup Guide

### 1. Get Your Jira API Token
1. Go to [Atlassian API Tokens](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click **"Create API token"**
3. Give it a name like "Jira New Ticket Notifier"
4. **Copy the token** (you won't see it again!)

### 2. Configure Your Environment

Create a `.env` file with your settings:

```bash
# Your Jira instance
JIRA_BASE_URL=https://your-company.atlassian.net

# Your credentials
JIRA_EMAIL=your-email@company.com
JIRA_API_TOKEN=your-secret-api-token

# Optional (legacy)
JIRA_PROJECT_KEY=YOUR-PROJECT
JIRA_BOARD_ID=123

# The magic happens here - customize this JQL query!
JIRA_JQL=project = "YOUR-PROJECT" AND assignee is EMPTY AND status = "To Do" AND (labels is EMPTY OR labels not IN (unwanted-label))

# How often to check (in minutes)
CHECK_INTERVAL_MINUTES=5
```

## 🎯 JQL Query Examples

The power is in the JQL! Here are some useful examples:

```bash
# Basic: All unassigned TO DO tickets
JIRA_JQL=project = "MYPROJECT" AND assignee is EMPTY AND status = "To Do"

# Monitor critical bugs only
JIRA_JQL=project = "MYPROJECT" AND priority = "Critical" AND status in ("Open", "To Do")

# Exclude specific labels (great for filtering out automated tickets)
JIRA_JQL=project = "MYPROJECT" AND assignee is EMPTY AND status = "To Do" AND (labels is EMPTY OR labels not IN (automated, scheduled))

# Multiple projects
JIRA_JQL=project in ("PROJECT1", "PROJECT2") AND assignee is EMPTY AND status = "To Do"

# Specific component focus
JIRA_JQL=project = "MYPROJECT" AND component = "Backend" AND status = "To Do"

# Recently created tickets only
JIRA_JQL=project = "MYPROJECT" AND created >= -24h AND status = "To Do"
```

## 📋 Log File Example

Every new ticket is automatically logged to `new_tickets.log`:

```
=== 2024-01-15 14:35:20 ===
• PROJ-1238: API rate limiting implementation
• PROJ-1239: Mobile app crash on iOS 17

=== 2024-01-15 15:42:15 ===
• PROJ-1240: Database connection timeout handling

=== 2024-01-15 16:18:33 ===
• PROJ-1241: User authentication refactor
• PROJ-1242: Payment processing improvements
• PROJ-1243: Frontend responsive design fixes
```

## 🏗 Production Deployment

### Build and Run
```bash
# Build the binary
go build -o jira-new-ticket-notifier main.go

# Run in foreground
./jira-new-ticket-notifier

# Run in background
nohup ./jira-new-ticket-notifier > jira-new-ticket-notifier.log 2>&1 &

# Check if running
ps aux | grep jira-new-ticket-notifier

# Stop background process
pkill jira-new-ticket-notifier
```

### macOS Service (Optional)
Create a Launch Agent for automatic startup:

```bash
# Create service file
cat > ~/Library/LaunchAgents/com.yourname.jira-new-ticket-notifier.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.yourname.jira-new-ticket-notifier</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/your/jira-new-ticket-notifier</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/path/to/your/project</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
EOF

# Load the service
launchctl load ~/Library/LaunchAgents/com.yourname.jira-new-ticket-notifier.plist
```

## 🔧 Troubleshooting

### No Notifications Appearing?

1. **Check Terminal app permissions:**
   - System Preferences → Notifications & Focus → Terminal
   - Enable "Allow Notifications"

2. **Test manually:**
   ```bash
   terminal-notifier -title "Test" -message "Working?" -sound Glass
   ```

3. **Check Do Not Disturb:**
   - Make sure DND is off or Terminal is allowed

### API Issues?

1. **Verify your token:**
   ```bash
   curl -u your-email@company.com:your-api-token \
        https://your-company.atlassian.net/rest/api/3/myself
   ```

2. **Test your JQL:**
   - Go to Jira → Issues → Search
   - Paste your JQL query and verify it returns expected results

### Performance Tips

- Set `CHECK_INTERVAL_MINUTES` to 5+ for production use
- Use specific JQL queries to reduce API load
- Monitor the log file size if running long-term

## 📊 System Requirements

- **Go 1.19+**
- **macOS** (for notifications - could be adapted for other OS)
- **Internet connection** to reach your Jira instance
- **Jira API access** (Cloud or Server)

## 🤝 Contributing

We'd love your help making this tool even better!

- 🐛 **Found a bug?** [Open an issue](https://github.com/vahiiiid/jira-new-ticket-notifier/issues)
- 💡 **Have an idea?** [Start a discussion](https://github.com/vahiiiid/jira-new-ticket-notifier/discussions)
- 🛠 **Want to contribute?** Fork, improve, and submit a PR!

### Development Setup
```bash
git clone https://github.com/vahiiiid/jira-new-ticket-notifier.git
cd jira-new-ticket-notifier
go mod tidy
# Create your .env file with the template above
go run main.go
```

## 📜 License

MIT License - feel free to use this in your own projects, modify it, or share it with your team!

## 🙏 Acknowledgments

- Built with ❤️ for developers who hate missing important tickets
- Inspired by the need for better Jira workflow automation
- Thanks to the Go community for excellent tooling

---

**Star ⭐ this repo if it helps you stay on top of your tickets!**