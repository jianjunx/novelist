package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
)

func GetSettings(ctx context.Context, userID uuid.UUID) (*model.Setting, error) {
	var settings model.Setting
	err := DB.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}
