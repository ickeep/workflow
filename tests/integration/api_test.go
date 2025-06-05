// Package integration HTTP API集成测试
// 提供完整的HTTP API集成测试用例，测试服务器端到端功能
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/workflow-engine/workflow-engine/internal/server"
)

// APITestSuite HTTP API集成测试套件
type APITestSuite struct {
	suite.Suite
	server *httptest.Server
	router *server.Router
	logger *zap.Logger
}

// SetupSuite 设置测试套件，在所有测试前执行一次
func (suite *APITestSuite) SetupSuite() {
	// 初始化测试日志器
	logger, err := zap.NewDevelopment()
	suite.Require().NoError(err, "初始化日志器失败")
	suite.logger = logger

	// 创建HTTP路由器
	suite.router = server.NewRouter(logger)

	// 创建测试服务器
	suite.server = httptest.NewServer(suite.router)

	suite.logger.Info("API集成测试服务器启动", zap.String("url", suite.server.URL))
}

// TearDownSuite 清理测试套件，在所有测试后执行一次
func (suite *APITestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
		suite.logger.Info("API集成测试服务器关闭")
	}
	if suite.logger != nil {
		suite.logger.Sync()
	}
}

// makeRequest 发送HTTP请求的辅助函数
func (suite *APITestSuite) makeRequest(method, path string, body interface{}) (*http.Response, []byte) {
	var reqBody bytes.Buffer
	if body != nil {
		err := json.NewEncoder(&reqBody).Encode(body)
		suite.Require().NoError(err, "编码请求体失败")
	}

	req, err := http.NewRequest(method, suite.server.URL+path, &reqBody)
	suite.Require().NoError(err, "创建请求失败")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	suite.Require().NoError(err, "发送请求失败")

	respBody := make([]byte, 0)
	if resp.Body != nil {
		defer resp.Body.Close()
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(resp.Body)
		suite.Require().NoError(err, "读取响应体失败")
		respBody = buf.Bytes()
	}

	return resp, respBody
}

// parseResponse 解析API响应的辅助函数
func (suite *APITestSuite) parseResponse(respBody []byte) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal(respBody, &result)
	suite.Require().NoError(err, "解析响应JSON失败")
	return result
}

// ====================
// 健康检查API测试
// ====================

// TestHealthCheck 测试健康检查API
func (suite *APITestSuite) TestHealthCheck() {
	suite.logger.Info("开始测试健康检查API")

	resp, body := suite.makeRequest("GET", "/health", nil)

	// 验证HTTP状态码
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "健康检查应返回200状态码")

	// 验证响应格式
	result := suite.parseResponse(body)
	assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")
	assert.Equal(suite.T(), "成功", result["message"], "响应消息应为'成功'")
	assert.NotNil(suite.T(), result["data"], "响应数据不应为空")
	assert.NotEmpty(suite.T(), result["timestamp"], "时间戳不应为空")

	// 验证健康状态数据
	data, ok := result["data"].(map[string]interface{})
	assert.True(suite.T(), ok, "data字段应为对象")
	assert.Equal(suite.T(), "healthy", data["status"], "健康状态应为healthy")
	assert.Equal(suite.T(), "workflow-engine", data["service"], "服务名应为workflow-engine")
	assert.Equal(suite.T(), "1.0.0", data["version"], "版本号应为1.0.0")

	suite.logger.Info("健康检查API测试通过")
}

// TestReadinessCheck 测试就绪检查API
func (suite *APITestSuite) TestReadinessCheck() {
	suite.logger.Info("开始测试就绪检查API")

	resp, body := suite.makeRequest("GET", "/ready", nil)

	// 验证HTTP状态码
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "就绪检查应返回200状态码")

	// 验证响应格式
	result := suite.parseResponse(body)
	assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

	// 验证就绪状态数据
	data, ok := result["data"].(map[string]interface{})
	assert.True(suite.T(), ok, "data字段应为对象")
	assert.Equal(suite.T(), "ready", data["status"], "就绪状态应为ready")

	suite.logger.Info("就绪检查API测试通过")
}

// ====================
// 流程定义API测试
// ====================

