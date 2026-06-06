package store

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jj/novelist/internal/ai"
	"github.com/jj/novelist/internal/model"
)

// FormatVector formats a []float32 as a PostgreSQL vector literal string.
func FormatVector(v []float32) string {
	s := "["
	for i, f := range v {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%f", f)
	}
	s += "]"
	return s
}

func setEmbedding(ctx context.Context, table string, id uuid.UUID, text string) {
	if ai.EmbeddingMgr == nil || text == "" {
		return
	}
	embedding, err := ai.EmbeddingMgr.GenerateEmbedding(ctx, text)
	if err != nil {
		log.Printf("embedding generation failed for %s %s: %v", table, id, err)
		return
	}
	if err := DB.Exec(
		fmt.Sprintf("UPDATE %s SET embedding = ?::vector WHERE id = ?", table),
		FormatVector(embedding), id,
	).Error; err != nil {
		log.Printf("embedding storage failed for %s %s: %v", table, id, err)
	}
}

func characterEmbeddingText(c *model.Character) string {
	return fmt.Sprintf("%s %s %s %s %s", c.Name, c.Role, c.Personality, c.Background, c.Appearance)
}

func worldSettingEmbeddingText(s *model.WorldSetting) string {
	return fmt.Sprintf("%s %s", s.Category, s.Content)
}

func outlineEmbeddingText(o *model.Outline) string {
	return fmt.Sprintf("第%d章 %s", o.ChapterNum, o.Summary)
}

func chapterEmbeddingText(c *model.Chapter) string {
	return fmt.Sprintf("%s %s", c.Title, c.Content)
}
