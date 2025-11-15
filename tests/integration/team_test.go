package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pr-reviewer-assignment/internal/dto"
	"pr-reviewer-assignment/tests/integration/helpers"

	"github.com/stretchr/testify/require"
)

func TestTeamEndpoints_CreateAndGet(t *testing.T) {
	resetTables(t)

	members := helpers.NewTeamMembersBuilder().
		With("user-1", "Alice", true).
		With("user-2", "Bob", true).
		With("user-3", "Charlie", false).
		With("user-2", "Duplicated", true).
		Build()

	created := testSuite.CreateTeam(t, testTeamCore, members)
	require.Equal(t, testTeamCore, created.TeamName)
	require.Len(t, created.Members, 3)

	resp := testSuite.PerformRequest(t, http.MethodGet, "/team/get?team_name="+testTeamCore, nil)
	require.Equal(t, http.StatusOK, resp.Code)

	var fetched dto.TeamDTO
	testSuite.DecodeBody(t, resp, &fetched)

	require.Equal(t, testTeamCore, fetched.TeamName)
	require.Len(t, fetched.Members, 3)
	require.Contains(t, fetched.Members, dto.TeamMemberDTO{UserID: "user-3", Username: "Charlie", IsActive: false})
}

func TestTeamEndpoints_CreateWithoutMembers(t *testing.T) {
	resetTables(t)

	created := testSuite.CreateTeam(t, "empty-team", nil)
	require.Equal(t, "empty-team", created.TeamName)
	require.Len(t, created.Members, 0)
}

func TestTeamEndpoints_CreateValidationErrors(t *testing.T) {
	resetTables(t)

	cases := []struct {
		name       string
		setup      func(t *testing.T)
		payload    any
		rawBody    string
		wantStatus int
		wantCode   string
	}{
		{
			name: "missing team name",
			payload: map[string]any{
				"team_name": " ",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name: "duplicate team",
			setup: func(t *testing.T) {
				testSuite.CreateTeam(t, "dup-team", nil)
			},
			payload: map[string]any{
				"team_name": "dup-team",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "TEAM_EXISTS",
		},
		{
			name:       "invalid json",
			rawBody:    "{invalid json",
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
				resp = testSuite.PerformRawRequest(t, http.MethodPost, "/team/add", tc.rawBody, "application/json")
			} else {
				resp = testSuite.PerformRequest(t, http.MethodPost, "/team/add", tc.payload)
			}

			testSuite.ExpectError(t, resp, tc.wantStatus, tc.wantCode)
		})
	}
}

func TestTeamEndpoints_GetValidationErrors(t *testing.T) {
	resetTables(t)

	cases := []struct {
		name       string
		path       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing query",
			path:       "/team/get",
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "team not found",
			path:       "/team/get?team_name=unknown",
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