// TestProcessDefinitionsAPI 测试流程定义相关API
func (suite *APITestSuite) TestProcessDefinitionsAPI() {
	suite.logger.Info("开始测试流程定义API")

	// 测试查询流程定义列表
	suite.Run("查询流程定义列表", func() {
		resp, body := suite.makeRequest("GET", "/api/v1/process-definitions", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "查询列表应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		// 验证分页数据结构
		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.NotNil(suite.T(), data["items"], "items字段不应为空")
		assert.NotNil(suite.T(), data["total"], "total字段不应为空")
		assert.NotNil(suite.T(), data["page"], "page字段不应为空")
		assert.NotNil(suite.T(), data["page_size"], "page_size字段不应为空")
	})

	// 测试创建流程定义
	suite.Run("创建流程定义", func() {
		createReq := map[string]interface{}{
			"key":         "test-process",
			"name":        "测试流程",
			"description": "这是一个测试流程定义",
			"resource":    `{"version": "1.0", "steps": []}`,
			"category":    "测试分类",
		}

		resp, body := suite.makeRequest("POST", "/api/v1/process-definitions", createReq)

		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode, "创建流程定义应返回201状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		// 验证创建结果
		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.NotEmpty(suite.T(), data["id"], "创建的流程定义应有ID")
		assert.Equal(suite.T(), "created", data["status"], "状态应为created")
	})

	// 测试获取流程定义详情
	suite.Run("获取流程定义详情", func() {
		resp, body := suite.makeRequest("GET", "/api/v1/process-definitions/1", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "获取详情应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		// 验证详情数据
		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "1", data["id"], "ID应匹配")
		assert.NotEmpty(suite.T(), data["name"], "名称不应为空")
	})

	// 测试更新流程定义
	suite.Run("更新流程定义", func() {
		updateReq := map[string]interface{}{
			"name":        "更新后的流程名称",
			"description": "更新后的描述",
		}

		resp, body := suite.makeRequest("PUT", "/api/v1/process-definitions/1", updateReq)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "更新流程定义应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "updated", data["status"], "状态应为updated")
	})

	// 测试部署流程定义
	suite.Run("部署流程定义", func() {
		resp, body := suite.makeRequest("POST", "/api/v1/process-definitions/1/deploy", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "部署流程定义应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "deployed", data["status"], "状态应为deployed")
	})

	// 测试删除流程定义 (放在最后)
	suite.Run("删除流程定义", func() {
		resp, body := suite.makeRequest("DELETE", "/api/v1/process-definitions/999", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "删除流程定义应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")
	})

	suite.logger.Info("流程定义API测试完成")
}

// ====================
// 流程实例API测试
// ====================

// TestProcessInstancesAPI 测试流程实例相关API
func (suite *APITestSuite) TestProcessInstancesAPI() {
	suite.logger.Info("开始测试流程实例API")

	// 测试启动流程实例
	suite.Run("启动流程实例", func() {
		startReq := map[string]interface{}{
			"process_definition_id": "1",
			"business_key":          "test-business-001",
			"variables": map[string]interface{}{
				"applicant": "张三",
				"amount":    10000,
			},
			"start_user_id": "user-1",
		}

		resp, body := suite.makeRequest("POST", "/api/v1/process-instances", startReq)

		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode, "启动流程实例应返回201状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		// 验证启动结果
		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.NotEmpty(suite.T(), data["id"], "流程实例应有ID")
		assert.Equal(suite.T(), "running", data["status"], "状态应为running")
	})

	// 测试查询流程实例列表
	suite.Run("查询流程实例列表", func() {
		resp, body := suite.makeRequest("GET", "/api/v1/process-instances", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "查询列表应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		// 验证分页数据结构
		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.NotNil(suite.T(), data["items"], "items字段不应为空")
	})

	// 测试获取流程实例详情
	suite.Run("获取流程实例详情", func() {
		resp, body := suite.makeRequest("GET", "/api/v1/process-instances/inst-1", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "获取详情应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "inst-1", data["id"], "ID应匹配")
	})

	// 测试挂起流程实例
	suite.Run("挂起流程实例", func() {
		resp, body := suite.makeRequest("POST", "/api/v1/process-instances/inst-1/suspend", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "挂起流程实例应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "suspended", data["status"], "状态应为suspended")
	})

	// 测试激活流程实例
	suite.Run("激活流程实例", func() {
		resp, body := suite.makeRequest("POST", "/api/v1/process-instances/inst-1/activate", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "激活流程实例应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "active", data["status"], "状态应为active")
	})

	// 测试终止流程实例
	suite.Run("终止流程实例", func() {
		terminateReq := map[string]interface{}{
			"reason": "测试终止流程",
		}

		resp, body := suite.makeRequest("POST", "/api/v1/process-instances/inst-1/terminate", terminateReq)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "终止流程实例应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "terminated", data["status"], "状态应为terminated")
	})

	suite.logger.Info("流程实例API测试完成")
}

// ====================
// 任务API测试
// ====================

// TestTasksAPI 测试任务相关API
func (suite *APITestSuite) TestTasksAPI() {
	suite.logger.Info("开始测试任务API")

	// 测试查询任务列表
	suite.Run("查询任务列表", func() {
		resp, body := suite.makeRequest("GET", "/api/v1/tasks", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "查询任务列表应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.NotNil(suite.T(), data["items"], "items字段不应为空")
	})

	// 测试获取任务详情
	suite.Run("获取任务详情", func() {
		resp, body := suite.makeRequest("GET", "/api/v1/tasks/task-1", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "获取任务详情应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "task-1", data["id"], "任务ID应匹配")
		assert.NotEmpty(suite.T(), data["name"], "任务名称不应为空")
	})

	// 测试认领任务
	suite.Run("认领任务", func() {
		claimReq := map[string]interface{}{
			"user_id": "user-1",
		}

		resp, body := suite.makeRequest("POST", "/api/v1/tasks/task-1/claim", claimReq)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "认领任务应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "claimed", data["status"], "状态应为claimed")
	})

	// 测试委派任务
	suite.Run("委派任务", func() {
		delegateReq := map[string]interface{}{
			"delegate_id": "user-2",
			"comment":     "委派给其他人处理",
		}

		resp, body := suite.makeRequest("POST", "/api/v1/tasks/task-1/delegate", delegateReq)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "委派任务应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "delegated", data["status"], "状态应为delegated")
	})

	// 测试完成任务
	suite.Run("完成任务", func() {
		completeReq := map[string]interface{}{
			"variables": map[string]interface{}{
				"approved": true,
				"comment":  "审批通过",
			},
			"comment": "任务完成",
		}

		resp, body := suite.makeRequest("POST", "/api/v1/tasks/task-1/complete", completeReq)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "完成任务应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "completed", data["status"], "状态应为completed")
	})

	suite.logger.Info("任务API测试完成")
}

// ====================
// 历史数据API测试
// ====================

// TestHistoryAPI 测试历史数据相关API
func (suite *APITestSuite) TestHistoryAPI() {
	suite.logger.Info("开始测试历史数据API")

	// 测试查询历史流程实例列表
	suite.Run("查询历史流程实例列表", func() {
		resp, body := suite.makeRequest("GET", "/api/v1/history/process-instances", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "查询历史列表应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.NotNil(suite.T(), data["items"], "items字段不应为空")
	})

	// 测试获取历史流程实例详情
	suite.Run("获取历史流程实例详情", func() {
		resp, body := suite.makeRequest("GET", "/api/v1/history/process-instances/hist-1", nil)

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "获取历史详情应返回200状态码")

		result := suite.parseResponse(body)
		assert.Equal(suite.T(), float64(200), result["code"], "响应码应为200")

		data, ok := result["data"].(map[string]interface{})
		assert.True(suite.T(), ok, "data字段应为对象")
		assert.Equal(suite.T(), "hist-1", data["id"], "历史ID应匹配")
		assert.NotEmpty(suite.T(), data["duration"], "持续时间不应为空")
	})

	suite.logger.Info("历史数据API测试完成")
}

// ====================
// CORS和中间件测试
// ====================

// TestCORSMiddleware 测试CORS中间件
func (suite *APITestSuite) TestCORSMiddleware() {
	suite.logger.Info("开始测试CORS中间件")

	// 测试预检请求
	req, err := http.NewRequest("OPTIONS", suite.server.URL+"/api/v1/process-definitions", nil)
	suite.Require().NoError(err, "创建OPTIONS请求失败")

	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	suite.Require().NoError(err, "发送OPTIONS请求失败")
	defer resp.Body.Close()

	// 验证CORS头
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode, "OPTIONS请求应返回200状态码")
	assert.Equal(suite.T(), "*", resp.Header.Get("Access-Control-Allow-Origin"), "应允许所有来源")
	assert.Contains(suite.T(), resp.Header.Get("Access-Control-Allow-Methods"), "POST", "应允许POST方法")

	suite.logger.Info("CORS中间件测试通过")
}

// TestAPITestSuite 运行完整的API测试套件
func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
