package cache

import "fmt"

// 缓存键定义 - 统一管理所有 Redis 键的命名规范

const (
	// 用户相关
	UserKeyPrefix       = "user:"
	UserInfoKey         = UserKeyPrefix + "info:%s"         // user:info:{user_id}
	UserTokenKey        = UserKeyPrefix + "token:%s"        // user:token:{user_id}
	UserSessionKey      = UserKeyPrefix + "session:%s"      // user:session:{session_id}
	UserPermissionsKey  = UserKeyPrefix + "perms:%s"        // user:perms:{user_id}

	// 租户相关
	TenantKeyPrefix     = "tenant:"
	TenantInfoKey       = TenantKeyPrefix + "info:%s"       // tenant:info:{tenant_id}
	TenantConfigKey     = TenantKeyPrefix + "config:%s"     // tenant:config:{tenant_id}
	TenantQuotaKey      = TenantKeyPrefix + "quota:%s"      // tenant:quota:{tenant_id}

	// 认证相关
	AuthKeyPrefix       = "auth:"
	AuthTokenBlacklist  = AuthKeyPrefix + "blacklist:%s"    // auth:blacklist:{token}
	AuthRefreshToken    = AuthKeyPrefix + "refresh:%s"      // auth:refresh:{user_id}
	AuthVerifyCode      = AuthKeyPrefix + "code:%s"         // auth:code:{phone}

	// 限流相关
	RateLimitKeyPrefix  = "ratelimit:"
	RateLimitIPKey      = RateLimitKeyPrefix + "ip:%s"      // ratelimit:ip:{ip}
	RateLimitUserKey    = RateLimitKeyPrefix + "user:%s"    // ratelimit:user:{user_id}
	RateLimitTenantKey  = RateLimitKeyPrefix + "tenant:%s"  // ratelimit:tenant:{tenant_id}

	// 会话相关
	ChatKeyPrefix       = "chat:"
	ChatContextKey      = ChatKeyPrefix + "ctx:%s"          // chat:ctx:{conversation_id}
	ChatMessageKey      = ChatKeyPrefix + "msg:%s"          // chat:msg:{conversation_id}
	ChatSessionIndex    = ChatKeyPrefix + "session_idx:%s"  // chat:session_idx:{tenant_id}

	// 智能体相关
	AgentKeyPrefix      = "agent:"
	AgentConfigKey      = AgentKeyPrefix + "config:%s"      // agent:config:{agent_id}
	AgentStatusKey      = AgentKeyPrefix + "status:%s"      // agent:status:{agent_id}

	// 模型路由缓存
	ModelKeyPrefix      = "model:"
	ModelRouteCache     = ModelKeyPrefix + "route:%s"       // model:route:{input_hash}
	ModelTokenCount     = ModelKeyPrefix + "tokens:%s"      // model:tokens:{tenant_id}
)

// FormatKey 格式化缓存键
func FormatKey(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}

// UserInfoKeyFormat 用户信息键
func UserInfoKeyFormat(userID string) string {
	return fmt.Sprintf(UserInfoKey, userID)
}

// TenantInfoKeyFormat 租户信息键
func TenantInfoKeyFormat(tenantID string) string {
	return fmt.Sprintf(TenantInfoKey, tenantID)
}

// AuthTokenBlacklistFormat Token黑名单键
func AuthTokenBlacklistFormat(token string) string {
	return fmt.Sprintf(AuthTokenBlacklist, token)
}

// AuthVerifyCodeFormat 验证码键
func AuthVerifyCodeFormat(phone string) string {
	return fmt.Sprintf(AuthVerifyCode, phone)
}

// RateLimitIPKeyFormat IP限流键
func RateLimitIPKeyFormat(ip string) string {
	return fmt.Sprintf(RateLimitIPKey, ip)
}

// RateLimitTenantKeyFormat 租户限流键
func RateLimitTenantKeyFormat(tenantID string) string {
	return fmt.Sprintf(RateLimitTenantKey, tenantID)
}

// ChatContextKeyFormat 会话上下文键
func ChatContextKeyFormat(conversationID string) string {
	return fmt.Sprintf(ChatContextKey, conversationID)
}

// AgentConfigKeyFormat 智能体配置键
func AgentConfigKeyFormat(agentID string) string {
	return fmt.Sprintf(AgentConfigKey, agentID)
}

// ModelRouteCacheFormat 模型路由缓存键
func ModelRouteCacheFormat(inputHash string) string {
	return fmt.Sprintf(ModelRouteCache, inputHash)
}
