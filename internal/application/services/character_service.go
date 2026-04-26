package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/mappers"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"gorm.io/gorm"
)

type CharacterService struct {
	repo repositories.CharacterRepository
}

func NewCharacterService(repo repositories.CharacterRepository) *CharacterService {
	return &CharacterService{repo: repo}
}

func (s *CharacterService) CreateCharacter(req dtos.CreateCharacterRequest) (*dtos.CharacterResponse, error) {
	skills := req.Skills.ToModel()
	if err := skills.Validate(); err != nil {
		return nil, fmt.Errorf("invalid skills: %w", err)
	}

	character := &models.Character{
		FullName:    strings.TrimSpace(req.FullName),
		ShortName:   strings.TrimSpace(req.ShortName),
		Description: strings.TrimSpace(req.Description),
		Occupation:  strings.TrimSpace(req.Occupation),
		Location:    strings.TrimSpace(req.Location),
		Portrait:    req.Portrait,
		Skills:      skills,
	}
	if err := s.repo.Create(character); err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	resp := mappers.ToCharacterResponse(character)
	return &resp, nil
}

func (s *CharacterService) GetCharacter(id string) (*dtos.CharacterResponse, error) {
	if !isValidUUID(id) {
		return nil, fmt.Errorf("invalid UUID format")
	}
	character, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch character: %w", err)
	}
	if character == nil {
		return nil, nil
	}
	resp := mappers.ToCharacterResponse(character)
	return &resp, nil
}

func (s *CharacterService) ListCharacters(search string) ([]dtos.CharacterResponse, error) {
	rows, err := s.repo.FindAll(models.CharacterFilters{Search: search})
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}
	out := make([]dtos.CharacterResponse, len(rows))
	for i, c := range rows {
		out[i] = mappers.ToCharacterResponse(c)
	}
	return out, nil
}

func (s *CharacterService) UpdateCharacter(id string, req dtos.UpdateCharacterRequest) (*dtos.CharacterResponse, error) {
	if !isValidUUID(id) {
		return nil, fmt.Errorf("invalid UUID format")
	}
	current, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch character: %w", err)
	}
	if current == nil {
		return nil, nil
	}

	if req.FullName != nil {
		current.FullName = strings.TrimSpace(*req.FullName)
	}
	if req.ShortName != nil {
		current.ShortName = strings.TrimSpace(*req.ShortName)
	}
	if req.Description != nil {
		current.Description = strings.TrimSpace(*req.Description)
	}
	if req.Occupation != nil {
		current.Occupation = strings.TrimSpace(*req.Occupation)
	}
	if req.Location != nil {
		current.Location = strings.TrimSpace(*req.Location)
	}
	if req.Portrait != nil {
		current.Portrait = *req.Portrait
	}
	if req.Skills != nil {
		skills := req.Skills.ToModel()
		if err := skills.Validate(); err != nil {
			return nil, fmt.Errorf("invalid skills: %w", err)
		}
		current.Skills = skills
	}

	if err := s.repo.Update(id, current); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	updated, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to refetch character: %w", err)
	}
	resp := mappers.ToCharacterResponse(updated)
	return &resp, nil
}

func (s *CharacterService) DeleteCharacter(id string) error {
	if !isValidUUID(id) {
		return fmt.Errorf("invalid UUID format")
	}
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete character: %w", err)
	}
	return nil
}
