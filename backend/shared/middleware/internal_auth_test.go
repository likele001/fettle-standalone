package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func getInternal(r *gin.Engine, remoteAddr, token, xff string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/internal", nil)
	req.RemoteAddr = remoteAddr
	if token != "" {
		req.Header.Set("X-Internal-Token", token)
	}
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func newInternalRouter(h gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.GET("/internal", h, func(c *gin.Context) { c.String(200, "ok") })
	return r
}

func TestInternalAuthWithToken(t *testing.T) {
	r := newInternalRouter(InternalAuthFromToken("s3cret"))
	for _, tc := range []struct {
		name  string
		token string
		want  int
	}{
		{"valid", "s3cret", 200},
		{"wrong", "nope", http.StatusUnauthorized},
		{"missing", "", http.StatusUnauthorized},
		{"prefix of expected", "s3cr", http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 非本机来源：只有 token 正确才放行
			w := getInternal(r, "10.0.0.5:1234", tc.token, "")
			if w.Code != tc.want {
				t.Fatalf("got %d want %d body=%s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

// 未配置 token 时只放行真正的本机连接，X-Forwarded-For 伪造无效。
func TestInternalAuthLoopbackFallbackIgnoresForwardedFor(t *testing.T) {
	r := newInternalRouter(InternalAuthFromToken(""))

	if w := getInternal(r, "127.0.0.1:5000", "", ""); w.Code != 200 {
		t.Fatalf("loopback should pass, got %d body=%s", w.Code, w.Body.String())
	}
	if w := getInternal(r, "[::1]:5000", "", ""); w.Code != 200 {
		t.Fatalf("ipv6 loopback should pass, got %d", w.Code)
	}
	if w := getInternal(r, "203.0.113.9:5000", "", "127.0.0.1"); w.Code != http.StatusForbidden {
		t.Fatalf("XFF spoof must be rejected, got %d", w.Code)
	}
	if w := getInternal(r, "203.0.113.9:5000", "", ""); w.Code != http.StatusForbidden {
		t.Fatalf("remote should be rejected, got %d", w.Code)
	}
}

func TestDirectLoopback(t *testing.T) {
	for addr, want := range map[string]bool{
		"127.0.0.1:9000":          true,
		"[::1]:9000":              true,
		"10.0.0.1:9000":           false,
		"127.0.0.1":               true,
		"nohost":                  false,
		"[::ffff:127.0.0.1]:9000": true,
	} {
		if got := DirectLoopback(addr); got != want {
			t.Errorf("DirectLoopback(%q)=%v want %v", addr, got, want)
		}
	}
}

// gin 默认信任所有代理；设置后仅回环来源的转发头生效。
func TestTrustLocalProxiesLimitsForwardedFor(t *testing.T) {
	r := gin.New()
	if err := TrustLocalProxies(r); err != nil {
		t.Fatalf("TrustLocalProxies: %v", err)
	}
	var seen string
	r.GET("/x", func(c *gin.Context) { seen = c.ClientIP(); c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/x", nil)
	req.RemoteAddr = "203.0.113.9:5000"
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	r.ServeHTTP(w, req)
	if seen != "203.0.113.9" {
		t.Fatalf("untrusted XFF honoured: ClientIP=%q", seen)
	}

	w = httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/x", nil)
	req2.RemoteAddr = "127.0.0.1:5000"
	req2.Header.Set("X-Forwarded-For", "203.0.113.9")
	r.ServeHTTP(w, req2)
	if seen != "203.0.113.9" {
		t.Fatalf("trusted XFF should be honoured: ClientIP=%q", seen)
	}
}
