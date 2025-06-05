// Package middleware JWT认证中间件
// 提供HTTP请求的JWT认证和权限控制中间件
package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"workflow-engine/internal/auth"
)

// AuthMiddleware 认证中间件配置
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
	logger     *zap.Logger
	skipPaths  map[string]bool // 跳过认证的路径
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtManager *auth.JWTManager, logger *zap.Logger) *AuthMiddleware {
	// 默认跳过认证的路径
	skipPaths := map[string]bool{
		"/health":               true,
		"/metrics":              true,
		"/api/v1/auth/login":    true,
		"/api/v1/auth/register": true,
		"/api/v1/docs":          true,
		"/swagger":              true,
	}

	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
		skipPaths:  skipPaths,
	}
}

// JWTAuth JWT认证中间件
func (m *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要跳过认证
		if m.shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn("缺少认证头",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40001,
				"message": "缺少认证头",
			})
			c.Abort()
			return
		}

		// 提取令牌
		tokenString, err := m.jwtManager.ExtractTokenFromHeader(authHeader)
		if err != nil {
			m.logger.Warn("认证头格式无效",
				zap.String("auth_header", authHeader),
				zap.String("path", c.Request.URL.Path),
				zap.Error(err),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40002,
				"message": "认证头格式无效",
			})
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			m.logger.Warn("令牌验证失败",
				zap.String("path", c.Request.URL.Path),
				zap.Error(err),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40003,
				"message": "令牌验证失败",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_roles", claims.Roles)
		c.Set("user_permissions", claims.Permissions)

		m.logger.Debug("用户认证成功",
			zap.Int64("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.String("path", c.Request.URL.Path),
		)

		c.Next()
	}
}

// RequirePermission 权限验证中间件
func (m *AuthMiddleware) RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户声明
		claims, exists := c.Get("user_claims")
		if !exists {
			m.logger.Error("用户认证信息丢失",
				zap.String("path", c.Request.URL.Path),
				zap.String("required_permission", requiredPermission),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40001,
				"message": "用户认证信息丢失",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.UserClaims)
		if !ok {
			m.logger.Error("用户认证信息格式错误",
				zap.String("path", c.Request.URL.Path),
				zap.String("required_permission", requiredPermission),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50001,
				"message": "用户认证信息格式错误",
			})
			c.Abort()
			return
		}

		// 检查权限
		if !m.jwtManager.HasPermission(userClaims, requiredPermission) {
			m.logger.Warn("权限不足",
				zap.Int64("user_id", userClaims.UserID),
				zap.String("username", userClaims.Username),
				zap.String("path", c.Request.URL.Path),
				zap.String("required_permission", requiredPermission),
				zap.Strings("user_permissions", userClaims.Permissions),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40301,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		m.logger.Debug("权限验证通过",
			zap.Int64("user_id", userClaims.UserID),
			zap.String("username", userClaims.Username),
			zap.String("path", c.Request.URL.Path),
			zap.String("required_permission", requiredPermission),
		)

		c.Next()
	}
}

