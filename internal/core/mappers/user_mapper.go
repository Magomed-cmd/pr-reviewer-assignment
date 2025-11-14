package mappers

import (
	"pr-reviewer-assignment/internal/core/domain/entities"
	"pr-reviewer-assignment/internal/dto"
)

func UserToDTO(user *entities.User) *dto.UserDTO {
	if user == nil {
		return nil
	}

	return &dto.UserDTO{
		UserID:   user.ID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

func UsersToDTO(users []*entities.User) []*dto.UserDTO {
	if len(users) == 0 {
		return nil
	}

	result := make([]*dto.UserDTO, 0, len(users))
	for _, user := range users {
		result = append(result, UserToDTO(user))
	}

	return result
}
