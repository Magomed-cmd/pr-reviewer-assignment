package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pr-reviewer-assignment/internal/dto"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type Suite struct {
	Router *gin.Engine
	Pool   *pgxpool.Pool
}

func NewSuite(router *gin.Engine, pool *pgxpool.Pool) *Suite {
	return &Suite{
		Router: router,
		Pool:   pool,
	}
}

func (s *Suite) ResetTables(t testing.TB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.Pool.Exec(ctx, `TRUNCATE TABLE pr_reviewers, pull_requests, users, teams RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func (s *Suite) PerformRequest(t testing.TB, method, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(data)
	} else {
		body = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, path, body)
	require.NoError(t, err)

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	s.Router.ServeHTTP(rec, req)
	return rec
}

func (s *Suite) PerformRawRequest(t testing.TB, method, path, body, contentType string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequest(method, path, strings.NewReader(body))
	require.NoError(t, err)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	rec := httptest.NewRecorder()
	s.Router.ServeHTTP(rec, req)
	return rec
}

func (s *Suite) DecodeBody(t testing.TB, rec *httptest.ResponseRecorder, target any) {
	t.Helper()

	err := json.NewDecoder(rec.Body).Decode(target)
	require.NoError(t, err)
}

func (s *Suite) ExpectError(t testing.TB, rec *httptest.ResponseRecorder, status int, code string) ErrorResponse {
	t.Helper()

	require.Equal(t, status, rec.Code)

	var resp ErrorResponse
	s.DecodeBody(t, rec, &resp)
	require.Equal(t, code, resp.Error.Code)
	return resp
}

func (s *Suite) CreateTeam(t testing.TB, teamName string, members []map[string]any) *dto.TeamDTO {
	t.Helper()

	resp := s.PerformRequest(t, http.MethodPost, "/team/add", map[string]any{
		"team_name": teamName,
		"members":   members,
	})
	require.Equal(t, http.StatusCreated, resp.Code)

	var parsed CreateTeamResponse
	s.DecodeBody(t, resp, &parsed)
	return parsed.Team
}

func (s *Suite) CreatePullRequest(t testing.TB, prID, name, authorID string) *dto.PullRequestDTO {
	t.Helper()

	resp := s.PerformRequest(t, http.MethodPost, "/pullRequest/create", map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": name,
		"author_id":         authorID,
	})
	require.Equal(t, http.StatusCreated, resp.Code)

	var parsed PullRequestResponse
	s.DecodeBody(t, resp, &parsed)
	return parsed.PR
}

func (s *Suite) MergePullRequest(t testing.TB, prID string) *dto.PullRequestDTO {
	t.Helper()

	resp := s.PerformRequest(t, http.MethodPost, "/pullRequest/merge", map[string]any{
		"pull_request_id": prID,
	})
	require.Equal(t, http.StatusOK, resp.Code)

	var parsed PullRequestResponse
	s.DecodeBody(t, resp, &parsed)
	return parsed.PR
}

type CreateTeamResponse struct {
	Team *dto.TeamDTO `json:"team"`
}

type PullRequestResponse struct {
	PR *dto.PullRequestDTO `json:"pr"`
}

type UserResponse struct {
	User *dto.UserDTO `json:"user"`
}

type ReassignResponse struct {
	PR         *dto.PullRequestDTO `json:"pr"`
	ReplacedBy string              `json:"replaced_by"`
}

type ErrorResponse struct {
	Error dto.ErrorBody `json:"error"`
}

type TeamMembersBuilder struct {
	members []map[string]any
}

func NewTeamMembersBuilder() *TeamMembersBuilder {
	return &TeamMembersBuilder{members: make([]map[string]any, 0)}
}

func (b *TeamMembersBuilder) With(id, username string, active bool) *TeamMembersBuilder {
	b.members = append(b.members, map[string]any{
		"user_id":   id,
		"username":  username,
		"is_active": active,
	})
	return b
}

func (b *TeamMembersBuilder) Build() []map[string]any {
	out := make([]map[string]any, len(b.members))
	copy(out, b.members)
	return out
}
