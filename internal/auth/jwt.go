// Package auth JWT认证系统
// 提供JWT令牌生成、验证和权限控制功能
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey       string        `yaml:"secret_key"`       // 密钥
	Issuer          string        `yaml:"issuer"`           // 签发者
	ExpirationTime  time.Duration `yaml:"expiration_time"`  // 访问令牌过期时间
	RefreshTime     time.Duration `yaml:"refresh_time"`     // 刷新令牌过期时间
	Algorithm       string        `yaml:"algorithm"`        // 签名算法
	EnableRefresh   bool          `yaml:"enable_refresh"`   // 启用刷新令牌
	EnableBlacklist bool          `yaml:"enable_blacklist"` // 启用黑名单
}

// UserClaims 用户声明
type UserClaims struct {
	UserID      int64    `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// JWTManager JWT管理器
type JWTManager struct {
	config    *JWTConfig
	logger    *zap.Logger
	blacklist map[string]time.Time // 简单的内存黑名单
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(config *JWTConfig, logger *zap.Logger) *JWTManager {
	if config.SecretKey == "" {
		// 生成随机密钥
		config.SecretKey = generateRandomSecret()
		logger.Warn("使用随机生成的JWT密钥，生产环境建议配置固定密钥")
	}

	if config.Algorithm == "" {
		config.Algorithm = "HS256"
	}

	if config.ExpirationTime == 0 {
		config.ExpirationTime = 24 * time.Hour
	}

	if config.RefreshTime == 0 {
		config.RefreshTime = 7 * 24 * time.Hour
	}

	return &JWTManager{
		config:    config,
		logger:    logger,
		blacklist: make(map[string]time.Time),
	}
}

// GenerateTokenPair 生成令牌对
func (j *JWTManager) GenerateTokenPair(userID int64, username, email string, roles, permissions []string) (*TokenPair, error) {
	now := time.Now()

	// 生成访问令牌
	accessClaims := &UserClaims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.ExpirationTime)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.GetSigningMethod(j.config.Algorithm), accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		j.logger.Error("生成访问令牌失败",
			zap.Int64("user_id", userID),
			zap.String("username", username),
			zap.Error(err),
		)
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	tokenPair := &TokenPair{
		AccessToken: accessTokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int64(j.config.ExpirationTime.Seconds()),
		ExpiresAt:   now.Add(j.config.ExpirationTime),
	}

	// 生成刷新令牌（如果启用）
	if j.config.EnableRefresh {
		refreshClaims := &UserClaims{
			UserID:   userID,
			Username: username,
			Email:    email,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    j.config.Issuer,
				Subject:   fmt.Sprintf("%d", userID),
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(j.config.RefreshTime)),
				NotBefore: jwt.NewNumericDate(now),
			},
		}

		refreshToken := jwt.NewWithClaims(jwt.GetSigningMethod(j.config.Algorithm), refreshClaims)
		refreshTokenString, err := refreshToken.SignedString([]byte(j.config.SecretKey))
		if err != nil {
			j.logger.Error("生成刷新令牌失败",
				zap.Int64("user_id", userID),
				zap.String("username", username),
				zap.Error(err),
			)
			return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
		}

		tokenPair.RefreshToken = refreshTokenString
	}

	j.logger.Info("令牌生成成功",
		zap.Int64("user_id", userID),
		zap.String("username", username),
		zap.Time("expires_at", tokenPair.ExpiresAt),
		zap.Bool("has_refresh_token", tokenPair.RefreshToken != ""),
	)

	return tokenPair, nil
}

// ValidateToken 验证令牌
func (j *JWTManager) ValidateToken(tokenString string) (*UserClaims, error) {
	// 检查黑名单
	if j.config.EnableBlacklist && j.isTokenBlacklisted(tokenString) {
		return nil, fmt.Errorf("令牌已被吊销")
	}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名算法: %v", token.Header["alg"])
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		j.logger.Warn("令牌验证失败",
			zap.String("error", err.Error()),
		)
		return nil, fmt.Errorf("令牌验证失败: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		j.logger.Warn("令牌声明无效")
		return nil, fmt.Errorf("令牌声明无效")
	}

	// 验证令牌是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		j.logger.Warn("令牌已过期",
			zap.Int64("user_id", claims.UserID),
			zap.Time("expires_at", claims.ExpiresAt.Time),
		)
		return nil, fmt.Errorf("令牌已过期")
	}

	j.logger.Debug("令牌验证成功",
		zap.Int64("user_id", claims.UserID),
		zap.String("username", claims.Username),
		zap.Strings("roles", claims.Roles),
	)

	return claims, nil
}

// RefreshToken 刷新令牌
func (j *JWTManager) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	if !j.config.EnableRefresh {
		return nil, fmt.Errorf("刷新令牌功能未启用")
	}

	// 验证刷新令牌
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("刷新令牌验证失败: %w", err)
	}

	// 将旧的刷新令牌加入黑名单
	if j.config.EnableBlacklist {
		j.blacklistToken(refreshTokenString, claims.ExpiresAt.Time)
	}

	// 生成新的令牌对
	return j.GenerateTokenPair(claims.UserID, claims.Username, claims.Email, claims.Roles, claims.Permissions)
}

// RevokeToken 吊销令牌
func (j *JWTManager) RevokeToken(tokenString string) error {
	if !j.config.EnableBlacklist {
		j.logger.Warn("令牌黑名单功能未启用，无法吊销令牌")
		return fmt.Errorf("令牌黑名单功能未启用")
	}

	// 解析令牌获取过期时间
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return fmt.Errorf("解析令牌失败: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return fmt.Errorf("令牌声明无效")
	}

	// 加入黑名单
	j.blacklistToken(tokenString, claims.ExpiresAt.Time)

	j.logger.Info("令牌已吊销",
		zap.Int64("user_id", claims.UserID),
		zap.String("username", claims.Username),
	)

	return nil
}

// HasPermission 检查权限
func (j *JWTManager) HasPermission(claims *UserClaims, requiredPermission string) bool {
	// 检查角色权限
	for _, role := range claims.Roles {
		if role == "admin" || role == "super_admin" {
			return true // 管理员拥有所有权限
		}
	}

	// 检查具体权限
	for _, permission := range claims.Permissions {
		if permission == requiredPermission || permission == "*" {
			return true
		}

		// 支持通配符权限检查
		if strings.HasSuffix(permission, "*") {
			prefix := strings.TrimSuffix(permission, "*")
			if strings.HasPrefix(requiredPermission, prefix) {
				return true
			}
		}
	}

	return false
}

// HasRole 检查角色
func (j *JWTManager) HasRole(claims *UserClaims, requiredRole string) bool {
	for _, role := range claims.Roles {
		if role == requiredRole {
			return true
		}
	}
	return false
}

// ExtractTokenFromHeader 从请求头提取令牌
func (j *JWTManager) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("认证头为空")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("认证头格式无效，期望格式: Bearer <token>")
	}

	return parts[1], nil
}

// blacklistToken 将令牌加入黑名单
func (j *JWTManager) blacklistToken(tokenString string, expiresAt time.Time) {
	j.blacklist[tokenString] = expiresAt

	// 清理过期的黑名单条目
	j.cleanupBlacklist()
}

// isTokenBlacklisted 检查令牌是否在黑名单中
func (j *JWTManager) isTokenBlacklisted(tokenString string) bool {
	expiresAt, exists := j.blacklist[tokenString]
	if !exists {
		return false
	}

	// 如果令牌已过期，从黑名单中移除
	if time.Now().After(expiresAt) {
		delete(j.blacklist, tokenString)
		return false
	}

	return true
}

// cleanupBlacklist 清理过期的黑名单条目
func (j *JWTManager) cleanupBlacklist() {
	now := time.Now()
	for token, expiresAt := range j.blacklist {
		if now.After(expiresAt) {
			delete(j.blacklist, token)
		}
	}
}

// GetBlacklistStats 获取黑名单统计信息
func (j *JWTManager) GetBlacklistStats() map[string]interface{} {
	j.cleanupBlacklist()

	return map[string]interface{}{
		"total_blacklisted": len(j.blacklist),
		"enabled":           j.config.EnableBlacklist,
	}
}

// generateRandomSecret 生成随机密钥
func generateRandomSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用固定的默认值
		return "workflow-engine-default-secret-key-please-change-in-production"
	}
	return hex.EncodeToString(bytes)
}

// 预定义的JWT配置
var (
	// DefaultJWTConfig 默认JWT配置
	DefaultJWTConfig = &JWTConfig{
		SecretKey:       "", // 将自动生成
		Issuer:          "workflow-engine",
		ExpirationTime:  24 * time.Hour,
		RefreshTime:     7 * 24 * time.Hour,
		Algorithm:       "HS256",
		EnableRefresh:   true,
		EnableBlacklist: true,
	}

	// ProductionJWTConfig 生产环境JWT配置
	ProductionJWTConfig = &JWTConfig{
		SecretKey:       "", // 生产环境必须设置
		Issuer:          "workflow-engine-prod",
		ExpirationTime:  2 * time.Hour,
		RefreshTime:     24 * time.Hour,
		Algorithm:       "HS256",
		EnableRefresh:   true,
		EnableBlacklist: true,
	}

	// TestJWTConfig 测试环境JWT配置
	TestJWTConfig = &JWTConfig{
		SecretKey:       "test-secret-key-for-testing-only",
		Issuer:          "workflow-engine-test",
		ExpirationTime:  1 * time.Hour,
		RefreshTime:     24 * time.Hour,
		Algorithm:       "HS256",
		EnableRefresh:   false,
		EnableBlacklist: false,
	}
)
