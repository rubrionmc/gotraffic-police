package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Backend struct {
	Address string
	Healthy bool
	mu      sync.RWMutex
}

func (b *Backend) IsHealthy() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Healthy
}

func (b *Backend) SetHealthy(healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Healthy = healthy
}

type ProxyServer struct {
	listenAddr string
	backends   []*Backend
	cfg        *Config
	mu         sync.RWMutex
}

func NewProxyServer(cfg *Config) *ProxyServer {
	backends := make([]*Backend, len(cfg.Backends.Servers))
	for i, addr := range cfg.Backends.Servers {
		backends[i] = &Backend{
			Address: addr,
			Healthy: false,
		}
	}

	return &ProxyServer{
		listenAddr: cfg.Server.Listen,
		backends:   backends,
		cfg:        cfg,
	}
}

func (p *ProxyServer) healthCheck(ctx context.Context) {
	ticker := time.NewTicker(p.cfg.Timeouts.HealthcheckInterval)
	defer ticker.Stop()

	p.checkAllBackends()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.checkAllBackends()
		}
	}
}

func (p *ProxyServer) checkAllBackends() {
	for _, backend := range p.backends {
		go func(b *Backend) {
			conn, err := net.DialTimeout(
				"tcp",
				b.Address,
				p.cfg.Timeouts.HealthcheckDial,
			)
			if err != nil {
				if b.IsHealthy() {
					log.Printf("[⚠] Backend %s DOWN", b.Address)
				}
				b.SetHealthy(false)
				return
			}
			_ = conn.Close()

			if !b.IsHealthy() {
				log.Printf("[✓] Backend %s UP", b.Address)
			}
			b.SetHealthy(true)
		}(backend)
	}
}

func (p *ProxyServer) getHealthyBackend() *Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, backend := range p.backends {
		if backend.IsHealthy() {
			return backend
		}
	}
	return nil
}

func (p *ProxyServer) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	backend := p.getHealthyBackend()
	if backend == nil {
		log.Printf("[✗] No healthy backend for %s", clientConn.RemoteAddr())
		return
	}

	backendConn, err := net.DialTimeout(
		"tcp",
		backend.Address,
		p.cfg.Timeouts.BackendDial,
	)
	if err != nil {
		log.Printf("[✗] Backend connect failed %s: %v", backend.Address, err)
		backend.SetHealthy(false)
		return
	}
	defer backendConn.Close()

	log.Printf("[C] %s -> %s", clientConn.RemoteAddr(), backend.Address)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(backendConn, clientConn)
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(clientConn, backendConn)
	}()

	wg.Wait()
	log.Printf("[✓] Connection closed %s", clientConn.RemoteAddr())
}

func (p *ProxyServer) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}
	defer listener.Close()

	log.Printf("[i] TCP Proxy listening on %s", p.listenAddr)
	for i, b := range p.backends {
		log.Printf("[i] Backend %d: %s", i+1, b.Address)
	}

	go p.healthCheck(ctx)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("[✗] Accept error: %v", err)
				continue
			}
		}

		go p.handleConnection(conn)
	}
}

func main() {
	configPath := flag.String("c", "", "Config file path (default: ./config.toml)")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	if len(cfg.Backends.Servers) == 0 {
		log.Fatal("[✗] No backends defined")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proxy := NewProxyServer(cfg)
	if err := proxy.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
