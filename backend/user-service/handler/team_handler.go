package handler

import (
	"net/http"

	"ai-platform/shared/middleware"
	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// CreateInvitation 创建邀请
func (h *TeamHandler) CreateInvitation(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)
	userID, _ := middleware.GetUserID(c)

	var req service.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "参数错误"})
		return
	}

	resp, err := h.teamService.CreateInvitation(uuid.MustParse(tenantID), uuid.MustParse(userID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// ListMembers 成员列表
func (h *TeamHandler) ListMembers(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)

	status := c.Query("status")
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	resp, err := h.teamService.ListMembers(uuid.MustParse(tenantID), status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// UpdateMemberRole 修改成员角色
func (h *TeamHandler) UpdateMemberRole(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)
	userID, _ := middleware.GetUserID(c)
	memberID := c.Param("id")

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "参数错误"})
		return
	}

	err := h.teamService.UpdateMemberRole(uuid.MustParse(tenantID), uuid.MustParse(userID), uuid.MustParse(memberID), req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ResignMember 设为离职
func (h *TeamHandler) ResignMember(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)
	userID, _ := middleware.GetUserID(c)
	memberID := c.Param("id")

	err := h.teamService.ResignMember(uuid.MustParse(tenantID), uuid.MustParse(userID), uuid.MustParse(memberID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ListInvitations 邀请列表
func (h *TeamHandler) ListInvitations(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)

	invitations, err := h.teamService.ListInvitations(uuid.MustParse(tenantID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": invitations})
}

// CancelInvitation 取消邀请
func (h *TeamHandler) CancelInvitation(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)
	userID, _ := middleware.GetUserID(c)
	invitationID := c.Param("id")

	err := h.teamService.CancelInvitation(uuid.MustParse(tenantID), uuid.MustParse(userID), uuid.MustParse(invitationID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// AcceptInvitation 接受邀请（登录用户调用）
func (h *TeamHandler) AcceptInvitation(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "参数错误"})
		return
	}

	err := h.teamService.AcceptInvitation(uuid.MustParse(userID), req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// GetInvitationInfo 获取邀请信息（公开接口，无需登录）
func (h *TeamHandler) GetInvitationInfo(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "缺少邀请码"})
		return
	}

	inv, err := h.teamService.GetInvitation(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	// 获取租户名称
	var tenantName string
	h.teamService.DB.Table("tenants").Select("name").Where("id = ?", inv.TenantID).Scan(&tenantName)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"code":        inv.Code,
			"tenant_name": tenantName,
			"role":        inv.Role,
		},
	})
}

func parseInt(s string, defaultVal int) int {
	val := defaultVal
	if s != "" {
		if v, err := uuid.Parse(s); err == nil {
			_ = v
		}
		for _, ch := range s {
			if ch < '0' || ch > '9' {
				return defaultVal
			}
			val = val*10 + int(ch-'0')
		}
	}
	return val
}
