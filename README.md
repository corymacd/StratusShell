# StratusShell

A web-based dual CLI session streaming application built with Go and GoTTY. This project provides a clean web interface that displays two terminal sessions stacked vertically, allowing you to interact with both simultaneously.

## Features

- 🖥️ **Dual Terminal Sessions**: Two independent CLI sessions running side-by-side (stacked)
- 🌐 **Web-Based Interface**: Access terminals through any modern web browser
- ⚡ **Real-Time Streaming**: Live terminal output using GoTTY's WebSocket technology
- ✏️ **Interactive**: Full input support for both terminals
- 🎨 **Modern UI**: Clean, professional interface with VS Code-inspired styling
- 🔄 **Auto-Reconnect**: Terminals automatically reconnect if connection is lost
- 🤖 **Claude Code Integration**: Configurable MCP servers for AI-powered development (Playwright, Linear, GitHub)

## Screenshots

![Dual CLI Sessions](https://github.com/user-attachments/assets/49013aeb-9e63-4f0c-b227-96955d9b8360)

*Two independent terminal sessions running in a web interface*

![Active Terminals](https://github.com/user-attachments/assets/9a25da22-766a-490b-af3d-c473d874ced4)

*Both terminals actively executing commands*

## Prerequisites

- Go 1.19 or later
- Bash (or sh) on Unix-like systems, cmd.exe on Windows

## Installation

1. Clone the repository:
```bash
git clone https://github.com/corymacd/StratusShell.git
cd StratusShell
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o stratusshell main.go
```

## Usage

1. Run the application:
```bash
./stratusshell
```

2. Open your web browser and navigate to:
```
http://localhost:8080
```

3. You'll see two terminal sessions:
   - **Terminal 1** on port 8081
   - **Terminal 2** on port 8082
   - Both embedded in the main interface on port 8080

4. Click on any terminal to start typing commands!

## Configuration

The application uses the following ports by default:
- **8080**: Main web interface
- **8081**: Terminal 1 (GoTTY instance)
- **8082**: Terminal 2 (GoTTY instance)

To modify ports, edit the constants in `main.go`:
```go
const (
    mainPort     = 8080
    terminal1Port = 8081
    terminal2Port = 8082
)
```

## Architecture

The application consists of:

1. **Main HTTP Server** (port 8080): Serves the HTML interface with embedded iframes
2. **GoTTY Server 1** (port 8081): Provides the first terminal session via WebSocket
3. **GoTTY Server 2** (port 8082): Provides the second terminal session via WebSocket

Each terminal runs in its own GoTTY instance with:
- Write permissions enabled
- Auto-reconnect after 10 seconds
- Full PTY support

## Graceful Shutdown

Press `Ctrl+C` to gracefully shutdown all servers. The application will:
1. Close the main HTTP server
2. Terminate both GoTTY instances
3. Clean up all resources

## Technologies Used

- **Go**: Backend server and orchestration
- **GoTTY**: Terminal streaming over WebSocket
- **HTML/CSS**: Frontend interface
- **WebSocket**: Real-time terminal communication

## License

This project is open source and available under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.