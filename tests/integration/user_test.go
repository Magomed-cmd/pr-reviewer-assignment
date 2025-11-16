package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pr-reviewer-assignment/internal/dto"
	helpers "pr-reviewer-assignment/tests/shared"

	"github.com/stretchr/testify/require"
)

func TestUserEndpoints_AssignmentsAndActivity(t *testing.T) {
	resetTables(t)

	members := helpers.NewTeamMembersBuilder().
		With(testAuthorID, "Author", true).
		With("reviewer-1", "Bob", true).
		With("reviewer-2", "Charlie", true).
		Build()
	testSuite.CreateTeam(t, testTeamCore, members)

	pr1 := testSuite.CreatePullRequest(t, "PR-2001", "Add logging", testAuthorID)
	require.Len(t, pr1.AssignedReviewers, 2)

	pr2 := testSuite.CreatePullRequest(t, "PR-2002", "Add tracing", testAuthorID)
	require.Len(t, pr2.AssignedReviewers, 2)

	targetReviewer := pr1.AssignedReviewers[0]

	resp := testSuite.PerformRequest(t, http.MethodGet, "/users/getReview?user_id="+targetReviewer, nil)
	require.Equal(t, http.StatusOK, resp.Code)

	var assignments dto.ListPullRequestsResponse
	testSuite.DecodeBody(t, resp, &assignments)
	require.Equal(t, targetReviewer, assignments.UserID)
	require.Len(t, assignments.PullRequests, 2)

	resp = testSuite.PerformRequest(t, http.MethodPost, "/users/setIsActive", map[string]any{
		"user_id":   targetReviewer,
		"is_active": false,
	})
	require.Equal(t, http.StatusOK, resp.Code)

	var updated helpers.UserResponse
	testSuite.DecodeBody(t, resp, &updated)
	require.NotNil(t, updated.User)
	require.False(t, updated.User.IsActive)
}

func TestUserEndpoints_GetReviewErrors(t *testing.T) {
	resetTables(t)

	cases := []struct {
		name       string
		path       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing query",
			path:       "/users/getReview",
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "user not found",
			path:       "/users/getReview?user_id=ghost",
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := testSuite.PerformRequest(t, http.MethodGet, tc.path, nil)
			testSuite.ExpectError(t, resp, tc.wantStatus, tc.wantCode)
		})
	}
}

func TestUserEndpoints_SetActivityErrors(t *testing.T) {
	resetTables(t)

	cases := []struct {
		name       string
		payload    any
		rawBody    string
		wantStatus int
		wantCode   string
	}{
		{
			name: "user not found",
			payload: map[string]any{
				"user_id":   "ghost",
				"is_active": true,
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name: "missing fields",
			payload: map[string]any{
				"user_id": "",
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
			var resp *httptest.ResponseRecorder
			if tc.rawBody != "" {
				resp = testSuite.PerformRawRequest(t, http.MethodPost, "/users/setIsActive", tc.rawBody, "application/json")
			} else {
				resp = testSuite.PerformRequest(t, http.MethodPost, "/users/setIsActive", tc.payload)
			}
			testSuite.ExpectError(t, resp, tc.wantStatus, tc.wantCode)
		})
	}
}
