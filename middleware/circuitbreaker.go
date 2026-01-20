package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/llmx-ai/llmx"
)

// CircuitState represents the state of circuit breaker
type CircuitState int

const (
	// StateClosed means requests are allowed
	StateClosed CircuitState = iota
	// StateOpen means requests are blocked
	StateOpen
	// StateHalfOpen means testing if circuit can close
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu sync.RWMutex

	maxFailures      int           // Max failures before opening
	timeout          time.Duration // Timeout before trying half-open
	resetSuccesses   int           // Successes needed to close from half-open
	halfOpenRequests int           // Max requests allowed in half-open

	state            CircuitState
	failures         int
	successes        int
	lastFailureTime  time.Time
	halfOpenAttempts int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:      maxFailures,
		timeout:          timeout,
		resetSuccesses:   2,
		halfOpenRequests: 3,
		state:            StateClosed,
	}
}

// WithResetSuccesses sets the number of successes needed to close
func (cb *CircuitBreaker) WithResetSuccesses(n int) *CircuitBreaker {
	cb.resetSuccesses = n
	return cb
}

// WithHalfOpenRequests sets the max requests in half-open state
func (cb *CircuitBreaker) WithHalfOpenRequests(n int) *CircuitBreaker {
	cb.halfOpenRequests = n
	return cb
}

// Allow checks if the request should be allowed
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailureTime) > cb.timeout {
			// Transition to half-open
			cb.state = StateHalfOpen
			cb.halfOpenAttempts = 0
			return nil
		}
		return fmt.Errorf("circuit breaker is open")

	case StateHalfOpen:
		if cb.halfOpenAttempts >= cb.halfOpenRequests {
			return fmt.Errorf("circuit breaker is half-open, max attempts reached")
		}
		cb.halfOpenAttempts++
		return nil

	default:
		return fmt.Errorf("unknown circuit breaker state")
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures = 0

	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.resetSuccesses {
			// Close the circuit
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.halfOpenAttempts = 0
		}
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.maxFailures {
			// Open the circuit
			cb.state = StateOpen
		}

	case StateHalfOpen:
		// Go back to open state
		cb.state = StateOpen
		cb.successes = 0
		cb.halfOpenAttempts = 0
	}
}

// State returns the current state
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Stats returns circuit breaker statistics
func (cb *CircuitBreaker) Stats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":               cb.state.String(),
		"failures":            cb.failures,
		"successes":           cb.successes,
		"half_open_attempts":  cb.halfOpenAttempts,
		"last_failure_time":   cb.lastFailureTime,
	}
}

// String returns string representation of circuit state
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerMiddleware creates a circuit breaker middleware
func CircuitBreakerMiddleware(cb *CircuitBreaker) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			// Check if request is allowed
			if err := cb.Allow(); err != nil {
				return nil, llmx.NewInternalError(
					fmt.Sprintf("circuit breaker: %v", err),
					nil,
				)
			}

			// Execute request
			resp, err := next(ctx, req)

			if err != nil {
				// Record all errors for circuit breaker
				// (Circuit breaker should trip on any error, not just retryable ones)
				cb.RecordFailure()
				return nil, err
			}

			// Record success
			cb.RecordSuccess()
			return resp, nil
		}
	}
}

// PerModelCircuitBreaker manages circuit breakers per model
type PerModelCircuitBreaker struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
	factory  func() *CircuitBreaker
}

// NewPerModelCircuitBreaker creates a per-model circuit breaker
func NewPerModelCircuitBreaker(factory func() *CircuitBreaker) *PerModelCircuitBreaker {
	return &PerModelCircuitBreaker{
		breakers: make(map[string]*CircuitBreaker),
		factory:  factory,
	}
}

// GetBreaker returns the circuit breaker for a model
func (p *PerModelCircuitBreaker) GetBreaker(model string) *CircuitBreaker {
	p.mu.RLock()
	breaker, ok := p.breakers[model]
	p.mu.RUnlock()

	if ok {
		return breaker
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check
	if breaker, ok := p.breakers[model]; ok {
		return breaker
	}

	breaker = p.factory()
	p.breakers[model] = breaker
	return breaker
}

// CircuitBreakerPerModel creates a per-model circuit breaker middleware
func CircuitBreakerPerModel(pmcb *PerModelCircuitBreaker) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			breaker := pmcb.GetBreaker(req.Model)

			// Check if request is allowed
			if err := breaker.Allow(); err != nil {
				return nil, llmx.NewInternalError(
					fmt.Sprintf("circuit breaker for model %s: %v", req.Model, err),
					nil,
				)
			}

			// Execute request
			resp, err := next(ctx, req)

			if err != nil {
				// Record all errors for circuit breaker
				breaker.RecordFailure()
				return nil, err
			}

			breaker.RecordSuccess()
			return resp, nil
		}
	}
}
