package dtos

import "github.com/davidrdsilva/blog-api/internal/domain/models"

type CharacterSkillsRequest struct {
	Melee      int `json:"melee" binding:"min=0,max=100"`
	Guns       int `json:"guns" binding:"min=0,max=100"`
	Stealth    int `json:"stealth" binding:"min=0,max=100"`
	Persuasion int `json:"persuasion" binding:"min=0,max=100"`
	Intellect  int `json:"intellect" binding:"min=0,max=100"`
	Endurance  int `json:"endurance" binding:"min=0,max=100"`
}

func (s CharacterSkillsRequest) ToModel() models.CharacterSkills {
	return models.CharacterSkills{
		Melee:      s.Melee,
		Guns:       s.Guns,
		Stealth:    s.Stealth,
		Persuasion: s.Persuasion,
		Intellect:  s.Intellect,
		Endurance:  s.Endurance,
	}
}

type CreateCharacterRequest struct {
	FullName    string                 `json:"full_name" binding:"required,min=1,max=120"`
	ShortName   string                 `json:"short_name" binding:"required,min=1,max=60"`
	Description string                 `json:"description" binding:"required,min=1"`
	Occupation  string                 `json:"occupation" binding:"required,min=1,max=120"`
	Location    string                 `json:"location" binding:"required,min=1,max=160"`
	Portrait    string                 `json:"portrait" binding:"required,url"`
	Skills      CharacterSkillsRequest `json:"skills" binding:"required"`
}

type UpdateCharacterRequest struct {
	FullName    *string                 `json:"full_name" binding:"omitempty,min=1,max=120"`
	ShortName   *string                 `json:"short_name" binding:"omitempty,min=1,max=60"`
	Description *string                 `json:"description" binding:"omitempty,min=1"`
	Occupation  *string                 `json:"occupation" binding:"omitempty,min=1,max=120"`
	Location    *string                 `json:"location" binding:"omitempty,min=1,max=160"`
	Portrait    *string                 `json:"portrait" binding:"omitempty,url"`
	Skills      *CharacterSkillsRequest `json:"skills"`
}

type CharacterResponse struct {
	ID          string                 `json:"id"`
	FullName    string                 `json:"full_name"`
	ShortName   string                 `json:"short_name"`
	Description string                 `json:"description"`
	Occupation  string                 `json:"occupation"`
	Location    string                 `json:"location"`
	Portrait    string                 `json:"portrait"`
	Skills      models.CharacterSkills `json:"skills"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
}

type CharacterListResponse struct {
	Data []CharacterResponse `json:"data"`
}
