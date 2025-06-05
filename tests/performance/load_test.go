// Package performance 性能和负载测试
// 提供HTTP API的性能测试、负载测试和压力测试功能
package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/workflow-engine/workflow-engine/internal/server"
)

// LoadTestConfig 负载测试配置
type LoadTestConfig struct {
	ConcurrentUsers int           // 并发用户数
	TestDuration    time.Duration // 测试持续时间
	RampUpTime      time.Duration // 加压时间
	RequestTimeout  time.Duration // 请求超时时间
}

// LoadTestResult 负载测试结果
type LoadTestResult struct {
	TotalRequests     int64         // 总请求数
	SuccessRequests   int64         // 成功请求数
	FailedRequests    int64         // 失败请求数
	TotalDuration     time.Duration // 总耗时
	AvgResponseTime   time.Duration // 平均响应时间
	MinResponseTime   time.Duration // 最小响应时间
	MaxResponseTime   time.Duration // 最大响应时间
	RequestsPerSecond float64       // 每秒请求数 (RPS)
	ErrorRate         float64       // 错误率
}

// PerformanceTestSuite 性能测试套件
type PerformanceTestSuite struct {
	server *httptest.Server
	router *server.Router
	logger *zap.Logger
}

// NewPerformanceTestSuite 创建性能测试套件
func NewPerformanceTestSuite() *PerformanceTestSuite {
	logger, _ := zap.NewDevelopment()
	router := server.NewRouter(logger)
	server := httptest.NewServer(router)

	return &PerformanceTestSuite{
		server: server,
		router: router,
		logger: logger,
	}
}

// Close 关闭测试套件
func (pts *PerformanceTestSuite) Close() {
	if pts.server != nil {
		pts.server.Close()
	}
	if pts.logger != nil {
		pts.logger.Sync()
	}
}

