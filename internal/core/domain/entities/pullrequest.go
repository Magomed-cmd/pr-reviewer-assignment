package entities

import (
	"strings"
	"time"

	domainErrors "pr-reviewer-assignment/internal/core/domain/errors"
	"pr-reviewer-assignment/internal/core/domain/types"
)

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            types.PRStatus
	AssignedReviewers []string
	NeedMoreReviewers bool
	CreatedAt         time.Time
	MergedAt          *time.Time
}

func NewPullRequest(id, name, authorID string, createdAt time.Time) *PullRequest {
	return &PullRequest{
		ID:                id,
		Name:              name,
		AuthorID:          authorID,
		Status:            types.PRStatusOpen,
		AssignedReviewers: make([]string, 0, 2),
		NeedMoreReviewers: true,
		CreatedAt:         createdAt,
	}
}

func (p *PullRequest) AssignReviewers(reviewers []string) error {
	if p.Status == types.PRStatusMerged {
		return domainErrors.PRMerged(p.ID)
	}

	unique := make(map[string]struct{}, len(reviewers))
	assigned := make([]string, 0, min(2, len(reviewers)))

	for _, reviewer := range reviewers {
		reviewer = strings.TrimSpace(reviewer)

		if reviewer == "" {
			continue
		}

		if reviewer == p.AuthorID {
			continue
		}

		if _, exists := unique[reviewer]; exists {
			continue
		}

		unique[reviewer] = struct{}{}
		assigned = append(assigned, reviewer)

		if len(assigned) == 2 {
			break
		}
	}

	p.AssignedReviewers = assigned
	p.updateNeedMoreReviewers()
	return nil
}

func (p *PullRequest) ReplaceReviewer(oldReviewer, newReviewer string) (string, error) {
	if p.Status == types.PRStatusMerged {
		return "", domainErrors.PRMerged(p.ID)
	}

	index := p.reviewerIndex(oldReviewer)
	if index == -1 {
		return "", domainErrors.NotAssigned(oldReviewer, p.ID)
	}

	if newReviewer == "" {
		p.AssignedReviewers = removeIndex(p.AssignedReviewers, index)
		p.updateNeedMoreReviewers()
		return "", nil
	}

	if p.HasReviewer(newReviewer) {
		p.AssignedReviewers = removeIndex(p.AssignedReviewers, index)
		p.updateNeedMoreReviewers()
		return "", nil
	}

	p.AssignedReviewers[index] = newReviewer
	p.updateNeedMoreReviewers()
	return newReviewer, nil
}

func (p *PullRequest) Merge(at time.Time) {
	if p.Status == types.PRStatusMerged {
		return
	}

	p.Status = types.PRStatusMerged
	p.MergedAt = &at
}

func (p *PullRequest) HasReviewer(userID string) bool {
	return p.reviewerIndex(userID) != -1
}

func (p *PullRequest) reviewerIndex(userID string) int {
	for i, reviewer := range p.AssignedReviewers {
		if reviewer == userID {
			return i
		}
	}

	return -1
}

func (p *PullRequest) SetReviewers(reviewers []string) {
	p.AssignedReviewers = reviewers
	p.updateNeedMoreReviewers()
}

func (p *PullRequest) updateNeedMoreReviewers() {
	p.NeedMoreReviewers = len(p.AssignedReviewers) < 2
}

func removeIndex(items []string, index int) []string {
	copy(items[index:], items[index+1:])
	return items[:len(items)-1]
}
