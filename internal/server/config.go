package server

import "fmt"

type Config struct {
	Address   string
	HandlerV0 handler
}

func NewConfig(addr string, h0 handler) (*Config, error) {
	if len(addr) == 0 {
		return nil, fmt.Errorf("address is empty")
	}

	if h0 == nil {
		return nil, fmt.Errorf("handler is nil")
	}

	return &Config{Address: addr, HandlerV0: h0}, nil
}
