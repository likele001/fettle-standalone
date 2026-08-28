package router

import (
	"net/http"

	"ai-platform/skill-service/handler"
	"ai-platform/skill-service/middleware"
	"ai-platform/skill-service/service"

	"github.com/gin-gonic/gin"
)

func Setup(skillService *service.SkillService, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	skillHandler := handler.NewSkillHandler(skillService)
	adminHandler := handler.NewAdminHandler(skillService)

	api := r.Group("")
	api.Use(middleware.Auth(jwtSecret))
	{
		skills := api.Group("/skills")
		{
			skills.GET("", skillHandler.ListSkills)
			skills.GET("/installed", skillHandler.GetInstalledSkills)
			skills.GET("/:id", skillHandler.GetSkill)
			skills.POST("", skillHandler.CreateSkill)
			skills.PUT("/:id", skillHandler.UpdateSkill)
			skills.DELETE("/:id", skillHandler.DeleteSkill)

			skills.POST("/:id/install", skillHandler.InstallSkill)
			skills.POST("/:id/uninstall", skillHandler.UninstallSkill)
			skills.POST("/:id/toggle", skillHandler.ToggleSkillStatus)
			skills.GET("/:id/config", skillHandler.GetSkillConfig)
			skills.PUT("/:id/config", skillHandler.UpdateSkillConfig)
		}
	}

	// 内部服务端点（ai-engine 直连，不经 gateway JWT）
	internal := r.Group("/internal")
	internal.Use(middleware.InternalAuth())
	{
		internal.GET("/skills/installed", skillHandler.InternalInstalledSkills)
	}

	// Admin routes
	admin := r.Group("/admin")
	admin.Use(middleware.AdminAuth(jwtSecret))
	{
		admin.GET("/skills", adminHandler.ListAllSkills)
		admin.GET("/skills/stats", adminHandler.GetSkillStats)
	}

	return r
}
