package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
)

func CreateVolume(ctx context.Context, volume *model.Volume) error {
	return DB.Create(volume).Error
}

func GetVolume(ctx context.Context, id uuid.UUID) (*model.Volume, error) {
	var volume model.Volume
	err := DB.Where("id = ?", id).First(&volume).Error
	return &volume, err
}

func GetVolumesByProject(ctx context.Context, projectID uuid.UUID) ([]model.Volume, error) {
	var volumes []model.Volume
	err := DB.Where("project_id = ?", projectID).Order("volume_num").Find(&volumes).Error
	return volumes, err
}

func UpdateVolume(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return DB.Model(&model.Volume{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteVolume(ctx context.Context, id uuid.UUID) error {
	return DB.Where("id = ?", id).Delete(&model.Volume{}).Error
}

// GetOrCreateDefaultVolume returns volume_num=1 for the project, creating it if needed.
func GetOrCreateDefaultVolume(ctx context.Context, projectID uuid.UUID) (*model.Volume, error) {
	var volume model.Volume
	err := DB.Where("project_id = ? AND volume_num = 1", projectID).First(&volume).Error
	if err == nil {
		return &volume, nil
	}
	volume = model.Volume{
		ProjectID: projectID,
		VolumeNum: 1,
		Title:     "第一篇",
	}
	if err := DB.Create(&volume).Error; err != nil {
		return nil, err
	}
	return &volume, nil
}

// GetLatestVolume returns the volume with the highest volume_num for the project.
func GetLatestVolume(ctx context.Context, projectID uuid.UUID) (*model.Volume, error) {
	var volume model.Volume
	err := DB.Where("project_id = ?", projectID).Order("volume_num DESC").First(&volume).Error
	return &volume, err
}
