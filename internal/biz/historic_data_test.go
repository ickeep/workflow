// Package biz 历史数据管理测试
package biz

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/workflow-engine/workflow-engine/internal/data/ent"
)

// MockHistoricProcessInstanceRepo 历史流程实例仓储模拟
type MockHistoricProcessInstanceRepo struct {
	mock.Mock
}

func (m *MockHistoricProcessInstanceRepo) Create(ctx context.Context, hpi *ent.HistoricProcessInstance) (*ent.HistoricProcessInstance, error) {
	args := m.Called(ctx, hpi)
	return args.Get(0).(*ent.HistoricProcessInstance), args.Error(1)
}

func (m *MockHistoricProcessInstanceRepo) GetHistoricProcessInstance(ctx context.Context, id int64) (*ent.HistoricProcessInstance, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*ent.HistoricProcessInstance), args.Error(1)
}

func (m *MockHistoricProcessInstanceRepo) ListHistoricProcessInstances(ctx context.Context, filter *HistoricProcessInstanceFilter) ([]*ent.HistoricProcessInstance, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*ent.HistoricProcessInstance), args.Int(1), args.Error(2)
}

func (m *MockHistoricProcessInstanceRepo) Count(ctx context.Context, filter *HistoricProcessInstanceFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

func (m *MockHistoricProcessInstanceRepo) ListByProcessDefinitionID(ctx context.Context, processDefinitionID string, opts *QueryOptions) ([]*ent.HistoricProcessInstance, *PaginationResult, error) {
	args := m.Called(ctx, processDefinitionID, opts)
	return args.Get(0).([]*ent.HistoricProcessInstance), args.Get(1).(*PaginationResult), args.Error(2)
}

func (m *MockHistoricProcessInstanceRepo) DeleteHistoricProcessInstance(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockHistoricProcessInstanceRepo) BatchDeleteHistoricProcessInstances(ctx context.Context, processDefinitionKey string, endTimeBefore time.Time) (int64, error) {
	args := m.Called(ctx, processDefinitionKey, endTimeBefore)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockHistoricProcessInstanceRepo) GetHistoricVariables(ctx context.Context, processInstanceID int64) ([]*HistoricVariableInstance, error) {
	args := m.Called(ctx, processInstanceID)
	return args.Get(0).([]*HistoricVariableInstance), args.Error(1)
}

func (m *MockHistoricProcessInstanceRepo) GetProcessStatistics(ctx context.Context, processDefinitionKey string, startTime, endTime time.Time) (*ProcessStatistics, error) {
	args := m.Called(ctx, processDefinitionKey, startTime, endTime)
	return args.Get(0).(*ProcessStatistics), args.Error(1)
}

func (m *MockHistoricProcessInstanceRepo) GetProcessTrend(ctx context.Context, processDefinitionKey string, startTime, endTime time.Time, granularity string) ([]*ProcessTrendData, error) {
	args := m.Called(ctx, processDefinitionKey, startTime, endTime, granularity)
	return args.Get(0).([]*ProcessTrendData), args.Error(1)
}

// TestHistoricDataUseCase_GetHistoricProcessInstance 测试获取历史流程实例
func TestHistoricDataUseCase_GetHistoricProcessInstance(t *testing.T) {
	mockRepo := new(MockHistoricProcessInstanceRepo)
	mockCache := new(MockCacheRepo)
	logger := zap.NewNop()

	useCase := NewHistoricDataUseCase(mockRepo, mockCache, logger)

	ctx := context.Background()
	instanceID := int64(1)

	// 准备测试数据
	now := time.Now()
	endTime := now.Add(time.Hour)
	historicInstance := &ent.HistoricProcessInstance{
		ID:                   1,
		ProcessDefinitionID:  1,
		ProcessDefinitionKey: "test-process",
		BusinessKey:          "business-123",
		StartTime:            now,
		EndTime:              &endTime,
		StartUserID:          "user123",
		DeleteReason:         "",
		TenantID:             "tenant1",
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	t.Run("成功获取历史流程实例", func(t *testing.T) {
		// 设置缓存未命中
		mockCache.On("Get", ctx, "historic_process_instance:1").Return("", assert.AnError).Once()

		// 设置数据库查询返回
		mockRepo.On("GetHistoricProcessInstance", ctx, instanceID).Return(historicInstance, nil).Once()

		// 设置变量查询返回
		mockRepo.On("GetHistoricVariables", ctx, instanceID).Return([]*HistoricVariableInstance{}, nil).Once()

		// 设置缓存设置
		mockCache.On("Set", ctx, "historic_process_instance:1", mock.AnythingOfType("string"), 2*time.Hour).Return(nil).Once()

		// 执行测试
		result, err := useCase.GetHistoricProcessInstance(ctx, instanceID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "1", result.ID)
		assert.Equal(t, "1", result.ProcessDefinitionID)
		assert.Equal(t, "test-process", result.ProcessDefinitionKey)
		assert.Equal(t, "business-123", result.BusinessKey)

		// 验证模拟对象调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

// TestHistoricDataUseCase_GetProcessStatistics 测试获取流程统计信息
func TestHistoricDataUseCase_GetProcessStatistics(t *testing.T) {
	mockRepo := new(MockHistoricProcessInstanceRepo)
	mockCache := new(MockCacheRepo)
	logger := zap.NewNop()

	useCase := NewHistoricDataUseCase(mockRepo, mockCache, logger)

	ctx := context.Background()
	req := &ProcessStatisticsRequest{
		ProcessDefinitionKey: "test-process",
		StartTime:            time.Now().AddDate(0, -1, 0),
		EndTime:              time.Now(),
	}

	t.Run("成功获取流程统计信息", func(t *testing.T) {
		// 设置缓存未命中
		mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return("", assert.AnError).Once()

		// 准备统计数据
		stats := &ProcessStatistics{
			TotalInstances:      100,
			CompletedInstances:  80,
			ActiveInstances:     15,
			SuspendedInstances:  3,
			TerminatedInstances: 2,
			AverageDuration:     time.Hour,
			MinDuration:         time.Minute * 30,
			MaxDuration:         time.Hour * 2,
		}

		// 设置数据库查询返回
		mockRepo.On("GetProcessStatistics", ctx, req.ProcessDefinitionKey, req.StartTime, req.EndTime).Return(stats, nil).Once()

		// 设置缓存设置
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 1*time.Hour).Return(nil).Once()

		// 执行测试
		result, err := useCase.GetProcessStatistics(ctx, req)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.ProcessDefinitionKey, result.ProcessDefinitionKey)
		assert.Equal(t, int64(100), result.TotalInstances)
		assert.Equal(t, int64(80), result.CompletedInstances)
		assert.Equal(t, float64(80), result.CompletionRate) // 80/100 * 100 = 80%

		// 验证模拟对象调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

// TestHistoricDataUseCase_DeleteHistoricProcessInstance 测试删除历史流程实例
func TestHistoricDataUseCase_DeleteHistoricProcessInstance(t *testing.T) {
	mockRepo := new(MockHistoricProcessInstanceRepo)
	mockCache := new(MockCacheRepo)
	logger := zap.NewNop()

	useCase := NewHistoricDataUseCase(mockRepo, mockCache, logger)

	ctx := context.Background()
	instanceID := int64(1)

	t.Run("成功删除历史流程实例", func(t *testing.T) {
		// 准备测试数据
		historicInstance := &ent.HistoricProcessInstance{
			ID: 1,
		}

		// 设置存在性检查
		mockRepo.On("GetHistoricProcessInstance", ctx, instanceID).Return(historicInstance, nil).Once()

		// 设置删除操作
		mockRepo.On("DeleteHistoricProcessInstance", ctx, instanceID).Return(nil).Once()

		// 设置缓存删除
		mockCache.On("Delete", ctx, "historic_process_instance:1").Return(nil).Once()

		// 执行测试
		err := useCase.DeleteHistoricProcessInstance(ctx, instanceID)

		// 验证结果
		assert.NoError(t, err)

		// 验证模拟对象调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}
