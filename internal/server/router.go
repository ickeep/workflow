// Package server HTTP路由器
// 提供RESTful API路由配置和基础处理器
package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Router HTTP路由器
type Router struct {
	router *mux.Router
	logger *zap.Logger
}

// NewRouter 创建新的HTTP路由器
func NewRouter(logger *zap.Logger) *Router {
	r := &Router{
		router: mux.NewRouter(),
		logger: logger,
	}

	r.setupRoutes()
	return r
}

// ServeHTTP 实现http.Handler接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Code      int         `json:"code"`            // 响应码
	Message   string      `json:"message"`         // 响应消息
	Data      interface{} `json:"data,omitempty"`  // 响应数据
	Error     string      `json:"error,omitempty"` // 错误信息
	Timestamp string      `json:"timestamp"`       // 时间戳
}

// setupRoutes 设置路由
func (r *Router) setupRoutes() {
	// 中间件
	r.router.Use(r.loggingMiddleware)
	r.router.Use(r.corsMiddleware)

	// API版本路由
	api := r.router.PathPrefix("/api/v1").Subrouter()

	// 流程定义路由
	processDefinitions := api.PathPrefix("/process-definitions").Subrouter()
	processDefinitions.HandleFunc("", r.handleListProcessDefinitions).Methods("GET")
	processDefinitions.HandleFunc("", r.handleCreateProcessDefinition).Methods("POST")
	processDefinitions.HandleFunc("/{id}", r.handleGetProcessDefinition).Methods("GET")
	processDefinitions.HandleFunc("/{id}", r.handleUpdateProcessDefinition).Methods("PUT")
	processDefinitions.HandleFunc("/{id}", r.handleDeleteProcessDefinition).Methods("DELETE")
	processDefinitions.HandleFunc("/{id}/deploy", r.handleDeployProcessDefinition).Methods("POST")

	// 流程实例路由
	processInstances := api.PathPrefix("/process-instances").Subrouter()
	processInstances.HandleFunc("", r.handleListProcessInstances).Methods("GET")
	processInstances.HandleFunc("", r.handleStartProcessInstance).Methods("POST")
	processInstances.HandleFunc("/{id}", r.handleGetProcessInstance).Methods("GET")
	processInstances.HandleFunc("/{id}/suspend", r.handleSuspendProcessInstance).Methods("POST")
	processInstances.HandleFunc("/{id}/activate", r.handleActivateProcessInstance).Methods("POST")
	processInstances.HandleFunc("/{id}/terminate", r.handleTerminateProcessInstance).Methods("POST")

	// 任务路由
	tasks := api.PathPrefix("/tasks").Subrouter()
	tasks.HandleFunc("", r.handleListTasks).Methods("GET")
	tasks.HandleFunc("/{id}", r.handleGetTask).Methods("GET")
	tasks.HandleFunc("/{id}/claim", r.handleClaimTask).Methods("POST")
	tasks.HandleFunc("/{id}/complete", r.handleCompleteTask).Methods("POST")
	tasks.HandleFunc("/{id}/delegate", r.handleDelegateTask).Methods("POST")

	// 历史数据路由
	history := api.PathPrefix("/history").Subrouter()
	history.HandleFunc("/process-instances", r.handleListHistoricProcessInstances).Methods("GET")
	history.HandleFunc("/process-instances/{id}", r.handleGetHistoricProcessInstance).Methods("GET")

	// 健康检查路由
	r.router.HandleFunc("/health", r.handleHealthCheck).Methods("GET")
	r.router.HandleFunc("/ready", r.handleReadinessCheck).Methods("GET")
}

// ====================
// 中间件
// ====================

// loggingMiddleware 日志中间件
func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		r.logger.Info("HTTP请求开始",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
			zap.String("remote_addr", req.RemoteAddr))

		next.ServeHTTP(w, req)

		r.logger.Info("HTTP请求完成",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
			zap.Duration("duration", time.Since(start)))
	})
}

// corsMiddleware CORS中间件
func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, req)
	})
}

// ====================
// 响应辅助函数
// ====================

// writeJSONResponse 写入JSON响应
func (r *Router) writeJSONResponse(w http.ResponseWriter, statusCode int, response *APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response.Timestamp = time.Now().Format(time.RFC3339)
	json.NewEncoder(w).Encode(response)
}

// successResponse 创建成功响应
func (r *Router) successResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Code:    200,
		Message: "成功",
		Data:    data,
	}
}

// errorResponse 创建错误响应
func (r *Router) errorResponse(code int, message string) *APIResponse {
	return &APIResponse{
		Code:    code,
		Message: message,
	}
}

// ====================
// 基础处理器实现
// ====================

