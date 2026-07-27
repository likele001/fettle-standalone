package repository

import (
	"gorm.io/gorm"
)

// PlatformStats holds platform-wide statistics
type PlatformStats struct {
	TotalTenants       int64 `json:"total_tenants"`
	TotalAgents        int64 `json:"total_agents"`
	TotalConversations int64 `json:"total_conversations"`
	TotalMessages      int64 `json:"total_messages"`
	ActiveChannels     int64 `json:"active_channels"`
}

// AdminRepository handles admin-level queries
type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// GetPlatformStats returns platform-wide statistics
func (r *AdminRepository) GetPlatformStats() (*PlatformStats, error) {
	stats := &PlatformStats{}

	// Total tenants
	r.db.Table("tenants").Where("deleted_at IS NULL").Count(&stats.TotalTenants)

	// Total agents
	r.db.Table("agents").Where("deleted_at IS NULL").Count(&stats.TotalAgents)

	// Total conversations
	r.db.Table("conversations").Where("deleted_at IS NULL").Count(&stats.TotalConversations)

	// Total messages
	r.db.Table("messages").Count(&stats.TotalMessages)

	// Active channels (status = active)
	r.db.Table("channels").Where("deleted_at IS NULL AND status = 'active'").Count(&stats.ActiveChannels)

	return stats, nil
}
