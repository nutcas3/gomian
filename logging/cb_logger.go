package logging

import (
	"go.uber.org/zap"
	"time"
)

type CircuitBreakerLogger struct {
	logger *zap.Logger
}

func NewCircuitBreakerLogger() *CircuitBreakerLogger {
	return &CircuitBreakerLogger{
		logger: Logger().With(zap.String("component", "circuit_breaker")),
	}
}

func (l *CircuitBreakerLogger) LogStateChange(name string, from, to string) {
	l.logger.Info("circuit breaker state changed",
		zap.String("circuit", name),
		zap.String("from_state", from),
		zap.String("to_state", to),
		zap.Time("timestamp", time.Now()),
	)
}

func (l *CircuitBreakerLogger) LogTrip(name string, err error) {
	fields := []zap.Field{
		zap.String("circuit", name),
		zap.Time("timestamp", time.Now()),
	}
	
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	
	l.logger.Warn("circuit breaker tripped", fields...)
}

func (l *CircuitBreakerLogger) LogReset(name string) {
	l.logger.Info("circuit breaker reset",
		zap.String("circuit", name),
		zap.Time("timestamp", time.Now()),
	)
}

func (l *CircuitBreakerLogger) LogSuccess(name string) {
	l.logger.Debug("circuit breaker request succeeded",
		zap.String("circuit", name),
		zap.Time("timestamp", time.Now()),
	)
}

func (l *CircuitBreakerLogger) LogFailure(name string, err error) {
	fields := []zap.Field{
		zap.String("circuit", name),
		zap.Time("timestamp", time.Now()),
	}
	
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	
	l.logger.Debug("circuit breaker request failed", fields...)
}

func (l *CircuitBreakerLogger) LogRejection(name string) {
	l.logger.Debug("circuit breaker request rejected",
		zap.String("circuit", name),
		zap.Time("timestamp", time.Now()),
	)
}

func (l *CircuitBreakerLogger) LogMetrics(name string, state string, totalRequests uint64, 
	totalFailures uint64, consecutiveFailures uint64, consecutiveSuccesses uint64, 
	timeInState time.Duration) {
	
	l.logger.Debug("circuit breaker metrics",
		zap.String("circuit", name),
		zap.String("state", state),
		zap.Uint64("total_requests", totalRequests),
		zap.Uint64("total_failures", totalFailures),
		zap.Uint64("consecutive_failures", consecutiveFailures),
		zap.Uint64("consecutive_successes", consecutiveSuccesses),
		zap.Duration("time_in_state", timeInState),
		zap.Time("timestamp", time.Now()),
	)
}
