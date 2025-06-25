package groprometheus

import (
	"sync/atomic"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Custom registry
	registry = prometheus.NewRegistry()
	hasInit  = atomic.Bool{}

	// Simple metric to count Discord servers
	discordServersGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "grocer_bot_discord_servers",
			Help: "Number of Discord servers the bot is in",
		},
	)

	// Counter to track bot command invocations
	commandInvocationCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grocer_bot_command_invocations_total",
			Help: "Total number of bot command invocations",
		},
		[]string{"command_name"},
	)
)

// PrometheusHandler returns the Prometheus metrics endpoint
func PrometheusHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Use our custom registry
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

// UpdateDiscordServers updates the Discord servers count
func UpdateDiscordServers(count int) {
	discordServersGauge.Set(float64(count))
}

// IncrementCommandInvocation increments the counter for a specific command
func IncrementCommandInvocation(commandName string) {
	commandInvocationCounter.WithLabelValues(commandName).Inc()
}

// InitMetrics initializes and registers all metrics
func InitMetrics() {
	if hasInit.Load() {
		return
	}
	hasInit.Store(true)

	// Register the metrics with our custom registry
	registry.MustRegister(discordServersGauge)
	registry.MustRegister(commandInvocationCounter)
}
