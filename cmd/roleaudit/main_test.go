package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleAuditSupportedRolesExitZero(t *testing.T) {
	queryer := &stubQueryer{
		rows: &stubRows{
			buckets: []roleBucket{
				{Role: "admin", Count: 1},
				{Role: "operator", Count: 2},
				{Role: "viewer", Count: 3},
			},
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runAudit(context.Background(), queryer, &stdout, &stderr)

	assert.Equal(t, 0, exitCode)
	assert.Empty(t, stderr.String())
	assert.Contains(t, stdout.String(), "Persisted role counts:")
	assert.Contains(t, stdout.String(), "- admin: 1")
	assert.Contains(t, stdout.String(), "- operator: 2")
	assert.Contains(t, stdout.String(), "- viewer: 3")
	assert.Contains(t, stdout.String(), "Audit passed: only supported roles found")
	require.Len(t, queryer.queries, 1)
	assert.Contains(t, queryer.queries[0], "GROUP BY role")
	assert.NotContains(t, queryer.queries[0], "INSERT")
	assert.NotContains(t, queryer.queries[0], "UPDATE")
	assert.NotContains(t, queryer.queries[0], "DELETE")
}

func TestRoleAuditUnsupportedRolesExitNonZero(t *testing.T) {
	queryer := &stubQueryer{
		rows: &stubRows{
			buckets: []roleBucket{
				{Role: "admin", Count: 1},
				{Role: "owner", Count: 2},
			},
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runAudit(context.Background(), queryer, &stdout, &stderr)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stdout.String(), "- admin: 1")
	assert.Contains(t, stdout.String(), "- owner: 2")
	assert.Contains(t, stderr.String(), "Unsupported persisted roles detected:")
	assert.Contains(t, stderr.String(), "- owner: 2")
}

func TestRoleAuditQueryFailureExitNonZero(t *testing.T) {
	queryer := &stubQueryer{
		err: errors.New("boom"),
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runAudit(context.Background(), queryer, &stdout, &stderr)

	assert.Equal(t, 1, exitCode)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "role audit failed: boom")
}

func TestSetSessionReadOnlyUsesExplicitReadOnlyStatement(t *testing.T) {
	execer := &stubSessionExecer{}

	err := setSessionReadOnly(context.Background(), execer)

	require.NoError(t, err)
	assert.Equal(t, []string{setReadOnlySQL}, execer.queries)
}

func TestSetSessionReadOnlyPropagatesFailure(t *testing.T) {
	execer := &stubSessionExecer{err: errors.New("read only failed")}

	err := setSessionReadOnly(context.Background(), execer)

	require.EqualError(t, err, "read only failed")
}

type stubQueryer struct {
	queries []string
	rows    rowScanner
	err     error
}

func (s *stubQueryer) QueryContext(_ context.Context, query string, _ ...any) (rowScanner, error) {
	s.queries = append(s.queries, query)
	if s.err != nil {
		return nil, s.err
	}
	return s.rows, nil
}

type stubRows struct {
	index   int
	buckets []roleBucket
	err     error
}

func (s *stubRows) Next() bool {
	return s.index < len(s.buckets)
}

func (s *stubRows) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}
	bucket := s.buckets[s.index]
	s.index++
	*(dest[0].(*string)) = bucket.Role
	*(dest[1].(*int64)) = bucket.Count
	return nil
}

func (s *stubRows) Err() error {
	return s.err
}

func (s *stubRows) Close() error {
	return nil
}

type stubSessionExecer struct {
	queries []string
	err     error
}

func (s *stubSessionExecer) ExecContext(_ context.Context, query string, _ ...any) (sql.Result, error) {
	s.queries = append(s.queries, query)
	return nil, s.err
}
