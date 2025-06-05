// Package biz 提供业务逻辑层功能的测试
// 包含流程定义管理业务逻辑的单元测试
package biz

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/workflow-engine/workflow-engine/internal/data/ent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// MockProcessDefinitionRepo 模拟流程定义仓储
type MockProcessDefinitionRepo struct {
	mock.Mock
}

func (m *MockProcessDefinitionRepo) Create(ctx context.Context, pd *ent.ProcessDefinition) (*ent.ProcessDefinition, error) {
	args := m.Called(ctx, pd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.ProcessDefinition), args.Error(1)
}

func (m *MockProcessDefinitionRepo) GetByID(ctx context.Context, id string) (*ent.ProcessDefinition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.ProcessDefinition), args.Error(1)
}

func (m *MockProcessDefinitionRepo) GetLatestByKey(ctx context.Context, key string) (*ent.ProcessDefinition, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.ProcessDefinition), args.Error(1)
}

func (m *MockProcessDefinitionRepo) GetByKeyAndVersion(ctx context.Context, key string, version int) (*ent.ProcessDefinition, error) {
	args := m.Called(ctx, key, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.ProcessDefinition), args.Error(1)
}

func (m *MockProcessDefinitionRepo) Update(ctx context.Context, pd *ent.ProcessDefinition) (*ent.ProcessDefinition, error) {
	args := m.Called(ctx, pd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.ProcessDefinition), args.Error(1)
}

func (m *MockProcessDefinitionRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProcessDefinitionRepo) List(ctx context.Context, filter *ProcessDefinitionFilter, opts *QueryOptions) ([]*ent.ProcessDefinition, *PaginationResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*ent.ProcessDefinition), args.Get(1).(*PaginationResult), args.Error(2)
}

func (m *MockProcessDefinitionRepo) Count(ctx context.Context, filter *ProcessDefinitionFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

func (m *MockProcessDefinitionRepo) Deploy(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProcessDefinitionRepo) Suspend(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCacheRepo 模拟缓存仓储
type MockCacheRepo struct {
	mock.Mock
}

func (m *MockCacheRepo) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheRepo) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheRepo) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheRepo) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheRepo) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockCacheRepo) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *MockCacheRepo) HSet(ctx context.Context, key, field string, value interface{}) error {
	args := m.Called(ctx, key, field, value)
	return args.Error(0)
}

