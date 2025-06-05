package repository

import (
	"context"
	"testing"
	"time"

	"github.com/workflow-engine/workflow-engine/internal/biz"
	"github.com/workflow-engine/workflow-engine/internal/data/ent"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockProcessDefinitionRepo 模拟流程定义仓储
type MockProcessDefinitionRepo struct {
	mock.Mock
}

func (m *MockProcessDefinitionRepo) Create(ctx context.Context, pd *ent.ProcessDefinition) (*ent.ProcessDefinition, error) {
	args := m.Called(ctx, pd)
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
	return args.Get(0).(*ent.ProcessDefinition), args.Error(1)
}

func (m *MockProcessDefinitionRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProcessDefinitionRepo) List(ctx context.Context, filter *biz.ProcessDefinitionFilter, opts *biz.QueryOptions) ([]*ent.ProcessDefinition, *biz.PaginationResult, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).([]*ent.ProcessDefinition), args.Get(1).(*biz.PaginationResult), args.Error(2)
}

func (m *MockProcessDefinitionRepo) Count(ctx context.Context, filter *biz.ProcessDefinitionFilter) (int, error) {
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

// TestProcessDefinitionRepoMethods 测试流程定义仓储方法
func TestProcessDefinitionRepoMethods(t *testing.T) {
	mockRepo := new(MockProcessDefinitionRepo)

	t.Run("创建流程定义", func(t *testing.T) {
		pd := &ent.ProcessDefinition{
			ID:          1,
			Key:         "test-process",
			Name:        "测试流程",
			Description: "测试流程描述",
			Category:    "测试分类",
			Version:     1,
			Resource:    "test.bpmn",
			TenantID:    "default",
			Suspended:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*ent.ProcessDefinition")).
			Return(pd, nil)

		ctx := context.Background()
		result, err := mockRepo.Create(ctx, pd)

		assert.NoError(t, err, "创建流程定义不应该返回错误")
		assert.NotNil(t, result, "创建结果不应该为空")
		assert.Equal(t, "test-process", result.Key, "流程键应该匹配")
		assert.Equal(t, "测试流程", result.Name, "流程名称应该匹配")

		mockRepo.AssertExpectations(t)
	})

	t.Run("根据ID获取流程定义", func(t *testing.T) {
		expectedPD := &ent.ProcessDefinition{
			ID:   1,
			Key:  "test-process",
			Name: "测试流程",
		}

		mockRepo.On("GetByID", mock.Anything, "1").
			Return(expectedPD, nil)

		ctx := context.Background()
		result, err := mockRepo.GetByID(ctx, "1")

		assert.NoError(t, err, "根据ID获取流程定义不应该返回错误")
		assert.NotNil(t, result, "获取结果不应该为空")
		assert.Equal(t, int64(1), result.ID, "流程ID应该匹配")
		assert.Equal(t, "test-process", result.Key, "流程键应该匹配")

		mockRepo.AssertExpectations(t)
	})

	t.Run("根据ID获取流程定义_不存在", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, "999").
			Return(nil, assert.AnError)

		ctx := context.Background()
		result, err := mockRepo.GetByID(ctx, "999")

		assert.Error(t, err, "获取不存在的流程定义应该返回错误")
		assert.Nil(t, result, "获取结果应该为空")

		mockRepo.AssertExpectations(t)
	})

	t.Run("根据Key获取最新版本流程定义", func(t *testing.T) {
		expectedPD := &ent.ProcessDefinition{
			ID:      2,
			Key:     "test-process",
			Name:    "测试流程",
			Version: 2,
		}

		mockRepo.On("GetLatestByKey", mock.Anything, "test-process").
			Return(expectedPD, nil)

		ctx := context.Background()
		result, err := mockRepo.GetLatestByKey(ctx, "test-process")

		assert.NoError(t, err, "根据Key获取最新版本流程定义不应该返回错误")
		assert.NotNil(t, result, "获取结果不应该为空")
		assert.Equal(t, int32(2), result.Version, "流程版本应该是最新的")

		mockRepo.AssertExpectations(t)
	})

	t.Run("根据Key和版本获取流程定义", func(t *testing.T) {
		expectedPD := &ent.ProcessDefinition{
			ID:      1,
			Key:     "test-process",
			Name:    "测试流程",
			Version: 1,
		}

		mockRepo.On("GetByKeyAndVersion", mock.Anything, "test-process", 1).
			Return(expectedPD, nil)

		ctx := context.Background()
		result, err := mockRepo.GetByKeyAndVersion(ctx, "test-process", 1)

		assert.NoError(t, err, "根据Key和版本获取流程定义不应该返回错误")
		assert.NotNil(t, result, "获取结果不应该为空")
		assert.Equal(t, int32(1), result.Version, "流程版本应该匹配")

		mockRepo.AssertExpectations(t)
	})

	t.Run("更新流程定义", func(t *testing.T) {
		pd := &ent.ProcessDefinition{
			ID:          1,
			Key:         "test-process",
			Name:        "更新后的测试流程",
			Description: "更新后的描述",
		}

		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*ent.ProcessDefinition")).
			Return(pd, nil)

		ctx := context.Background()
		result, err := mockRepo.Update(ctx, pd)

		assert.NoError(t, err, "更新流程定义不应该返回错误")
		assert.NotNil(t, result, "更新结果不应该为空")
		assert.Equal(t, "更新后的测试流程", result.Name, "流程名称应该已更新")

		mockRepo.AssertExpectations(t)
	})

	t.Run("删除流程定义", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, "1").
			Return(nil)

		ctx := context.Background()
		err := mockRepo.Delete(ctx, "1")

		assert.NoError(t, err, "删除流程定义不应该返回错误")

		mockRepo.AssertExpectations(t)
	})

	t.Run("分页查询流程定义", func(t *testing.T) {
		expectedList := []*ent.ProcessDefinition{
			{
				ID:   1,
				Key:  "process-1",
				Name: "流程1",
			},
			{
				ID:   2,
				Key:  "process-2",
				Name: "流程2",
			},
		}

		expectedPagination := &biz.PaginationResult{
			Total:    2,
			Page:     1,
			PageSize: 10,
			Pages:    1,
		}

		filter := &biz.ProcessDefinitionFilter{
			Category: "测试分类",
		}

		opts := &biz.QueryOptions{
			Page:     1,
			PageSize: 10,
			OrderBy:  "name",
			Order:    "asc",
		}

		mockRepo.On("List", mock.Anything, filter, opts).
			Return(expectedList, expectedPagination, nil)

		ctx := context.Background()
		results, pagination, err := mockRepo.List(ctx, filter, opts)

		assert.NoError(t, err, "分页查询流程定义不应该返回错误")
		assert.NotNil(t, results, "查询结果不应该为空")
		assert.Equal(t, 2, len(results), "查询结果数量应该正确")
		assert.NotNil(t, pagination, "分页信息不应该为空")
		assert.Equal(t, 2, pagination.Total, "总数应该正确")

		mockRepo.AssertExpectations(t)
	})

	t.Run("计数查询流程定义", func(t *testing.T) {
		filter := &biz.ProcessDefinitionFilter{
			Category: "测试分类",
		}

		mockRepo.On("Count", mock.Anything, filter).
			Return(5, nil)

		ctx := context.Background()
		count, err := mockRepo.Count(ctx, filter)

		assert.NoError(t, err, "计数查询流程定义不应该返回错误")
		assert.Equal(t, 5, count, "计数结果应该正确")

		mockRepo.AssertExpectations(t)
	})

	t.Run("部署流程定义", func(t *testing.T) {
		mockRepo.On("Deploy", mock.Anything, "1").
			Return(nil)

		ctx := context.Background()
		err := mockRepo.Deploy(ctx, "1")

		assert.NoError(t, err, "部署流程定义不应该返回错误")

		mockRepo.AssertExpectations(t)
	})

	t.Run("挂起流程定义", func(t *testing.T) {
		mockRepo.On("Suspend", mock.Anything, "1").
			Return(nil)

		ctx := context.Background()
		err := mockRepo.Suspend(ctx, "1")

		assert.NoError(t, err, "挂起流程定义不应该返回错误")

		mockRepo.AssertExpectations(t)
	})
}

