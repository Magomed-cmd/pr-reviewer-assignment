package e2e

import (
	"net/http"
	"testing"

	"pr-reviewer-assignment/internal/dto"
	helpers "pr-reviewer-assignment/tests/shared"

	"github.com/stretchr/testify/require"
)

const (
	e2eTeamName = "e2e-team"
	e2eAuthorID = "e2e-author"
	e2eTeamSize = 4
	e2eReviewerSlots = 2
)

func TestE2E_FullFlow(t *testing.T) {
	resetTables(t)

	members := helpers.NewTeamMembersBuilder().
		With(e2eAuthorID, "Author", true).
		With("e2e-reviewer-1", "Reviewer A", true).
		With("e2e-reviewer-2", "Reviewer B", true).
		With("e2e-reviewer-3", "Reviewer C", true).
		Build()

	team := testSuite.CreateTeam(t, e2eTeamName, members)
	require.Equal(t, e2eTeamName, team.TeamName)
	require.Len(t, team.Members, e2eTeamSize)

	resp := testSuite.PerformRequest(t, http.MethodGet, "/team/get?team_name="+e2eTeamName, nil)
	testSuite.ExpectStatus(t, resp, http.StatusOK)

	var fetched dto.TeamDTO
	testSuite.DecodeBody(t, resp, &fetched)
	require.Equal(t, e2eTeamName, fetched.TeamName)
	require.Len(t, fetched.Members, e2eTeamSize)

	pr := testSuite.CreatePullRequest(t, "E2E-PR-1", "Add e2e tests", e2eAuthorID)
	require.Equal(t, "E2E-PR-1", pr.PullRequestID)
	require.Len(t, pr.AssignedReviewers, e2eReviewerSlots)

	targetReviewer := pr.AssignedReviewers[0]

	resp = testSuite.PerformRequest(t, http.MethodGet, "/users/getReview?user_id="+targetReviewer, nil)
	testSuite.ExpectStatus(t, resp, http.StatusOK)

	var review dto.ListPullRequestsResponse
	testSuite.DecodeBody(t, resp, &review)
	require.Equal(t, targetReviewer, review.UserID)
	require.NotEmpty(t, review.PullRequests)

	resp = testSuite.PerformRequest(t, http.MethodPost, "/users/setIsActive", map[string]any{
		"user_id":   targetReviewer,
		"is_active": false,
	})
	testSuite.ExpectStatus(t, resp, http.StatusOK)

	var updated helpers.UserResponse
	testSuite.DecodeBody(t, resp, &updated)
	require.False(t, updated.User.IsActive)

	resp = testSuite.PerformRequest(t, http.MethodPost, "/pullRequest/reassign", map[string]any{
		"pull_request_id": pr.PullRequestID,
		"old_user_id":     targetReviewer,
	})
	testSuite.ExpectStatus(t, resp, http.StatusOK)

	var reassign helpers.ReassignResponse
	testSuite.DecodeBody(t, resp, &reassign)
	require.Equal(t, pr.PullRequestID, reassign.PR.PullRequestID)
	require.NotContains(t, reassign.PR.AssignedReviewers, targetReviewer)
	require.NotEmpty(t, reassign.ReplacedBy)
}