func (m *MockCacheRepo) HDel(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

func (m *MockCacheRepo) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

// 创建测试用的流程定义
func createTestProcessDefinition() *ent.ProcessDefinition {
	now := time.Now()
	return &ent.ProcessDefinition{
		ID:          1,
		Key:         "test-process",
		Name:        "测试流程",
		Description: "这是一个测试流程",
		Category:    "test",
		Version:     1,
		Resource:    `{"id":"test-process","name":"测试流程","elements":[{"id":"start","type":"startEvent"}]}`,
		Suspended:   false,
		TenantID:    "default",
		DeployTime:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// 创建测试logger
func createTestLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	return logger, recorded
}

// TestProcessDefinitionUseCase_CreateProcessDefinition 测试创建流程定义功能
func TestProcessDefinitionUseCase_CreateProcessDefinition(t *testing.T) {
	logger, _ := createTestLogger()

	t.Run("创建新流程定义成功", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		req := &CreateProcessDefinitionRequest{
			Key:         "test-process",
			Name:        "测试流程",
			Description: "这是一个测试流程",
			Category:    "test",
			Resource:    `{"id":"test-process","name":"测试流程","elements":[{"id":"start","type":"startEvent"}]}`,
			TenantID:    "default",
		}

		expectedPD := createTestProcessDefinition()

		// 设置mock期望
		mockRepo.On("GetLatestByKey", mock.Anything, "test-process").Return(nil, errors.New("not found"))
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(pd *ent.ProcessDefinition) bool {
			return pd.Key == "test-process" && pd.Version == 1
		})).Return(expectedPD, nil)
		mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), time.Hour).Return(nil)

		// 执行测试
		result, err := uc.CreateProcessDefinition(context.Background(), req)

		// 验证结果
		assert.NoError(t, err, "创建流程定义不应该返回错误")
		assert.NotNil(t, result, "创建的流程定义不应该为空")
		assert.Equal(t, "test-process", result.Key, "流程定义键应该匹配")
		assert.Equal(t, "测试流程", result.Name, "流程定义名称应该匹配")
		assert.Equal(t, int32(1), result.Version, "流程定义版本应该为1")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("创建新版本流程定义成功", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		req := &CreateProcessDefinitionRequest{
			Key:         "test-process",
			Name:        "测试流程",
			Description: "这是一个测试流程",
			Category:    "test",
			Resource:    `{"id":"test-process","name":"测试流程","elements":[{"id":"start","type":"startEvent"}]}`,
			TenantID:    "default",
		}

		existingPD := createTestProcessDefinition()
		newPD := createTestProcessDefinition()
		newPD.ID = 2
		newPD.Version = 2

		// 设置mock期望
		mockRepo.On("GetLatestByKey", mock.Anything, "test-process").Return(existingPD, nil)
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(pd *ent.ProcessDefinition) bool {
			return pd.Key == "test-process" && pd.Version == 2
		})).Return(newPD, nil)
		mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), time.Hour).Return(nil)

		// 执行测试
		result, err := uc.CreateProcessDefinition(context.Background(), req)

		// 验证结果
		assert.NoError(t, err, "创建流程定义不应该返回错误")
		assert.NotNil(t, result, "创建的流程定义不应该为空")
		assert.Equal(t, int32(2), result.Version, "流程定义版本应该为2")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("参数验证失败", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		req := &CreateProcessDefinitionRequest{
			// 缺少必要字段
			Name: "测试流程",
		}

		// 执行测试
		result, err := uc.CreateProcessDefinition(context.Background(), req)

		// 验证结果
		assert.Error(t, err, "应该返回参数验证错误")
		assert.Nil(t, result, "结果应该为空")
		assert.Contains(t, err.Error(), "流程键不能为空", "错误信息应该包含验证失败原因")
	})

	t.Run("流程定义内容验证失败", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		req := &CreateProcessDefinitionRequest{
			Key:         "test-process",
			Name:        "测试流程",
			Description: "这是一个测试流程",
			Category:    "test",
			Resource:    `{"name":"测试流程"}`, // 缺少必要字段
			TenantID:    "default",
		}

		// 设置mock期望
		mockRepo.On("GetLatestByKey", mock.Anything, "test-process").Return(nil, errors.New("not found"))

		// 执行测试
		result, err := uc.CreateProcessDefinition(context.Background(), req)

		// 验证结果
		assert.Error(t, err, "应该返回流程定义验证错误")
		assert.Nil(t, result, "结果应该为空")
		assert.Contains(t, err.Error(), "流程定义缺少", "错误信息应该包含验证失败原因")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
	})
}

