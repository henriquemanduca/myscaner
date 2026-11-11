package scanner

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Target representa o destino de uma varredura.
type Target struct {
	Host string
	Port int
}

// Result representa o resultado da verificação de um target.
type Result struct {
	Target Target
	Open   bool
	Error  error
}

// Scanner define a interface para varredura de portas.
// Segue o princípio de inversão de dependências (DIP) e facilita mocks nos testes.
type Scanner interface {
	ScanPort(ctx context.Context, target Target, timeout time.Duration) Result
}

type defaultScanner struct{}

// New instancia a implementação padrão de Scanner.
func New() Scanner {
	return &defaultScanner{}
}

// ScanPort tenta estabelecer uma conexão TCP com o target.
func (s *defaultScanner) ScanPort(ctx context.Context, target Target, timeout time.Duration) Result {
	address := fmt.Sprintf("%s:%d", target.Host, target.Port)
	dialer := net.Dialer{Timeout: timeout}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return Result{
			Target: target,
			Open:   false,
			Error:  err,
		}
	}
	_ = conn.Close()

	return Result{
		Target: target,
		Open:   true,
		Error:  nil,
	}
}