// makeHTTPRequest 发送HTTP请求并记录性能指标
func (pts *PerformanceTestSuite) makeHTTPRequest(method, path string, body interface{}) (int, time.Duration, error) {
	start := time.Now()

	var reqBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&reqBody).Encode(body); err != nil {
			return 0, 0, fmt.Errorf("编码请求体失败: %w", err)
		}
	}

	req, err := http.NewRequest(method, pts.server.URL+path, &reqBody)
	if err != nil {
		return 0, 0, fmt.Errorf("创建请求失败: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	return resp.StatusCode, duration, nil
}

// runLoadTest 执行负载测试
func (pts *PerformanceTestSuite) runLoadTest(config LoadTestConfig, requestFunc func() (int, time.Duration, error)) *LoadTestResult {
	result := &LoadTestResult{
		MinResponseTime: time.Hour, // 初始化为一个很大的值
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	// 记录开始时间
	startTime := time.Now()

	// 结果收集器
	responseTimes := make([]time.Duration, 0)

	// 创建并发用户
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// 加压期间逐步启动用户
			if config.RampUpTime > 0 {
				delay := time.Duration(userID) * config.RampUpTime / time.Duration(config.ConcurrentUsers)
				time.Sleep(delay)
			}

			userStartTime := time.Now()
			for time.Since(userStartTime) < config.TestDuration {
				statusCode, responseTime, err := requestFunc()

				mu.Lock()
				result.TotalRequests++
				if err != nil || statusCode >= 400 {
					result.FailedRequests++
				} else {
					result.SuccessRequests++
				}

				responseTimes = append(responseTimes, responseTime)

				// 更新响应时间统计
				if responseTime < result.MinResponseTime {
					result.MinResponseTime = responseTime
				}
				if responseTime > result.MaxResponseTime {
					result.MaxResponseTime = responseTime
				}
				mu.Unlock()

				// 稍微休息一下，避免过度压力
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// 计算总耗时
	result.TotalDuration = time.Since(startTime)

	// 计算平均响应时间
	if len(responseTimes) > 0 {
		var totalTime time.Duration
		for _, rt := range responseTimes {
			totalTime += rt
		}
		result.AvgResponseTime = totalTime / time.Duration(len(responseTimes))
	}

	// 计算RPS
	if result.TotalDuration > 0 {
		result.RequestsPerSecond = float64(result.TotalRequests) / result.TotalDuration.Seconds()
	}

	// 计算错误率
	if result.TotalRequests > 0 {
		result.ErrorRate = float64(result.FailedRequests) / float64(result.TotalRequests) * 100
	}

	return result
}

// TestHealthCheckPerformance 测试健康检查API性能
func TestHealthCheckPerformance(t *testing.T) {
	pts := NewPerformanceTestSuite()
	defer pts.Close()

	pts.logger.Info("开始健康检查API性能测试")

	config := LoadTestConfig{
		ConcurrentUsers: 50,
		TestDuration:    30 * time.Second,
		RampUpTime:      5 * time.Second,
		RequestTimeout:  5 * time.Second,
	}

	requestFunc := func() (int, time.Duration, error) {
		return pts.makeHTTPRequest("GET", "/health", nil)
	}

	result := pts.runLoadTest(config, requestFunc)

	// 验证性能指标
	assert.Greater(t, result.TotalRequests, int64(0), "应该有请求被发送")
	assert.Greater(t, result.SuccessRequests, int64(0), "应该有成功的请求")
	assert.Less(t, result.ErrorRate, 5.0, "错误率应该低于5%")
	assert.Less(t, result.AvgResponseTime, 100*time.Millisecond, "平均响应时间应该小于100ms")
	assert.Greater(t, result.RequestsPerSecond, 100.0, "RPS应该大于100")

	pts.logger.Info("健康检查API性能测试结果",
		zap.Int64("total_requests", result.TotalRequests),
		zap.Int64("success_requests", result.SuccessRequests),
		zap.Int64("failed_requests", result.FailedRequests),
		zap.Duration("avg_response_time", result.AvgResponseTime),
		zap.Duration("min_response_time", result.MinResponseTime),
		zap.Duration("max_response_time", result.MaxResponseTime),
		zap.Float64("requests_per_second", result.RequestsPerSecond),
		zap.Float64("error_rate", result.ErrorRate),
	)
}

// TestProcessDefinitionsPerformance 测试流程定义API性能
func TestProcessDefinitionsPerformance(t *testing.T) {
	pts := NewPerformanceTestSuite()
	defer pts.Close()

	pts.logger.Info("开始流程定义API性能测试")

	config := LoadTestConfig{
		ConcurrentUsers: 20,
		TestDuration:    20 * time.Second,
		RampUpTime:      3 * time.Second,
		RequestTimeout:  10 * time.Second,
	}

	requestFunc := func() (int, time.Duration, error) {
		return pts.makeHTTPRequest("GET", "/api/v1/process-definitions", nil)
	}

	result := pts.runLoadTest(config, requestFunc)

	// 验证性能指标
	assert.Greater(t, result.TotalRequests, int64(0), "应该有请求被发送")
	assert.Less(t, result.ErrorRate, 10.0, "错误率应该低于10%")
	assert.Less(t, result.AvgResponseTime, 200*time.Millisecond, "平均响应时间应该小于200ms")

	pts.logger.Info("流程定义API性能测试结果",
		zap.Int64("total_requests", result.TotalRequests),
		zap.Float64("requests_per_second", result.RequestsPerSecond),
		zap.Duration("avg_response_time", result.AvgResponseTime),
		zap.Float64("error_rate", result.ErrorRate),
	)
}

// TestCreateProcessDefinitionPerformance 测试创建流程定义API性能
func TestCreateProcessDefinitionPerformance(t *testing.T) {
	pts := NewPerformanceTestSuite()
	defer pts.Close()

	pts.logger.Info("开始创建流程定义API性能测试")

	config := LoadTestConfig{
		ConcurrentUsers: 10,
		TestDuration:    15 * time.Second,
		RampUpTime:      2 * time.Second,
		RequestTimeout:  15 * time.Second,
	}

	requestID := 0
	requestFunc := func() (int, time.Duration, error) {
		requestID++
		createReq := map[string]interface{}{
			"key":         fmt.Sprintf("test-process-%d", requestID),
			"name":        fmt.Sprintf("测试流程-%d", requestID),
			"description": "性能测试流程定义",
			"resource":    `{"version": "1.0", "steps": []}`,
			"category":    "性能测试",
		}
		return pts.makeHTTPRequest("POST", "/api/v1/process-definitions", createReq)
	}

	result := pts.runLoadTest(config, requestFunc)

	// 验证性能指标
	assert.Greater(t, result.TotalRequests, int64(0), "应该有请求被发送")
	assert.Less(t, result.ErrorRate, 15.0, "错误率应该低于15%")
	assert.Less(t, result.AvgResponseTime, 500*time.Millisecond, "平均响应时间应该小于500ms")

	pts.logger.Info("创建流程定义API性能测试结果",
		zap.Int64("total_requests", result.TotalRequests),
		zap.Float64("requests_per_second", result.RequestsPerSecond),
		zap.Duration("avg_response_time", result.AvgResponseTime),
		zap.Float64("error_rate", result.ErrorRate),
	)
}

// TestMixedWorkloadPerformance 测试混合工作负载性能
func TestMixedWorkloadPerformance(t *testing.T) {
	pts := NewPerformanceTestSuite()
	defer pts.Close()

	pts.logger.Info("开始混合工作负载性能测试")

	config := LoadTestConfig{
		ConcurrentUsers: 30,
		TestDuration:    25 * time.Second,
		RampUpTime:      5 * time.Second,
		RequestTimeout:  10 * time.Second,
	}

	// 混合请求函数
	requestCount := 0
	requestFunc := func() (int, time.Duration, error) {
		requestCount++
		switch requestCount % 4 {
		case 0:
			// 健康检查 (25%)
			return pts.makeHTTPRequest("GET", "/health", nil)
		case 1:
			// 查询流程定义 (25%)
			return pts.makeHTTPRequest("GET", "/api/v1/process-definitions", nil)
		case 2:
			// 查询任务列表 (25%)
			return pts.makeHTTPRequest("GET", "/api/v1/tasks", nil)
		case 3:
			// 查询流程实例 (25%)
			return pts.makeHTTPRequest("GET", "/api/v1/process-instances", nil)
		default:
			return pts.makeHTTPRequest("GET", "/health", nil)
		}
	}

	result := pts.runLoadTest(config, requestFunc)

	// 验证性能指标
	assert.Greater(t, result.TotalRequests, int64(0), "应该有请求被发送")
	assert.Less(t, result.ErrorRate, 10.0, "错误率应该低于10%")
	assert.Less(t, result.AvgResponseTime, 300*time.Millisecond, "平均响应时间应该小于300ms")
	assert.Greater(t, result.RequestsPerSecond, 50.0, "RPS应该大于50")

	pts.logger.Info("混合工作负载性能测试结果",
		zap.Int64("total_requests", result.TotalRequests),
		zap.Float64("requests_per_second", result.RequestsPerSecond),
		zap.Duration("avg_response_time", result.AvgResponseTime),
		zap.Duration("min_response_time", result.MinResponseTime),
		zap.Duration("max_response_time", result.MaxResponseTime),
		zap.Float64("error_rate", result.ErrorRate),
	)
}

// TestStressTest 压力测试 - 测试系统在极限负载下的表现
func TestStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试 (使用 -short 标志)")
	}

	pts := NewPerformanceTestSuite()
	defer pts.Close()

	pts.logger.Info("开始压力测试")

	config := LoadTestConfig{
		ConcurrentUsers: 100,              // 高并发
		TestDuration:    60 * time.Second, // 较长时间
		RampUpTime:      10 * time.Second,
		RequestTimeout:  15 * time.Second,
	}

	requestFunc := func() (int, time.Duration, error) {
		return pts.makeHTTPRequest("GET", "/health", nil)
	}

	result := pts.runLoadTest(config, requestFunc)

	// 压力测试的验证条件更宽松
	assert.Greater(t, result.TotalRequests, int64(0), "应该有请求被发送")
	assert.Less(t, result.ErrorRate, 20.0, "压力测试错误率应该低于20%")
	assert.Less(t, result.AvgResponseTime, 1*time.Second, "平均响应时间应该小于1s")

	pts.logger.Info("压力测试结果",
		zap.Int64("total_requests", result.TotalRequests),
		zap.Float64("requests_per_second", result.RequestsPerSecond),
		zap.Duration("avg_response_time", result.AvgResponseTime),
		zap.Duration("max_response_time", result.MaxResponseTime),
		zap.Float64("error_rate", result.ErrorRate),
	)

	// 如果压力测试表现良好，记录一下
	if result.ErrorRate < 5.0 && result.AvgResponseTime < 200*time.Millisecond {
		pts.logger.Info("系统在压力测试中表现优秀！",
			zap.Float64("error_rate", result.ErrorRate),
			zap.Duration("avg_response_time", result.AvgResponseTime),
		)
	}
}

// BenchmarkHealthCheck 基准测试 - 健康检查API
func BenchmarkHealthCheck(b *testing.B) {
	pts := NewPerformanceTestSuite()
	defer pts.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			statusCode, _, err := pts.makeHTTPRequest("GET", "/health", nil)
			if err != nil || statusCode != 200 {
				b.Errorf("健康检查失败: status=%d, err=%v", statusCode, err)
			}
		}
	})
}

// BenchmarkProcessDefinitionsList 基准测试 - 流程定义列表API
func BenchmarkProcessDefinitionsList(b *testing.B) {
	pts := NewPerformanceTestSuite()
	defer pts.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			statusCode, _, err := pts.makeHTTPRequest("GET", "/api/v1/process-definitions", nil)
			if err != nil || statusCode != 200 {
				b.Errorf("查询流程定义失败: status=%d, err=%v", statusCode, err)
			}
		}
	})
}
