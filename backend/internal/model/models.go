package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// GenerateShortID creates an 8-character random hex string
func GenerateShortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Project struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ShortID     string    `gorm:"uniqueIndex;not null" json:"short_id"`
	UserID      uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Title       string    `gorm:"not null" json:"title"`
	Genre       string    `json:"genre"`
	Description string    `json:"description"`
	StyleGuide  string    `json:"style_guide"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Volume struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID `gorm:"type:uuid;index" json:"project_id"`
	VolumeNum   int       `gorm:"not null" json:"volume_num"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Summary     string    `json:"summary"`
	Status      string    `gorm:"default:draft" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Character struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID     uuid.UUID `gorm:"type:uuid;index" json:"project_id"`
	Name          string    `gorm:"not null" json:"name"`
	Role          string    `json:"role"`
	Personality   string    `json:"personality"`
	Background    string    `json:"background"`
	Appearance    string    `json:"appearance"`
	Relationships datatypes.JSON  `gorm:"type:jsonb" json:"relationships"`
	Embedding     []float32 `gorm:"type:vector(1536)" json:"-"`
	CreatedAt     time.Time `json:"created_at"`
}

type WorldSetting struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;index" json:"project_id"`
	Category  string    `gorm:"not null" json:"category"`
	Content   string    `gorm:"not null" json:"content"`
	Embedding []float32 `gorm:"type:vector(1536)" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Outline struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID  uuid.UUID  `gorm:"type:uuid;index" json:"project_id"`
	VolumeID   *uuid.UUID `gorm:"type:uuid;index" json:"volume_id"`
	Act        int        `json:"act"`
	ChapterNum int       `json:"chapter_num"`
	Summary    string    `json:"summary"`
	KeyEvents  datatypes.JSON  `gorm:"type:jsonb" json:"key_events"`
	Status     string    `gorm:"default:draft" json:"status"`
	Embedding  []float32 `gorm:"type:vector(1536)" json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

type Chapter struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID  uuid.UUID  `gorm:"type:uuid;index" json:"project_id"`
	OutlineID  *uuid.UUID `gorm:"type:uuid" json:"outline_id"`
	ChapterNum int        `gorm:"not null" json:"chapter_num"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	WordCount  int        `gorm:"default:0" json:"word_count"`
	Status     string     `gorm:"default:draft" json:"status"`
	Embedding  []float32  `gorm:"type:vector(1536)" json:"-"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Discussion struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChapterID      uuid.UUID `gorm:"type:uuid;index" json:"chapter_id"`
	RoundNum       int       `gorm:"not null" json:"round_num"`
	AgentRole      string    `gorm:"not null" json:"agent_role"`
	Content        string    `gorm:"not null" json:"content"`
	SuggestionType string    `json:"suggestion_type"`
	Priority       int       `gorm:"default:0" json:"priority"`
	CreatedAt      time.Time `json:"created_at"`
}

type Conversation struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;index" json:"project_id"`
	Role      string    `gorm:"not null" json:"role"`
	Content   string    `gorm:"not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Setting struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"user_id"`
	DefaultModel     string    `gorm:"default:deepseek-chat" json:"default_model"`
	DeepSeekKey      string    `json:"deepseek_key"`
	ClaudeKey        string    `json:"claude_key"`
	OpenAIKey        string    `json:"openai_key"`
	LocalModelURL    string    `json:"local_model_url"`
	AgentModelConfig datatypes.JSON  `gorm:"type:jsonb" json:"agent_model_config"`
	DefaultWordCount int       `gorm:"default:800" json:"default_word_count"`
	DiscussionRounds int       `gorm:"default:1" json:"discussion_rounds"`
	LanguageStyle    string    `gorm:"default:现代中文" json:"language_style"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