// handleListProcessDefinitions 查询流程定义列表
func (r *Router) handleListProcessDefinitions(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("处理查询流程定义列表请求")

	// 模拟数据
	data := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"id":      "1",
				"key":     "sample-process",
				"name":    "示例流程",
				"version": 1,
			},
		},
		"total":     1,
		"page":      1,
		"page_size": 20,
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleCreateProcessDefinition 创建流程定义
func (r *Router) handleCreateProcessDefinition(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("处理创建流程定义请求")

	// 模拟创建成功
	data := map[string]interface{}{
		"id":      "2",
		"key":     "new-process",
		"name":    "新流程",
		"version": 1,
		"status":  "created",
	}

	r.writeJSONResponse(w, http.StatusCreated, r.successResponse(data))
}

// handleGetProcessDefinition 获取流程定义
func (r *Router) handleGetProcessDefinition(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理获取流程定义请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"key":     "sample-process",
		"name":    "示例流程",
		"version": 1,
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleUpdateProcessDefinition 更新流程定义
func (r *Router) handleUpdateProcessDefinition(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理更新流程定义请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"status":  "updated",
		"message": "流程定义更新成功",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleDeleteProcessDefinition 删除流程定义
func (r *Router) handleDeleteProcessDefinition(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理删除流程定义请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"message": "流程定义删除成功",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleDeployProcessDefinition 部署流程定义
func (r *Router) handleDeployProcessDefinition(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理部署流程定义请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"status":  "deployed",
		"message": "流程定义部署成功",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleStartProcessInstance 启动流程实例
func (r *Router) handleStartProcessInstance(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("处理启动流程实例请求")

	data := map[string]interface{}{
		"id":                    "inst-1",
		"process_definition_id": "1",
		"business_key":          "business-001",
		"status":                "running",
	}

	r.writeJSONResponse(w, http.StatusCreated, r.successResponse(data))
}

// handleGetProcessInstance 获取流程实例
func (r *Router) handleGetProcessInstance(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理获取流程实例请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":                    id,
		"process_definition_id": "1",
		"status":                "running",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleListProcessInstances 查询流程实例列表
func (r *Router) handleListProcessInstances(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("处理查询流程实例列表请求")

	data := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"id":                    "inst-1",
				"process_definition_id": "1",
				"status":                "running",
			},
		},
		"total":     1,
		"page":      1,
		"page_size": 20,
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleSuspendProcessInstance 挂起流程实例
func (r *Router) handleSuspendProcessInstance(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理挂起流程实例请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"status":  "suspended",
		"message": "流程实例挂起成功",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleActivateProcessInstance 激活流程实例
func (r *Router) handleActivateProcessInstance(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理激活流程实例请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"status":  "active",
		"message": "流程实例激活成功",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleTerminateProcessInstance 终止流程实例
func (r *Router) handleTerminateProcessInstance(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理终止流程实例请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"status":  "terminated",
		"message": "流程实例终止成功",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleListTasks 查询任务列表
func (r *Router) handleListTasks(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("处理查询任务列表请求")

	data := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"id":          "task-1",
				"name":        "审批任务",
				"assignee":    "user1",
				"status":      "pending",
				"create_time": time.Now().Format(time.RFC3339),
			},
		},
		"total":     1,
		"page":      1,
		"page_size": 20,
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleGetTask 获取任务
func (r *Router) handleGetTask(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理获取任务请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":          id,
		"name":        "审批任务",
		"assignee":    "user1",
		"status":      "pending",
		"create_time": time.Now().Format(time.RFC3339),
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleClaimTask 认领任务
func (r *Router) handleClaimTask(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理认领任务请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"status":  "claimed",
		"message": "任务认领成功",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleCompleteTask 完成任务
func (r *Router) handleCompleteTask(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理完成任务请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"status":  "completed",
		"message": "任务完成成功",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleDelegateTask 委派任务
func (r *Router) handleDelegateTask(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理委派任务请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"status":  "delegated",
		"message": "任务委派成功",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleListHistoricProcessInstances 查询历史流程实例列表
func (r *Router) handleListHistoricProcessInstances(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("处理查询历史流程实例列表请求")

	data := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"id":         "hist-1",
				"start_time": time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
				"end_time":   time.Now().Format(time.RFC3339),
				"status":     "completed",
			},
		},
		"total":     1,
		"page":      1,
		"page_size": 20,
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleGetHistoricProcessInstance 获取历史流程实例
func (r *Router) handleGetHistoricProcessInstance(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	r.logger.Info("处理获取历史流程实例请求", zap.String("id", id))

	data := map[string]interface{}{
		"id":         id,
		"start_time": time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
		"end_time":   time.Now().Format(time.RFC3339),
		"status":     "completed",
		"duration":   "1h30m",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleHealthCheck 健康检查
func (r *Router) handleHealthCheck(w http.ResponseWriter, req *http.Request) {
	data := map[string]interface{}{
		"status":  "healthy",
		"service": "workflow-engine",
		"version": "1.0.0",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}

// handleReadinessCheck 就绪检查
func (r *Router) handleReadinessCheck(w http.ResponseWriter, req *http.Request) {
	data := map[string]interface{}{
		"status":  "ready",
		"service": "workflow-engine",
		"version": "1.0.0",
	}

	r.writeJSONResponse(w, http.StatusOK, r.successResponse(data))
}
