package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
	mu         sync.RWMutex
}

func NewProxyServer(listenAddr string, backendAddrs []string) *ProxyServer {
	backends := make([]*Backend, len(backendAddrs))
	for i, addr := range backendAddrs {
		backends[i] = &Backend{
			Address: addr,
			Healthy: false,
		}
	}

	return &ProxyServer{
		listenAddr: listenAddr,
		backends:   backends,
	}
}

func (p *ProxyServer) healthCheck(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
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
			conn, err := net.DialTimeout("tcp", b.Address, 3*time.Second)
			if err != nil {
				if b.IsHealthy() {
					log.Printf("[⚠] Backend %s is DOWN", b.Address)
				}
				b.SetHealthy(false)
				return
			}
			err = conn.Close()
			if err != nil {
				return
			}

			if !b.IsHealthy() {
				log.Printf("[✓]  Backend %s is UP", b.Address)
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
	defer func(clientConn net.Conn) {
		err := clientConn.Close()
		if err != nil {
			log.Printf("[✗] Error closing client connection: %v", err)
		}
	}(clientConn)

	backend := p.getHealthyBackend()
	if backend == nil {
		log.Printf("[✗] No healthy backends available for %s", clientConn.RemoteAddr())
		return
	}

	backendConn, err := net.DialTimeout("tcp", backend.Address, 5*time.Second)
	if err != nil {
		log.Printf("[✗] Failed to connect to backend %s: %v", backend.Address, err)
		backend.SetHealthy(false)
		return
	}
	defer func(backendConn net.Conn) {
		err := backendConn.Close()
		if err != nil {
			log.Printf("[✗] Error closing backend connection: %v", err)
		}
	}(backendConn)

	log.Printf("[C] %s -> %s", clientConn.RemoteAddr(), backend.Address)

	var wg sync.WaitGroup
	wg.Add(2)

	// client -> proxy -> backend
	go func() {
		defer wg.Done()
		_, err := io.Copy(clientConn, backendConn)
		if err != nil {
			log.Printf("[✗] Error copying data from backend to client: %v", err)
			return
		}
		err = clientConn.(*net.TCPConn).CloseWrite()
		if err != nil {
			log.Printf("[✗] Error closing client write: %v", err)
		}
	}()

	// backend -> proxy -> client
	go func() {
		defer wg.Done()
		_, err := io.Copy(clientConn, backendConn)
		if err != nil {
			log.Printf("[✗] Error copying data from backend to client: %v", err)
			return
		}

		err = clientConn.(*net.TCPConn).CloseWrite()
		if err != nil {
			log.Printf("[✗] Error closing client write: %v", err)
		}
	}()

	wg.Wait()
	log.Printf("[✓] Connection closed: %s", clientConn.RemoteAddr())
}

func (p *ProxyServer) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Printf("[✗] Error closing listener: %v", err)
		}
	}(listener)

	log.Printf("----------------------------------------")
	log.Printf("TCP Proxy running on %s", p.listenAddr)
	log.Printf("----------------------------------------")
	log.Printf("Backends:")
	for i, b := range p.backends {
		log.Printf("   %d. %s", i+1, b.Address)
	}
	log.Printf("----------------------------------------")

	go p.healthCheck(ctx)

	time.Sleep(1 * time.Second)

	go func() {
		<-ctx.Done()
		err := listener.Close()
		if err != nil {
			log.Printf("[✗] Error closing listener: %v", err)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("[✗] Error accepting connection: %v", err)
			}
		}

		go p.handleConnection(conn)
	}
}

func main() {
	listenAddr := flag.String("listen", ":25565", "Listen address (e.g., :25565)")
	backendsFlag := flag.String("backends", "", "Comma-separated list of backend servers (e.g., 'localhost:9001,localhost:9002')")
	flag.Parse()

	if *backendsFlag == "" {
		log.Fatal("[✗] Please specify backends with -backends flag")
	}

	backendAddrs := strings.Split(*backendsFlag, ",")
	for i := range backendAddrs {
		backendAddrs[i] = strings.TrimSpace(backendAddrs[i])
	}

	proxy := NewProxyServer(*listenAddr, backendAddrs)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := proxy.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
