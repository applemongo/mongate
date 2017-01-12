package gate

import (
	"log"
	"os"
	"time"

	"github.com/rcrowley/go-metrics"
)

var (
	tcpLiveConnections metrics.Counter
	tcpRequestTimer    metrics.Timer
)

func init() {
	tcpLiveConnections = metrics.NewCounter()
	metrics.Register("tcp_live_connections", tcpLiveConnections)

	tcpRequestTimer = metrics.NewTimer()
	metrics.Register("tcp_requests_timer", tcpRequestTimer)
}

func CollectStats() {
	metrics.Log(metrics.DefaultRegistry, 60*time.Second, log.New(os.Stderr, "[stats] ", log.LstdFlags))
}
