package scanner_test

import (
	"context"
	"net"
	"testing"
	"time"

	"my-scaner/pkg/scanner"
)

func TestScanPort_Open(t *testing.T) {
	// Cria um listener TCP local temporário para simular porta aberta
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("falha ao iniciar listener local: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	s := scanner.New()
	ctx := context.Background()
	result := s.ScanPort(ctx, scanner.Target{Host: "127.0.0.1", Port: addr.Port}, 1*time.Second)

	if !result.Open {
		t.Errorf("esperava porta aberta, obteve fechada: %v", result.Error)
	}
}

func TestScanPort_Closed(t *testing.T) {
	// Porta improvável de estar em uso ou escuta
	s := scanner.New()
	ctx := context.Background()
	result := s.ScanPort(ctx, scanner.Target{Host: "127.0.0.1", Port: 59999}, 200*time.Millisecond)

	if result.Open {
		t.Errorf("esperava porta fechada, obteve aberta")
	}
}
