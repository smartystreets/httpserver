package httpserver

import (
	"context"
	"net"
	"net/http"
	"syscall"
	"time"
)

type configuration struct {
	ctx             context.Context
	handler         http.Handler
	shutdownTimeout time.Duration
	listenAddress   string
	listenConfig    listenConfig
	logger          logger

	httpServer httpServer
}

func New(options ...option) ListenCloser {
	var config configuration
	Options.apply(options...)(&config)
	return newServer(config)
}

var Options singleton

type singleton struct{}
type option func(*configuration)

func (singleton) Context(value context.Context) option {
	return func(this *configuration) { this.ctx = value }
}
func (singleton) ListenAddress(value string) option {
	return func(this *configuration) { this.listenAddress = value }
}
func (singleton) Handler(value http.Handler) option {
	return func(this *configuration) { this.handler = value }
}
func (singleton) HTTPServer(value httpServer) option {
	return func(this *configuration) { this.httpServer = value }
}
func (singleton) ShutdownTimeout(value time.Duration) option {
	return func(this *configuration) { this.shutdownTimeout = value }
}
func (singleton) SocketConfig(value listenConfig) option {
	return func(this *configuration) { this.listenConfig = value }
}
func (singleton) Logger(value logger) option {
	return func(this *configuration) { this.logger = value }
}

func (singleton) apply(options ...option) option {
	return func(this *configuration) {
		for _, option := range Options.defaults(options...) {
			option(this)
		}

		if this.httpServer == nil {
			this.httpServer = &http.Server{Addr: this.listenAddress, Handler: this.handler}
		}
	}
}
func (singleton) defaults(options ...option) []option {
	const defaultShutdownTimeout = time.Second * 5
	const defaultListenAddress = ":http"

	var defaultContext = context.Background()
	var defaultHandler = nop{}
	var defaultLogger = nop{}
	var defaultListenConfig = &net.ListenConfig{Control: func(_, _ string, conn syscall.RawConn) error {
		return conn.Control(func(descriptor uintptr) {
			_ = syscall.SetsockoptInt(int(descriptor), syscall.SOL_SOCKET, socketReusePort, 1)
		})
	}}

	return append([]option{
		Options.ListenAddress(defaultListenAddress),
		Options.ShutdownTimeout(defaultShutdownTimeout),
		Options.Context(defaultContext),
		Options.Handler(defaultHandler),
		Options.Logger(defaultLogger),
		Options.SocketConfig(defaultListenConfig),
	}, options...)
}

type nop struct{}

func (nop) Printf(_ string, _ ...interface{}) {}

func (nop) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {}