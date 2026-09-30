package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestJWTAuthMiddleware_MissingHeader(t *testing.T) {
	r := gin.New()
	r.Use(JWTAuthMiddleware("test-secret"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_InvalidFormat(t *testing.T) {
	r := gin.New()
	r.Use(JWTAuthMiddleware("test-secret"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidToken")
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	r := gin.New()
	r.Use(JWTAuthMiddleware("test-secret"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_WrongSecret(t *testing.T) {
	claims := &Claims{
		UserID:   "user-1",
		TenantID: "tenant-1",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("correct-secret"))

	r := gin.New()
	r.Use(JWTAuthMiddleware("wrong-secret"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	claims := &Claims{
		UserID:   "user-1",
		TenantID: "tenant-1",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-secret"))

	var gotUserID, gotTenantID, gotRole interface{}
	r := gin.New()
	r.Use(JWTAuthMiddleware("test-secret"))
	r.GET("/test", func(c *gin.Context) {
		gotUserID, _ = c.Get("user_id")
		gotTenantID, _ = c.Get("tenant_id")
		gotRole, _ = c.Get("role")
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	uid, _ := gotUserID.(string)
	tid, _ := gotTenantID.(string)
	role, _ := gotRole.(string)
	if uid != "user-1" {
		t.Errorf("expected user_id=user-1, got %v", uid)
	}
	if tid != "tenant-1" {
		t.Errorf("expected tenant_id=tenant-1, got %v", tid)
	}
	if role != "admin" {
		t.Errorf("expected role=admin, got %v", role)
	}
}

func TestTenantMiddleware_NoExistingTenant(t *testing.T) {
	r := gin.New()
	r.Use(TenantMiddleware())
	r.GET("/test", func(c *gin.Context) {
		val, exists := c.Get("tenant_id")
		if !exists {
			t.Error("tenant_id should exist")
			return
		}
		if val != "" {
			t.Errorf("expected empty tenant_id, got %v", val)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestTenantMiddleware_ExistingTenant(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "existing-tenant")
		c.Next()
	})
	r.Use(TenantMiddleware())
	r.GET("/test", func(c *gin.Context) {
		val, _ := c.Get("tenant_id")
		if val != "existing-tenant" {
			t.Errorf("expected existing-tenant, got %v", val)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetTenantID(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		id, ok := GetTenantID(c)
		if !ok {
			t.Error("expected ok=true")
		}
		if id != "test-tenant" {
			t.Errorf("expected test-tenant, got %v", id)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
}

func TestGetTenantID_Missing(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		id, ok := GetTenantID(c)
		if ok {
			t.Error("expected ok=false")
		}
		if id != "" {
			t.Errorf("expected empty, got %v", id)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
}

func TestGetUserID(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		id, ok := GetUserID(c)
		if !ok {
			t.Error("expected ok=true")
		}
		if id != "test-user" {
			t.Errorf("expected test-user, got %v", id)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
}

// TestCORSMiddleware_DefaultDeny 未配置 ALLOWED_ORIGINS 时不得放行任何跨域。
// 安全策略：不回退到 "*"（全开放），避免生产环境误放行。
func TestCORSMiddleware_DefaultDeny(t *testing.T) {
	os.Unsetenv("ALLOWED_ORIGINS")

	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("未配置白名单时应拒绝跨域，却返回了 Access-Control-Allow-Origin=%q", got)
	}
}

// TestCORSMiddleware_AllowListedOrigin 白名单命中的 Origin 应被回显放行。
func TestCORSMiddleware_AllowListedOrigin(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "https://app.example.com,https://admin.example.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Errorf("白名单内 Origin 应被放行，实际 Access-Control-Allow-Origin=%q", got)
	}
}

// TestCORSMiddleware_RejectUnlistedOrigin 不在白名单内的 Origin 应被拒绝。
func TestCORSMiddleware_RejectUnlistedOrigin(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "https://app.example.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("白名单外 Origin 应被拒绝，实际 Access-Control-Allow-Origin=%q", got)
	}
}

// TestCORSMiddleware_Preflight OPTIONS 预检应返回 204 且带 CORS 方法声明。
func TestCORSMiddleware_Preflight(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "https://app.example.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("预检请求应返回 204，实际 %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("预检响应缺少 Access-Control-Allow-Methods")
	}
}

func TestTraceMiddleware_GeneratesID(t *testing.T) {
	r := gin.New()
	r.Use(TraceMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	traceID := w.Header().Get("X-Trace-ID")
	if traceID == "" {
		t.Error("expected trace ID to be generated")
	}
}

func TestTraceMiddleware_PreservesExistingID(t *testing.T) {
	r := gin.New()
	r.Use(TraceMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-ID", "my-trace-id")
	r.ServeHTTP(w, req)

	traceID := w.Header().Get("X-Trace-ID")
	if traceID != "my-trace-id" {
		t.Errorf("expected my-trace-id, got %v", traceID)
	}
}
