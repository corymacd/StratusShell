package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/sorenisanerd/gotty/backend/localcommand"
	"github.com/sorenisanerd/gotty/server"
)

const (
	mainPort     = 8080
	terminal1Port = 8081
	terminal2Port = 8082
)

func main() {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start GoTTY servers for two terminals
	go startGoTTYServer(ctx, terminal1Port, "Terminal 1")
	go startGoTTYServer(ctx, terminal2Port, "Terminal 2")

	// Give GoTTY servers time to start
	time.Sleep(2 * time.Second)

	// Start main HTTP server
	http.HandleFunc("/", serveMainPage)
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", mainPort),
		Handler: http.DefaultServeMux,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Main web interface starting on http://localhost:%d", mainPort)
		log.Printf("Terminal 1 on port %d", terminal1Port)
		log.Printf("Terminal 2 on port %d", terminal2Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down gracefully...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	cancel() // Cancel context for GoTTY servers
}

func startGoTTYServer(ctx context.Context, port int, title string) {
	// Determine shell command based on OS
	var shellCmd []string
	if runtime.GOOS == "windows" {
		shellCmd = []string{"cmd.exe"}
	} else {
		// Try to use bash, fall back to sh
		if _, err := exec.LookPath("bash"); err == nil {
			shellCmd = []string{"bash"}
		} else {
			shellCmd = []string{"sh"}
		}
	}

	appOptions := &server.Options{
		Address:             "localhost",
		Port:                fmt.Sprintf("%d", port),
		PermitWrite:         true,
		EnableReconnect:     true,
		ReconnectTime:       10,
		TitleVariables:      map[string]interface{}{"title": title},
		EnableTLSClientAuth: false,
	}

	backendOptions := &localcommand.Options{
		CloseSignal: 1, // SIGHUP
	}

	factory, err := localcommand.NewFactory(shellCmd[0], shellCmd[1:], backendOptions)
	if err != nil {
		log.Printf("Failed to create command factory for %s: %v", title, err)
		return
	}

	srv, err := server.New(factory, appOptions)
	if err != nil {
		log.Printf("Failed to create GoTTY server for %s: %v", title, err)
		return
	}

	// Run server
	err = srv.Run(ctx, server.WithGracefullContext(ctx))
	if err != nil {
		log.Printf("GoTTY server error for %s: %v", title, err)
	}
}

func serveMainPage(w http.ResponseWriter, r *http.Request) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dual CLI Sessions</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background-color: #1e1e1e;
            color: #ffffff;
            display: flex;
            flex-direction: column;
            height: 100vh;
            overflow: hidden;
        }
        header {
            background-color: #252526;
            padding: 15px 20px;
            border-bottom: 1px solid #3e3e42;
            text-align: center;
        }
        h1 {
            font-size: 24px;
            font-weight: 500;
        }
        .container {
            display: flex;
            flex-direction: column;
            flex: 1;
            overflow: hidden;
        }
        .terminal-wrapper {
            flex: 1;
            display: flex;
            flex-direction: column;
            border-bottom: 2px solid #3e3e42;
            position: relative;
            overflow: hidden;
        }
        .terminal-wrapper:last-child {
            border-bottom: none;
        }
        .terminal-header {
            background-color: #2d2d30;
            padding: 8px 15px;
            border-bottom: 1px solid #3e3e42;
            font-size: 14px;
            font-weight: 500;
            color: #cccccc;
        }
        .terminal-frame {
            flex: 1;
            border: none;
            width: 100%%;
            height: 100%%;
            background-color: #1e1e1e;
        }
        .status {
            background-color: #007acc;
            color: white;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 11px;
            margin-left: 10px;
        }
    </style>
</head>
<body>
    <header>
        <h1>🖥️ Dual CLI Sessions</h1>
    </header>
    <div class="container">
        <div class="terminal-wrapper">
            <div class="terminal-header">
                Terminal 1 <span class="status">Active</span>
            </div>
            <iframe class="terminal-frame" src="http://localhost:%d/"></iframe>
        </div>
        <div class="terminal-wrapper">
            <div class="terminal-header">
                Terminal 2 <span class="status">Active</span>
            </div>
            <iframe class="terminal-frame" src="http://localhost:%d/"></iframe>
        </div>
    </div>
</body>
</html>`, terminal1Port, terminal2Port)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}
