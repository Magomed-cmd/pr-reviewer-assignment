package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

type IntegrationSuite struct {
	router *gin.Engine
	pool   *pgxpool.Pool
}

func NewIntegrationSuite(router *gin.Engine, pool *pgxpool.Pool) *IntegrationSuite {
	return &IntegrationSuite{
		router: router,
		pool:   pool,
	}
}

func (s *IntegrationSuite) ResetTables(t testing.TB) {
	resetTables(t, s.pool)
}

func (s *IntegrationSuite) PerformRequest(t testing.TB, method, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()

	req := newRequest(t, method, path, payload)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}

func (s *IntegrationSuite) PerformRawRequest(t testing.TB, method, path, rawBody, contentType string) *httptest.ResponseRecorder {
	t.Helper()

	req := newRawRequest(t, method, path, rawBody, contentType)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}

func (s *IntegrationSuite) DecodeBody(t testing.TB, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), target))
}

func (s *IntegrationSuite) ExpectError(t testing.TB, rec *httptest.ResponseRecorder, status int, code string) ErrorResponse {
	t.Helper()
	require.Equal(t, status, rec.Code)

	var resp ErrorResponse
	s.DecodeBody(t, rec, &resp)
	require.Equal(t, code, resp.Error.Code)
	return resp
}

func (s *IntegrationSuite) CreateTeam(t testing.TB, teamName string, members []map[string]any) *dto.TeamDTO {
	t.Helper()

	rec := s.PerformRequest(t, http.MethodPost, "/team/add", map[string]any{
		"team_name": teamName,
		"members":   members,
	})
	require.Equal(t, http.StatusCreated, rec.Code)

	var parsed CreateTeamResponse
	s.DecodeBody(t, rec, &parsed)
	return parsed.Team
}

func (s *IntegrationSuite) CreatePullRequest(t testing.TB, prID, name, authorID string) *dto.PullRequestDTO {
	t.Helper()

	rec := s.PerformRequest(t, http.MethodPost, "/pullRequest/create", map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": name,
		"author_id":         authorID,
	})
	require.Equal(t, http.StatusCreated, rec.Code)

	var parsed PullRequestResponse
	s.DecodeBody(t, rec, &parsed)
	return parsed.PR
}

func (s *IntegrationSuite) MergePullRequest(t testing.TB, prID string) *dto.PullRequestDTO {
	t.Helper()

	rec := s.PerformRequest(t, http.MethodPost, "/pullRequest/merge", map[string]any{
		"pull_request_id": prID,
	})
	require.Equal(t, http.StatusOK, rec.Code)

	var parsed PullRequestResponse
	s.DecodeBody(t, rec, &parsed)
	return parsed.PR
}

type E2ESuite struct {
	baseURL string
	client  *http.Client
	pool    *pgxpool.Pool
}

func NewE2ESuite(baseURL string, client *http.Client, pool *pgxpool.Pool) *E2ESuite {
	if client == nil {
		client = http.DefaultClient
	}

	return &E2ESuite{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
		pool:    pool,
	}
}

func (s *E2ESuite) ResetTables(t testing.TB) {
	resetTables(t, s.pool)
}

func (s *E2ESuite) PerformRequest(t testing.TB, method, path string, payload any) *http.Response {
	t.Helper()

	req := newRequest(t, method, s.baseURL+path, payload)
	resp, err := s.client.Do(req)
	require.NoError(t, err)
	return resp
}

func (s *E2ESuite) PerformRawRequest(t testing.TB, method, path, rawBody, contentType string) *http.Response {
	t.Helper()

	req := newRawRequest(t, method, s.baseURL+path, rawBody, contentType)
	resp, err := s.client.Do(req)
	require.NoError(t, err)
	return resp
}

func (s *E2ESuite) DecodeBody(t testing.TB, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(target))
}

func (s *E2ESuite) ExpectStatus(t testing.TB, resp *http.Response, status int) {
	t.Helper()
	if resp.StatusCode == status {
		return
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
	require.Equal(t, status, resp.StatusCode, "unexpected status: %s", string(body))
}

func (s *E2ESuite) ExpectError(t testing.TB, resp *http.Response, status int, code string) ErrorResponse {
	t.Helper()
	s.ExpectStatus(t, resp, status)
	var parsed ErrorResponse
	s.DecodeBody(t, resp, &parsed)
	require.Equal(t, code, parsed.Error.Code)
	return parsed
}

func (s *E2ESuite) CreateTeam(t testing.TB, teamName string, members []map[string]any) *dto.TeamDTO {
	t.Helper()
	resp := s.PerformRequest(t, http.MethodPost, "/team/add", map[string]any{
		"team_name": teamName,
		"members":   members,
	})
	s.ExpectStatus(t, resp, http.StatusCreated)

	var parsed CreateTeamResponse
	s.DecodeBody(t, resp, &parsed)
	return parsed.Team
}

func (s *E2ESuite) CreatePullRequest(t testing.TB, prID, name, authorID string) *dto.PullRequestDTO {
	t.Helper()

	resp := s.PerformRequest(t, http.MethodPost, "/pullRequest/create", map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": name,
		"author_id":         authorID,
	})
	s.ExpectStatus(t, resp, http.StatusCreated)

	var parsed PullRequestResponse
	s.DecodeBody(t, resp, &parsed)
	return parsed.PR
}

func (s *E2ESuite) MergePullRequest(t testing.TB, prID string) *dto.PullRequestDTO {
	t.Helper()

	resp := s.PerformRequest(t, http.MethodPost, "/pullRequest/merge", map[string]any{
		"pull_request_id": prID,
	})
	s.ExpectStatus(t, resp, http.StatusOK)

	var parsed PullRequestResponse
	s.DecodeBody(t, resp, &parsed)
	return parsed.PR
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

func newRequest(t testing.TB, method, url string, payload any) *http.Request {
	t.Helper()

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req
}

func newRawRequest(t testing.TB, method, url, rawBody, contentType string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, url, strings.NewReader(rawBody))
	require.NoError(t, err)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req
}

func resetTables(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const truncateTablesSQL = `TRUNCATE TABLE pr_reviewers, pull_requests, users, teams RESTART IDENTITY CASCADE`
	_, err := pool.Exec(ctx, truncateTablesSQL)
	require.NoError(t, err)
}
