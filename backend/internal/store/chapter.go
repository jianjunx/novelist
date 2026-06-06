package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
)

func CreateChapter(ctx context.Context, chapter *model.Chapter) error {
	if err := DB.Create(chapter).Error; err != nil {
		return err
	}
	setEmbedding(ctx, "chapters", chapter.ID, chapterEmbeddingText(chapter))
	return nil
}

func UpdateChapter(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if err := DB.Model(&model.Chapter{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	var chapter model.Chapter
	if err := DB.Where("id = ?", id).First(&chapter).Error; err != nil {
		return err
	}
	setEmbedding(ctx, "chapters", id, chapterEmbeddingText(&chapter))
	return nil
}
