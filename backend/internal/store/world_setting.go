package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
)

func CreateWorldSetting(ctx context.Context, setting *model.WorldSetting) error {
	if err := DB.Create(setting).Error; err != nil {
		return err
	}
	setEmbedding(ctx, "world_settings", setting.ID, worldSettingEmbeddingText(setting))
	return nil
}

func UpdateWorldSetting(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if err := DB.Model(&model.WorldSetting{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	var setting model.WorldSetting
	if err := DB.Where("id = ?", id).First(&setting).Error; err != nil {
		return err
	}
	setEmbedding(ctx, "world_settings", id, worldSettingEmbeddingText(&setting))
	return nil
}
