package observability

import "github.com/prometheus/client_golang/prometheus"

// Metrics contains the bounded-cardinality operational signals used by Falzo.
// User IDs, room IDs, request IDs, and connection IDs must never be labels.
type Metrics struct {
	RealtimeConnections      prometheus.Gauge
	RealtimeConnectionsTotal *prometheus.CounterVec
	RealtimeDisconnects      *prometheus.CounterVec
	RealtimeReconnects       *prometheus.CounterVec
	RealtimeCommands         *prometheus.CounterVec
	RealtimeQueueDropped     *prometheus.CounterVec
	RealtimeRedisOperations  *prometheus.CounterVec
	RealtimeRedisDuration    *prometheus.HistogramVec
	GamePhaseTransitions     *prometheus.CounterVec
	GamePhaseDeadlineLag     prometheus.Histogram
	GameOverdueRounds        prometheus.Gauge
	AlertNotifications       *prometheus.CounterVec
	AlertQueueDepth          prometheus.Gauge
}

func NewMetrics(registerer prometheus.Registerer) *Metrics {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}
	m := &Metrics{
		RealtimeConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "falzo",
			Subsystem: "realtime",
			Name:      "connections",
			Help:      "Current number of active WebSocket connections.",
		}),
		RealtimeConnectionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "falzo",
			Subsystem: "realtime",
			Name:      "connections_total",
			Help:      "WebSocket connection attempts by result.",
		}, []string{"scope", "result"}),
		RealtimeDisconnects: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "falzo",
			Subsystem: "realtime",
			Name:      "disconnects_total",
			Help:      "WebSocket disconnections by bounded reason.",
		}, []string{"scope", "reason"}),
		RealtimeReconnects: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "falzo",
			Subsystem: "realtime",
			Name:      "reconnects_total",
			Help:      "Client-reported successful WebSocket reconnects.",
		}, []string{"scope"}),
		RealtimeCommands: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "falzo",
			Subsystem: "realtime",
			Name:      "commands_total",
			Help:      "WebSocket commands processed by command and result.",
		}, []string{"command", "result"}),
		RealtimeQueueDropped: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "falzo",
			Subsystem: "realtime",
			Name:      "queue_dropped_total",
			Help:      "Outbound WebSocket events rejected because a client queue was full.",
		}, []string{"event"}),
		RealtimeRedisOperations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "falzo",
			Subsystem: "realtime",
			Name:      "redis_operations_total",
			Help:      "Realtime Redis operations by operation and result.",
		}, []string{"operation", "result"}),
		RealtimeRedisDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "falzo",
			Subsystem: "realtime",
			Name:      "redis_duration_seconds",
			Help:      "Realtime Redis operation duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"operation"}),
		GamePhaseTransitions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "falzo",
			Subsystem: "game",
			Name:      "phase_transitions_total",
			Help:      "Server-driven game phase transitions by source, destination, and result.",
		}, []string{"from", "to", "result"}),
		GamePhaseDeadlineLag: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "falzo",
			Subsystem: "game",
			Name:      "phase_deadline_lag_seconds",
			Help:      "Delay between a phase deadline and its server-side transition.",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 30},
		}),
		GameOverdueRounds: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "falzo",
			Subsystem: "game",
			Name:      "overdue_rounds",
			Help:      "Rounds whose current phase deadline has elapsed.",
		}),
		AlertNotifications: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "falzo",
			Subsystem: "alerts",
			Name:      "notifications_total",
			Help:      "Error alert events by processing stage and result.",
		}, []string{"stage", "result"}),
		AlertQueueDepth: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "falzo",
			Subsystem: "alerts",
			Name:      "queue_depth",
			Help:      "Current number of error alerts waiting to be published to NATS JetStream.",
		}),
	}
	registerer.MustRegister(
		m.RealtimeConnections,
		m.RealtimeConnectionsTotal,
		m.RealtimeDisconnects,
		m.RealtimeReconnects,
		m.RealtimeCommands,
		m.RealtimeQueueDropped,
		m.RealtimeRedisOperations,
		m.RealtimeRedisDuration,
		m.GamePhaseTransitions,
		m.GamePhaseDeadlineLag,
		m.GameOverdueRounds,
		m.AlertNotifications,
		m.AlertQueueDepth,
	)
	return m
}
