package services

import (
	"log"
	"runtime"
	"time"

	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/utils"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemMetricsService handles system metrics collection
type SystemMetricsService struct {
	startTime time.Time
}

// NewSystemMetricsService creates a new system metrics service
func NewSystemMetricsService() *SystemMetricsService {
	return &SystemMetricsService{
		startTime: time.Now(),
	}
}

// SystemMetric represents a system metric record
type SystemMetric struct {
	ID         string    `json:"id"`
	MetricType string    `json:"metric_type"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Metadata   string    `json:"metadata,omitempty"`
	RecordedAt time.Time `json:"recorded_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// CollectMetrics collects all system metrics and stores them in database
func (s *SystemMetricsService) CollectMetrics() error {
	db := config.DB

	// Collect CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		metric := SystemMetric{
			ID:         utils.GenerateID(),
			MetricType: "cpu",
			Value:      cpuPercent[0],
			Unit:       "percent",
			RecordedAt: time.Now(),
			CreatedAt:  time.Now(),
		}
		db.Table("system_metrics").Create(&metric)
	}

	// Collect memory usage
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metric := SystemMetric{
			ID:         utils.GenerateID(),
			MetricType: "memory",
			Value:      memInfo.UsedPercent,
			Unit:       "percent",
			RecordedAt: time.Now(),
			CreatedAt:  time.Now(),
		}
		db.Table("system_metrics").Create(&metric)
	}

	// Collect disk usage
	diskInfo, err := disk.Usage("/")
	if err == nil {
		metric := SystemMetric{
			ID:         utils.GenerateID(),
			MetricType: "disk",
			Value:      diskInfo.UsedPercent,
			Unit:       "percent",
			RecordedAt: time.Now(),
			CreatedAt:  time.Now(),
		}
		db.Table("system_metrics").Create(&metric)
	}

	// Collect network I/O
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		// Bytes sent
		metricSent := SystemMetric{
			ID:         utils.GenerateID(),
			MetricType: "network_sent",
			Value:      float64(netIO[0].BytesSent),
			Unit:       "bytes",
			RecordedAt: time.Now(),
			CreatedAt:  time.Now(),
		}
		db.Table("system_metrics").Create(&metricSent)

		// Bytes received
		metricRecv := SystemMetric{
			ID:         utils.GenerateID(),
			MetricType: "network_received",
			Value:      float64(netIO[0].BytesRecv),
			Unit:       "bytes",
			RecordedAt: time.Now(),
			CreatedAt:  time.Now(),
		}
		db.Table("system_metrics").Create(&metricRecv)
	}

	return nil
}

// GetLatestMetrics returns the latest metrics for each type
func (s *SystemMetricsService) GetLatestMetrics() (map[string]interface{}, error) {
	db := config.DB

	metrics := make(map[string]interface{})

	// Get latest CPU
	var cpuMetric SystemMetric
	if err := db.Table("system_metrics").
		Where("metric_type = ?", "cpu").
		Order("recorded_at DESC").
		First(&cpuMetric).Error; err == nil {
		metrics["cpu_usage"] = cpuMetric.Value
	} else {
		metrics["cpu_usage"] = 0.0
	}

	// Get latest memory
	var memMetric SystemMetric
	if err := db.Table("system_metrics").
		Where("metric_type = ?", "memory").
		Order("recorded_at DESC").
		First(&memMetric).Error; err == nil {
		metrics["memory_usage"] = memMetric.Value
	} else {
		metrics["memory_usage"] = 0.0
	}

	// Get latest disk
	var diskMetric SystemMetric
	if err := db.Table("system_metrics").
		Where("metric_type = ?", "disk").
		Order("recorded_at DESC").
		First(&diskMetric).Error; err == nil {
		metrics["disk_usage"] = diskMetric.Value
	} else {
		metrics["disk_usage"] = 0.0
	}

	// Get network I/O
	var netSent, netRecv SystemMetric
	if err := db.Table("system_metrics").
		Where("metric_type = ?", "network_sent").
		Order("recorded_at DESC").
		First(&netSent).Error; err == nil {
		if err := db.Table("system_metrics").
			Where("metric_type = ?", "network_received").
			Order("recorded_at DESC").
			First(&netRecv).Error; err == nil {
			metrics["network_io"] = map[string]interface{}{
				"bytes_sent":     netSent.Value,
				"bytes_received": netRecv.Value,
			}
		}
	}

	// Calculate uptime
	uptime := time.Since(s.startTime)
	metrics["uptime"] = uptime.String()
	metrics["uptime_seconds"] = uptime.Seconds()

	return metrics, nil
}

// GetMetricsHistory returns metrics history for a time range
func (s *SystemMetricsService) GetMetricsHistory(metricType string, hours int) ([]SystemMetric, error) {
	db := config.DB

	var metrics []SystemMetric
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	err := db.Table("system_metrics").
		Where("metric_type = ? AND recorded_at >= ?", metricType, since).
		Order("recorded_at ASC").
		Find(&metrics).Error

	return metrics, err
}

// CheckServiceHealth checks the health of a service
func (s *SystemMetricsService) CheckServiceHealth(serviceName string) (string, error) {
	db := config.DB

	switch serviceName {
	case "database":
		// Check database connection
		sqlDB, err := db.DB()
		if err != nil {
			return "unhealthy", err
		}
		if err := sqlDB.Ping(); err != nil {
			return "unhealthy", err
		}
		return "healthy", nil

	case "api_server":
		// API server is healthy if we can execute this code
		return "healthy", nil

	default:
		return "unknown", nil
	}
}

// UpdateServiceStatus updates the status of a service in database
func (s *SystemMetricsService) UpdateServiceStatus(serviceName, status string, responseTime float64) error {
	db := config.DB

	update := map[string]interface{}{
		"status":           status,
		"response_time_ms": responseTime,
		"last_check_at":    time.Now(),
		"updated_at":       time.Now(),
	}

	return db.Table("system_services").
		Where("service_name = ?", serviceName).
		Updates(update).Error
}

// StartMetricsCollection starts periodic metrics collection
func (s *SystemMetricsService) StartMetricsCollection(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := s.CollectMetrics(); err != nil {
				log.Printf("Error collecting metrics: %v", err)
			}

			// Check service health
			services := []string{"database", "api_server"}
			for _, service := range services {
				start := time.Now()
				status, _ := s.CheckServiceHealth(service)
				responseTime := time.Since(start).Milliseconds()
				
				if err := s.UpdateServiceStatus(service, status, float64(responseTime)); err != nil {
					log.Printf("Error updating service status: %v", err)
				}
			}
		}
	}()
	log.Printf("System metrics collection started (interval: %v)", interval)
}

// CleanupOldMetrics removes metrics older than specified days
func (s *SystemMetricsService) CleanupOldMetrics(days int) error {
	db := config.DB

	cutoff := time.Now().AddDate(0, 0, -days)

	result := db.Table("system_metrics").
		Where("recorded_at < ?", cutoff).
		Delete(&SystemMetric{})

	log.Printf("Cleaned up %d old metrics records", result.RowsAffected)
	return result.Error
}

// GetGoRuntimeStats returns Go runtime statistics
func (s *SystemMetricsService) GetGoRuntimeStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"goroutines":     runtime.NumGoroutine(),
		"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":         float64(m.Sys) / 1024 / 1024,
		"num_gc":         m.NumGC,
		"gc_pause_ns":    m.PauseNs[(m.NumGC+255)%256],
	}
}