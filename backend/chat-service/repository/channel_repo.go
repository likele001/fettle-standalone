package repository

import (
	"ai-platform/chat-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) ListByTenant(tenantID string) ([]models.Channel, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	var channels []models.Channel
	err = r.db.Where("tenant_id = ?", tID).Find(&channels).Error
	return channels, err
}

func (r *ChannelRepository) GetByID(tenantID, channelID string) (*models.Channel, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	cID, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}
	var ch models.Channel
	err = r.db.Where("id = ? AND tenant_id = ?", cID, tID).First(&ch).Error
	return &ch, err
}

func (r *ChannelRepository) GetByType(tenantID, channelType string) (*models.Channel, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	var ch models.Channel
	err = r.db.Where("tenant_id = ? AND type = ?", tID, channelType).First(&ch).Error
	return &ch, err
}

func (r *ChannelRepository) Upsert(ch *models.Channel) error {
	var existing models.Channel
	result := r.db.Where("tenant_id = ? AND type = ?", ch.TenantID, ch.Type).First(&existing)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return r.db.Create(ch).Error
		}
		return result.Error
	}
	ch.ID = existing.ID
	return r.db.Save(ch).Error
}

func (r *ChannelRepository) Update(ch *models.Channel) error {
	return r.db.Save(ch).Error
}


// ListAllChannels returns all channels across all tenants (for admin)
func (r *ChannelRepository) ListAllChannels() ([]models.Channel, error) {
	var channels []models.Channel
	err := r.db.Order("created_at DESC").Find(&channels).Error
	return channels, err
}

// ChannelStat represents a channel type group count
type ChannelStat struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// GetChannelStats returns channel counts grouped by type (for admin)
func (r *ChannelRepository) GetChannelStats() ([]ChannelStat, error) {
	var stats []ChannelStat
	err := r.db.Model(&models.Channel{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Order("count DESC").
		Scan(&stats).Error
	return stats, err
}