// RequireRole 角色验证中间件
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户声明
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40001,
				"message": "用户认证信息丢失",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.UserClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50001,
				"message": "用户认证信息格式错误",
			})
			c.Abort()
			return
		}

		// 检查角色
		if !m.jwtManager.HasRole(userClaims, requiredRole) {
			m.logger.Warn("角色权限不足",
				zap.Int64("user_id", userClaims.UserID),
				zap.String("username", userClaims.Username),
				zap.String("path", c.Request.URL.Path),
				zap.String("required_role", requiredRole),
				zap.Strings("user_roles", userClaims.Roles),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40302,
				"message": "角色权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole 要求任意角色中间件
func (m *AuthMiddleware) RequireAnyRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40001,
				"message": "用户认证信息丢失",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.UserClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50001,
				"message": "用户认证信息格式错误",
			})
			c.Abort()
			return
		}

		// 检查是否拥有任意一个所需角色
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if m.jwtManager.HasRole(userClaims, requiredRole) {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			m.logger.Warn("缺少所需角色",
				zap.Int64("user_id", userClaims.UserID),
				zap.String("username", userClaims.Username),
				zap.String("path", c.Request.URL.Path),
				zap.Strings("required_roles", requiredRoles),
				zap.Strings("user_roles", userClaims.Roles),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40303,
				"message": "缺少所需角色",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 要求任意权限中间件
func (m *AuthMiddleware) RequireAnyPermission(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40001,
				"message": "用户认证信息丢失",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.UserClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50001,
				"message": "用户认证信息格式错误",
			})
			c.Abort()
			return
		}

		// 检查是否拥有任意一个所需权限
		hasRequiredPermission := false
		for _, requiredPermission := range requiredPermissions {
			if m.jwtManager.HasPermission(userClaims, requiredPermission) {
				hasRequiredPermission = true
				break
			}
		}

		if !hasRequiredPermission {
			m.logger.Warn("缺少所需权限",
				zap.Int64("user_id", userClaims.UserID),
				zap.String("username", userClaims.Username),
				zap.String("path", c.Request.URL.Path),
				zap.Strings("required_permissions", requiredPermissions),
				zap.Strings("user_permissions", userClaims.Permissions),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40304,
				"message": "缺少所需权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly 仅管理员中间件
func (m *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return m.RequireAnyRole("admin", "super_admin")
}

// AddSkipPath 添加跳过认证的路径
func (m *AuthMiddleware) AddSkipPath(path string) {
	m.skipPaths[path] = true
}

// RemoveSkipPath 移除跳过认证的路径
func (m *AuthMiddleware) RemoveSkipPath(path string) {
	delete(m.skipPaths, path)
}

// shouldSkipAuth 检查是否应该跳过认证
func (m *AuthMiddleware) shouldSkipAuth(path string) bool {
	// 精确匹配
	if skip, exists := m.skipPaths[path]; exists && skip {
		return true
	}

	// 前缀匹配
	for skipPath := range m.skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// GetCurrentUser 获取当前用户信息的辅助函数
func GetCurrentUser(c *gin.Context) (*auth.UserClaims, error) {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil, fmt.Errorf("用户认证信息不存在")
	}

	userClaims, ok := claims.(*auth.UserClaims)
	if !ok {
		return nil, fmt.Errorf("用户认证信息格式错误")
	}

	return userClaims, nil
}

// GetCurrentUserID 获取当前用户ID的辅助函数
func GetCurrentUserID(c *gin.Context) (int64, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("用户ID不存在")
	}

	id, ok := userID.(int64)
	if !ok {
		return 0, fmt.Errorf("用户ID格式错误")
	}

	return id, nil
}

// GetCurrentUsername 获取当前用户名的辅助函数
func GetCurrentUsername(c *gin.Context) (string, error) {
	username, exists := c.Get("username")
	if !exists {
		return "", fmt.Errorf("用户名不存在")
	}

	name, ok := username.(string)
	if !ok {
		return "", fmt.Errorf("用户名格式错误")
	}

	return name, nil
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有认证头，继续执行但不设置用户信息
			c.Next()
			return
		}

		// 有认证头，尝试解析
		tokenString, err := m.jwtManager.ExtractTokenFromHeader(authHeader)
		if err != nil {
			// 认证头格式错误，继续执行但不设置用户信息
			c.Next()
			return
		}

		// 验证令牌
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			// 令牌无效，继续执行但不设置用户信息
			c.Next()
			return
		}

		// 设置用户信息
		c.Set("user_claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_roles", claims.Roles)
		c.Set("user_permissions", claims.Permissions)

		c.Next()
	}
}

// RateLimitByUser 按用户限流中间件
func (m *AuthMiddleware) RateLimitByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			c.Next()
			return
		}

		userClaims, ok := claims.(*auth.UserClaims)
		if !ok {
			c.Next()
			return
		}

		// 这里可以实现按用户的限流逻辑
		// 例如检查Redis中的用户请求计数
		// 暂时跳过具体实现

		m.logger.Debug("用户请求记录",
			zap.Int64("user_id", userClaims.UserID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		c.Next()
	}
}