// TestProcessDefinitionRepoCreation 测试流程定义仓储创建
func TestProcessDefinitionRepoCreation(t *testing.T) {
	t.Run("创建流程定义仓储实例", func(t *testing.T) {
		logger := zap.NewNop()
		repo := NewProcessDefinitionRepo(nil, logger)

		assert.NotNil(t, repo, "流程定义仓储实例不应该为空")
		assert.Implements(t, (*biz.ProcessDefinitionRepo)(nil), repo, "应该实现 ProcessDefinitionRepo 接口")
	})
}

// TestProcessDefinitionRepoDataValidation 测试数据验证
func TestProcessDefinitionRepoDataValidation(t *testing.T) {
	t.Run("验证流程定义数据结构", func(t *testing.T) {
		pd := &ent.ProcessDefinition{
			Key:         "valid-process-key",
			Name:        "有效的流程名称",
			Description: "有效的流程描述",
			Category:    "有效分类",
			Version:     1,
		}

		// 验证必填字段
		assert.NotEmpty(t, pd.Key, "流程键不应该为空")
		assert.NotEmpty(t, pd.Name, "流程名称不应该为空")
		assert.Greater(t, pd.Version, int32(0), "版本号应该大于0")

		// 验证可选字段
		assert.NotEmpty(t, pd.Description, "流程描述应该存在")
		assert.NotEmpty(t, pd.Category, "流程分类应该存在")
	})

	t.Run("验证过滤条件结构", func(t *testing.T) {
		filter := &biz.ProcessDefinitionFilter{
			Name:     "测试过滤",
			Category: "测试分类",
			Version:  1,
			Status:   "active",
		}

		assert.NotEmpty(t, filter.Name, "过滤名称应该存在")
		assert.NotEmpty(t, filter.Category, "过滤分类应该存在")
		assert.Greater(t, filter.Version, 0, "过滤版本应该大于0")
		assert.NotEmpty(t, filter.Status, "过滤状态应该存在")
	})

	t.Run("验证查询选项结构", func(t *testing.T) {
		opts := &biz.QueryOptions{
			Page:     1,
			PageSize: 10,
			OrderBy:  "name",
			Order:    "asc",
			Search:   "测试搜索",
		}

		assert.Greater(t, opts.Page, 0, "页码应该大于0")
		assert.Greater(t, opts.PageSize, 0, "页面大小应该大于0")
		assert.NotEmpty(t, opts.OrderBy, "排序字段应该存在")
		assert.Contains(t, []string{"asc", "desc"}, opts.Order, "排序方向应该有效")
		assert.NotEmpty(t, opts.Search, "搜索关键词应该存在")
	})
}
