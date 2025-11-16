package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	helpers "pr-reviewer-assignment/tests/shared"

	"github.com/stretchr/testify/require"
)

func TestPullRequestEndpoints_CreateAndReassign(t *testing.T) {
	resetTables(t)

	members := helpers.NewTeamMembersBuilder().
		With(testAuthorID, "Author", true).
		With("reviewer-1", "Bob", true).
		With("reviewer-2", "Charlie", true).
		With("reviewer-3", "Dana", true).
		Build()
	testSuite.CreateTeam(t, testTeamPlatform, members)

	const prID = "PR-1001"
	pr := testSuite.CreatePullRequest(t, prID, "Add metrics", testAuthorID)
	require.Equal(t, prID, pr.PullRequestID)
	require.Equal(t, "OPEN", pr.Status)
	require.Len(t, pr.AssignedReviewers, 2)
	require.NotContains(t, pr.AssignedReviewers, testAuthorID)
	require.False(t, pr.NeedMoreReviewers)

	oldReviewer := pr.AssignedReviewers[0]
	resp := testSuite.PerformRequest(t, http.MethodPost, "/pullRequest/reassign", map[string]any{
		"pull_request_id": pr.PullRequestID,
		"old_user_id":     oldReviewer,
	})
	require.Equal(t, http.StatusOK, resp.Code)

	var reassign helpers.ReassignResponse
	testSuite.DecodeBody(t, resp, &reassign)

	require.NotNil(t, reassign.PR)
	require.Equal(t, pr.PullRequestID, reassign.PR.PullRequestID)
	require.Len(t, reassign.PR.AssignedReviewers, 2)
	require.NotContains(t, reassign.PR.AssignedReviewers, oldReviewer)
	require.NotEmpty(t, reassign.ReplacedBy)
	require.Contains(t, reassign.PR.AssignedReviewers, reassign.ReplacedBy)
}

func TestPullRequestEndpoints_CreateValidationErrors(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(t *testing.T)
		payload    any
		rawBody    string
		wantStatus int
		wantCode   string
	}{
		{
			name: "author not found",
			payload: map[string]any{
				"pull_request_id":   "PR-404",
				"pull_request_name": "Ghost",
				"author_id":         "ghost",
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name: "no reviewers available",
			setup: func(t *testing.T) {
				testSuite.CreateTeam(t, "solo", helpers.NewTeamMembersBuilder().
					With(testAuthorID, "Author", true).
					Build())
			},
			payload: map[string]any{
				"pull_request_id":   "PR-100",
				"pull_request_name": "Solo",
				"author_id":         testAuthorID,
			},
			wantStatus: http.StatusConflict,
			wantCode:   "NO_CANDIDATE",
		},
		{
			name: "invalid payload",
			payload: map[string]any{
				"pull_request_id": "",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid json",
			rawBody:    "{",
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetTables(t)
			if tc.setup != nil {
				tc.setup(t)
			}

			var resp *httptest.ResponseRecorder
			if tc.rawBody != "" {
				resp = testSuite.PerformRawRequest(t, http.MethodPost, "/pullRequest/create", tc.rawBody, "application/json")
			} else {
				resp = testSuite.PerformRequest(t, http.MethodPost, "/pullRequest/create", tc.payload)
			}
			testSuite.ExpectError(t, resp, tc.wantStatus, tc.wantCode)
		})
	}
}

func TestPullRequestEndpoints_ReassignValidationErrors(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(t *testing.T) map[string]any
		payload    map[string]any
		rawBody    string
		wantStatus int
		wantCode   string
	}{
		{
			name: "reviewer not assigned",
			setup: func(t *testing.T) map[string]any {
				members := helpers.NewTeamMembersBuilder().
					With(testAuthorID, "Author", true).
					With("reviewer-1", "Bob", true).
					With("reviewer-2", "Charlie", true).
					With("reviewer-3", "Dana", true).
					With("reviewer-4", "Eve", true).
					Build()
				testSuite.CreateTeam(t, testTeamCore, members)

				pr := testSuite.CreatePullRequest(t, "PR-300", "Edge", testAuthorID)
				var notAssigned string
				for _, candidate := range []string{"reviewer-1", "reviewer-2", "reviewer-3", "reviewer-4"} {
					if !contains(pr.AssignedReviewers, candidate) {
						notAssigned = candidate
						break
					}
				}

				require.NotEmpty(t, notAssigned, "expected at least one unassigned reviewer")

				return map[string]any{
					"pull_request_id": pr.PullRequestID,
					"old_user_id":     notAssigned,
				}
			},
			wantStatus: http.StatusConflict,
			wantCode:   "NOT_ASSIGNED",
		},
		{
			name: "merged pull request",
			setup: func(t *testing.T) map[string]any {
				members := helpers.NewTeamMembersBuilder().
					With(testAuthorID, "Author", true).
					With("reviewer-1", "Bob", true).
					With("reviewer-2", "Charlie", true).
					With("reviewer-3", "Dana", true).
					Build()
				testSuite.CreateTeam(t, testTeamCore, members)

				pr := testSuite.CreatePullRequest(t, "PR-301", "Merge me", testAuthorID)
				testSuite.MergePullRequest(t, pr.PullRequestID)
				return map[string]any{
					"pull_request_id": pr.PullRequestID,
					"old_user_id":     pr.AssignedReviewers[0],
				}
			},
			wantStatus: http.StatusConflict,
			wantCode:   "PR_MERGED",
		},
		{
			name: "invalid payload",
			payload: map[string]any{
				"pull_request_id": "",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid json",
			rawBody:    "{",
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetTables(t)

			var resp *httptest.ResponseRecorder
			if tc.rawBody != "" {
				resp = testSuite.PerformRawRequest(t, http.MethodPost, "/pullRequest/reassign", tc.rawBody, "application/json")
			} else {
				body := tc.payload
				if tc.setup != nil {
					body = tc.setup(t)
				}
				resp = testSuite.PerformRequest(t, http.MethodPost, "/pullRequest/reassign", body)
			}
			testSuite.ExpectError(t, resp, tc.wantStatus, tc.wantCode)
		})
	}
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
