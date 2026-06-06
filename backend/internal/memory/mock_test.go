package memory

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
	"sort"
)

// mockQuerier is an in-memory store for testing without SQLite/CGO.
type mockQuerier struct {
	projects   map[uuid.UUID]*model.Project
	worldSets  map[uuid.UUID][]model.WorldSetting
	characters map[uuid.UUID][]model.Character
	outlines   map[uuid.UUID][]model.Outline
	chapters   map[uuid.UUID][]model.Chapter
}

func newMockQuerier() *mockQuerier {
	return &mockQuerier{
		projects:   make(map[uuid.UUID]*model.Project),
		worldSets:  make(map[uuid.UUID][]model.WorldSetting),
		characters: make(map[uuid.UUID][]model.Character),
		outlines:   make(map[uuid.UUID][]model.Outline),
		chapters:   make(map[uuid.UUID][]model.Chapter),
	}
}

func (m *mockQuerier) GetProject(id uuid.UUID) (*model.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, fmt.Errorf("project not found")
	}
	return p, nil
}

func (m *mockQuerier) GetWorldSettings(projectID uuid.UUID) ([]model.WorldSetting, error) {
	return m.worldSets[projectID], nil
}

func (m *mockQuerier) GetCharacters(projectID uuid.UUID) ([]model.Character, error) {
	return m.characters[projectID], nil
}

func (m *mockQuerier) GetOutlines(projectID uuid.UUID) ([]model.Outline, error) {
	outlines := m.outlines[projectID]
	sort.Slice(outlines, func(i, j int) bool {
		if outlines[i].Act != outlines[j].Act {
			return outlines[i].Act < outlines[j].Act
		}
		return outlines[i].ChapterNum < outlines[j].ChapterNum
	})
	return outlines, nil
}

func (m *mockQuerier) GetChaptersBefore(projectID uuid.UUID, chapterNum int, limit int) ([]model.Chapter, error) {
	var result []model.Chapter
	for _, ch := range m.chapters[projectID] {
		if ch.ChapterNum < chapterNum {
			result = append(result, ch)
		}
	}
	// Sort DESC by chapter_num (matches GORM Order("chapter_num DESC"))
	sort.Slice(result, func(i, j int) bool {
		return result[i].ChapterNum > result[j].ChapterNum
	})
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockQuerier) GetChaptersWithEmbedding(projectID uuid.UUID, limit int) ([]model.Character, []model.WorldSetting, []model.Outline, []model.Chapter, error) {
	return nil, nil, nil, nil, nil
}

// Helper methods to seed test data
func (m *mockQuerier) addProject(p model.Project) {
	m.projects[p.ID] = &p
}

func (m *mockQuerier) addWorldSetting(s model.WorldSetting) {
	m.worldSets[s.ProjectID] = append(m.worldSets[s.ProjectID], s)
}

func (m *mockQuerier) addCharacter(c model.Character) {
	m.characters[c.ProjectID] = append(m.characters[c.ProjectID], c)
}

func (m *mockQuerier) addOutline(o model.Outline) {
	m.outlines[o.ProjectID] = append(m.outlines[o.ProjectID], o)
}

func (m *mockQuerier) addChapter(ch model.Chapter) {
	m.chapters[ch.ProjectID] = append(m.chapters[ch.ProjectID], ch)
}