// TestProcessDefinitionUseCase_GetProcessDefinition 测试获取流程定义功能
func TestProcessDefinitionUseCase_GetProcessDefinition(t *testing.T) {
	logger, _ := createTestLogger()

	t.Run("获取流程定义成功", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		expectedPD := createTestProcessDefinition()
		id := strconv.FormatInt(expectedPD.ID, 10)

		// 设置mock期望
		mockCache.On("Get", mock.Anything, "process_definition:1").Return("", errors.New("cache miss"))
		mockRepo.On("GetByID", mock.Anything, id).Return(expectedPD, nil)
		mockCache.On("Set", mock.Anything, "process_definition:1", mock.AnythingOfType("string"), time.Hour).Return(nil)

		// 执行测试
		result, err := uc.GetProcessDefinition(context.Background(), id)

		// 验证结果
		assert.NoError(t, err, "获取流程定义不应该返回错误")
		assert.NotNil(t, result, "获取的流程定义不应该为空")
		assert.Equal(t, "test-process", result.Key, "流程定义键应该匹配")
		assert.Equal(t, "测试流程", result.Name, "流程定义名称应该匹配")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("流程定义不存在", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		id := "999"

		// 设置mock期望
		mockCache.On("Get", mock.Anything, "process_definition:999").Return("", errors.New("cache miss"))
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

		// 执行测试
		result, err := uc.GetProcessDefinition(context.Background(), id)

		// 验证结果
		assert.Error(t, err, "应该返回未找到错误")
		assert.Nil(t, result, "结果应该为空")
		assert.Contains(t, err.Error(), "获取流程定义失败", "错误信息应该包含失败原因")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

// TestProcessDefinitionUseCase_UpdateProcessDefinition 测试更新流程定义功能
func TestProcessDefinitionUseCase_UpdateProcessDefinition(t *testing.T) {
	logger, _ := createTestLogger()

	t.Run("更新流程定义成功", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		existingPD := createTestProcessDefinition()
		id := strconv.FormatInt(existingPD.ID, 10)

		req := &UpdateProcessDefinitionRequest{
			Name:        "更新后的流程",
			Description: "更新后的描述",
		}

		updatedPD := createTestProcessDefinition()
		updatedPD.Name = "更新后的流程"
		updatedPD.Description = "更新后的描述"

		// 设置mock期望
		mockRepo.On("GetByID", mock.Anything, id).Return(existingPD, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(pd *ent.ProcessDefinition) bool {
			return pd.Name == "更新后的流程" && pd.Description == "更新后的描述"
		})).Return(updatedPD, nil)
		mockCache.On("Delete", mock.Anything, "process_definition:1").Return(nil)

		// 执行测试
		result, err := uc.UpdateProcessDefinition(context.Background(), id, req)

		// 验证结果
		assert.NoError(t, err, "更新流程定义不应该返回错误")
		assert.NotNil(t, result, "更新的流程定义不应该为空")
		assert.Equal(t, "更新后的流程", result.Name, "流程定义名称应该已更新")
		assert.Equal(t, "更新后的描述", result.Description, "流程定义描述应该已更新")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("流程定义不存在", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		id := "999"
		req := &UpdateProcessDefinitionRequest{
			Name: "更新后的流程",
		}

		// 设置mock期望
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

		// 执行测试
		result, err := uc.UpdateProcessDefinition(context.Background(), id, req)

		// 验证结果
		assert.Error(t, err, "应该返回未找到错误")
		assert.Nil(t, result, "结果应该为空")
		assert.Contains(t, err.Error(), "获取待更新的流程定义失败", "错误信息应该包含失败原因")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
	})
}

// TestProcessDefinitionUseCase_DeleteProcessDefinition 测试删除流程定义功能
func TestProcessDefinitionUseCase_DeleteProcessDefinition(t *testing.T) {
	logger, _ := createTestLogger()

	t.Run("删除流程定义成功", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		id := "1"

		// 设置mock期望
		mockRepo.On("Delete", mock.Anything, id).Return(nil)
		mockCache.On("Delete", mock.Anything, "process_definition:1").Return(nil)

		// 执行测试
		err := uc.DeleteProcessDefinition(context.Background(), id)

		// 验证结果
		assert.NoError(t, err, "删除流程定义不应该返回错误")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("删除流程定义失败", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		id := "1"

		// 设置mock期望
		mockRepo.On("Delete", mock.Anything, id).Return(errors.New("delete failed"))

		// 执行测试
		err := uc.DeleteProcessDefinition(context.Background(), id)

		// 验证结果
		assert.Error(t, err, "应该返回删除失败错误")
		assert.Contains(t, err.Error(), "删除流程定义失败", "错误信息应该包含失败原因")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
	})
}

// TestProcessDefinitionUseCase_DeployProcessDefinition 测试部署流程定义功能
func TestProcessDefinitionUseCase_DeployProcessDefinition(t *testing.T) {
	logger, _ := createTestLogger()

	t.Run("部署流程定义成功", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		id := "1"

		// 设置mock期望
		mockRepo.On("Deploy", mock.Anything, id).Return(nil)
		mockCache.On("Delete", mock.Anything, "process_definition:1").Return(nil)

		// 执行测试
		err := uc.DeployProcessDefinition(context.Background(), id)

		// 验证结果
		assert.NoError(t, err, "部署流程定义不应该返回错误")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

// TestProcessDefinitionUseCase_SuspendProcessDefinition 测试挂起流程定义功能
func TestProcessDefinitionUseCase_SuspendProcessDefinition(t *testing.T) {
	logger, _ := createTestLogger()

	t.Run("挂起流程定义成功", func(t *testing.T) {
		// 准备测试数据
		mockRepo := new(MockProcessDefinitionRepo)
		mockCache := new(MockCacheRepo)
		uc := NewProcessDefinitionUseCase(mockRepo, mockCache, logger)

		id := "1"

		// 设置mock期望
		mockRepo.On("Suspend", mock.Anything, id).Return(nil)
		mockCache.On("Delete", mock.Anything, "process_definition:1").Return(nil)

		// 执行测试
		err := uc.SuspendProcessDefinition(context.Background(), id)

		// 验证结果
		assert.NoError(t, err, "挂起流程定义不应该返回错误")

		// 验证mock调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}
