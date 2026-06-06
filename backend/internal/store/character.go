package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
)

func CreateCharacter(ctx context.Context, character *model.Character) error {
	if err := DB.Create(character).Error; err != nil {
		return err
	}
	setEmbedding(ctx, "characters", character.ID, characterEmbeddingText(character))
	return nil
}

func UpdateCharacter(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if err := DB.Model(&model.Character{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	var character model.Character
	if err := DB.Where("id = ?", id).First(&character).Error; err != nil {
		return err
	}
	setEmbedding(ctx, "characters", id, characterEmbeddingText(&character))
	return nil
}
