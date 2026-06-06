package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
)

func CreateOutline(ctx context.Context, outline *model.Outline) error {
	if err := DB.Create(outline).Error; err != nil {
		return err
	}
	setEmbedding(ctx, "outlines", outline.ID, outlineEmbeddingText(outline))
	return nil
}

func UpdateOutline(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if err := DB.Model(&model.Outline{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	var outline model.Outline
	if err := DB.Where("id = ?", id).First(&outline).Error; err != nil {
		return err
	}
	setEmbedding(ctx, "outlines", id, outlineEmbeddingText(&outline))
	return nil
}
