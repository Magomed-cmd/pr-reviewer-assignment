package mappers

import (
	"time"

	"pr-reviewer-assignment/internal/core/domain/entities"
	"pr-reviewer-assignment/internal/dto"
)

func TeamToDTO(team *entities.Team) *dto.TeamDTO {
	if team == nil {
		return nil
	}

	members := make([]dto.TeamMemberDTO, 0, len(team.Members))
	for _, member := range team.Members {
		if member == nil {
			continue
		}

		members = append(members, dto.TeamMemberDTO{
			UserID:   member.ID,
			Username: member.Username,
			IsActive: member.IsActive,
		})
	}

	return &dto.TeamDTO{
		TeamName: team.Name,
		Members:  members,
	}
}

func TeamMembersFromDTO(teamName string, members []dto.TeamMemberDTO) []*entities.User {
	if len(members) == 0 {
		return nil
	}

	result := make([]*entities.User, 0, len(members))
	for _, member := range members {
		user := entities.NewUser(member.UserID, member.Username, teamName, member.IsActive, time.Time{}, time.Time{})
		result = append(result, user)
	}

	return result
}
