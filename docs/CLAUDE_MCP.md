# Claude Code and MCP Server Configuration

This document describes how StratusShell integrates with Claude Code and Model Context Protocol (MCP) servers.

## Overview

StratusShell can automatically configure Claude Code settings and install MCP servers during user provisioning. This enables AI-powered development workflows with access to browser automation, project management, and repository operations.

## What are MCP Servers?

Model Context Protocol (MCP) servers are specialized tools that extend Claude Code's capabilities by providing:

- **Data access**: Connect to external APIs and services
- **Tool execution**: Perform automated tasks like browser testing
- **Context enrichment**: Provide domain-specific knowledge

## Supported MCP Servers

StratusShell includes built-in support for three MCP servers:

### 1. Playwright MCP (@playwright/mcp)
**Purpose**: Browser automation and testing

**Capabilities**:
- Automated browser interactions
- Web scraping and testing
- Screenshot and PDF generation
- Cross-browser testing

**Use cases**:
- Writing end-to-end tests
- Debugging web applications
- Automated UI testing

### 2. Linear MCP (@mseep/linear-mcp)
**Purpose**: Linear project management integration

**Capabilities**:
- Create and update issues
- Query project status
- Manage teams and workflows
- Track progress

**Use cases**:
- Issue tracking from your terminal
- Automated project updates
- Workflow automation

### 3. GitHub MCP (github-mcp-server)
**Purpose**: GitHub API integration

**Capabilities**:
- Repository management
- Pull request operations
- Issue tracking
- Code review automation

**Use cases**:
- Automated repository tasks
- PR management
- Issue triage and labeling

## Configuration

### Configuration File Structure

MCP servers are configured in `/etc/stratusshell/default.yaml`:

```yaml
claude:
  enabled: true
  allow:
    - gh
  deny: []
  ask: []
  mcp_servers:
    - name: "playwright"
      package: "@playwright/mcp"
      command: "npx"
      args:
        - "-y"
        - "@playwright/mcp"
      env: {}
    - name: "github"
      package: "github-mcp-server"
      command: "npx"
      args:
        - "-y"
        - "github-mcp-server"
      env: {}
    - name: "linear"
      package: "@mseep/linear-mcp"
      command: "npx"
      args:
        - "-y"
        - "@mseep/linear-mcp"
      env: {}
```

### Configuration Fields

Each MCP server entry requires:

- **name**: Unique identifier for the server
- **package**: npm package name to install
- **command**: Command to execute the server (typically `npx`)
- **args**: Command-line arguments to pass to the server
- **env**: Environment variables (optional)

### Generated Settings File

During provisioning, StratusShell creates `~/.claude/settings.json`:

```json
{
  "permissions": {
    "allow": ["gh"],
    "deny": [],
    "ask": []
  },
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["-y", "@playwright/mcp"]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "github-mcp-server"]
    },
    "linear": {
      "command": "npx",
      "args": ["-y", "@mseep/linear-mcp"]
    }
  }
}
```

## Provisioning

### Automatic Installation

MCP servers are installed automatically during user provisioning:

```bash
sudo stratusshell init --user=developer
```

The provisioning process:
1. Creates the user account
2. Installs base packages and tools
3. Configures Claude Code settings
4. Installs MCP servers globally via npm
5. Sets up the `.claude/settings.json` file

### Manual Installation

To install MCP servers manually:

```bash
# Install individual MCP servers
npm install -g @playwright/mcp
npm install -g github-mcp-server
npm install -g @mseep/linear-mcp
```

Then manually create or update `~/.claude/settings.json`.

## Customization

### Adding Custom MCP Servers

To add a custom MCP server, edit `/etc/stratusshell/default.yaml`:

```yaml
claude:
  mcp_servers:
    - name: "custom-server"
      package: "my-custom-mcp"
      command: "npx"
      args:
        - "-y"
        - "my-custom-mcp"
      env:
        API_KEY: "${MY_API_KEY}"
```

### Disabling MCP Servers

To disable a specific MCP server, remove its entry from the `mcp_servers` list in the configuration.

To disable all MCP functionality:

```yaml
claude:
  enabled: false
```

### Environment Variables

MCP servers can use environment variables for configuration:

```yaml
claude:
  mcp_servers:
    - name: "github"
      package: "github-mcp-server"
      command: "npx"
      args:
        - "-y"
        - "github-mcp-server"
      env:
        GITHUB_TOKEN: "${GITHUB_TOKEN}"
```

## Permissions

The `permissions` section controls which commands Claude Code can execute:

```yaml
claude:
  allow:
    - gh      # Allow GitHub CLI
    - npm     # Allow npm commands
  deny:
    - rm      # Deny file deletion
    - dd      # Deny disk operations
  ask:
    - git     # Prompt before git operations
```

- **allow**: Commands that can be executed without prompting
- **deny**: Commands that are explicitly blocked
- **ask**: Commands that require user confirmation

## Security Considerations

1. **Package Sources**: MCP servers are installed from npm. Verify package authenticity before adding custom servers.

2. **Environment Variables**: Avoid storing secrets in configuration files. Use environment variable substitution instead.

3. **Command Permissions**: Use the `deny` list to block dangerous commands.

4. **Network Access**: MCP servers may make network requests. Ensure your firewall rules are appropriate.

## Troubleshooting

### MCP Server Installation Fails

**Problem**: npm install fails during provisioning

**Solutions**:
- Check internet connectivity
- Verify npm is installed: `which npm`
- Try manual installation: `npm install -g <package>`
- Check npm logs: `npm config get prefix`

### Claude Code Doesn't Recognize MCP Servers

**Problem**: MCP servers not showing in Claude Code

**Solutions**:
- Verify `~/.claude/settings.json` exists and is valid JSON
- Check file permissions: `ls -la ~/.claude/settings.json`
- Restart Claude Code
- Check MCP server is installed: `npm list -g <package>`

### MCP Server Command Not Found

**Problem**: `npx` command not found when running MCP server

**Solutions**:
- Ensure Node.js/npm is in PATH
- Check npm global bin directory: `npm config get prefix`
- Add npm bin to PATH: `export PATH="$(npm config get prefix)/bin:$PATH"`

## Development

### Testing MCP Configuration

To test MCP server configuration without full provisioning:

```bash
# Test JSON marshaling
go test ./internal/provision -v -run TestClaudeConfigWithMCPServers

# Test YAML loading
go test ./internal/provision -v -run TestLoadConfigWithMCPServers
```

### Adding New MCP Server Support

1. Add the server to `configs/default.yaml`:
   ```yaml
   mcp_servers:
     - name: "new-server"
       package: "new-mcp-package"
       command: "npx"
       args: ["-y", "new-mcp-package"]
   ```

2. Update `.claude/settings.json` with the new server configuration

3. Test the configuration:
   ```bash
   go test ./internal/provision
   ```

## References

- [Model Context Protocol Documentation](https://modelcontextprotocol.io/)
- [Claude Code Documentation](https://docs.anthropic.com/)
- [Playwright MCP](https://www.npmjs.com/package/@playwright/mcp)
- [Linear MCP](https://www.npmjs.com/package/@mseep/linear-mcp)
- [GitHub MCP Server](https://www.npmjs.com/package/github-mcp-server)

## Support

For issues with:
- **StratusShell MCP integration**: Open an issue on this repository
- **MCP servers themselves**: Contact the respective package maintainers
- **Claude Code**: Refer to Anthropic's support channels
