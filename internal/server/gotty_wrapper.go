package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/sorenisanerd/gotty/backend/localcommand"
	"github.com/sorenisanerd/gotty/server"
)

// GoTTYServer wraps a GoTTY server instance
type GoTTYServer struct {
	server     *server.Server
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewGoTTYServer creates and starts a new GoTTY server
func NewGoTTYServer(ctx context.Context, port int, credential, title, shell, workingDir string) (*GoTTYServer, error) {
	// Create options for GoTTY
	options := &server.Options{
		Address:          "localhost",
		Port:             fmt.Sprintf("%d", port),
		PermitWrite:      true,
		EnableReconnect:  true,
		ReconnectTime:    10,
		TitleFormat:      title,
		EnableBasicAuth:  false,
		Credential:       "",
	}

	// Parse credential for basic auth
	if credential != "" {
		options.Credential = credential
		options.EnableBasicAuth = true
	}

	// Build command with working directory
	command := shell
	var args []string
	if workingDir != "" {
		// Wrap shell to cd into working directory first
		command = "sh"
		args = []string{"-c", fmt.Sprintf("cd %s && exec %s", workingDir, shell)}
	}

	// Create factory for local command
	// Note: GoTTY's NewFactory requires a non-nil Options struct (even if empty)
	// to avoid nil pointer dereference in the library. Using &localcommand.Options{}
	// provides default values (CloseSignal=1/SIGHUP, CloseTimeout=-1).
	factory, err := localcommand.NewFactory(command, args, &localcommand.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create command factory: %w", err)
	}

	// Create server
	srv, err := server.New(factory, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create gotty server: %w", err)
	}

	// Create context for server lifecycle
	serverCtx, cancel := context.WithCancel(ctx)

	// Start server in goroutine
	go func() {
		if err := srv.Run(serverCtx, server.WithGracefullContext(serverCtx)); err != nil {
			if err != http.ErrServerClosed && serverCtx.Err() == nil {
				log.Printf("GoTTY server error on port %d: %v", port, err)
			}
		}
	}()

	return &GoTTYServer{
		server:     srv,
		ctx:        serverCtx,
		cancelFunc: cancel,
	}, nil
}

// Stop gracefully stops the GoTTY server
func (g *GoTTYServer) Stop() error {
	g.cancelFunc()

	// Wait briefly for graceful shutdown
	<-g.ctx.Done()

	return nil
}
