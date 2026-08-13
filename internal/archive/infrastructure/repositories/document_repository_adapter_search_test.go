package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
)

func TestAppendSearchQueryFilter_MultiTokenAndYear(t *testing.T) {
	where := []string{sqlWorkspaceIDEq}
	args := []any{"ws-1"}
	argN := 2

	where, args, argN = appendSearchQueryFilter(where, args, argN, "cita 1998")
	requireTokens := entities.TokenizeSearchQuery("cita 1998")
	assert.Equal(t, []string{"cita", "1998"}, requireTokens)
	assert.Len(t, where, 3) // workspace + 2 token clauses
	assert.Contains(t, where[1], "title ILIKE")
	assert.Contains(t, where[2], "to_char(document_date")
	assert.Equal(t, "ws-1", args[0])
	assert.Equal(t, "%cita%", args[1])
	assert.Equal(t, "%1998%", args[2])
	assert.Equal(t, "1998", args[3])
	assert.Equal(t, 5, argN)
}

func TestAppendSearchQueryFilter_Empty(t *testing.T) {
	where := []string{sqlWorkspaceIDEq}
	args := []any{"ws-1"}
	gotWhere, gotArgs, gotN := appendSearchQueryFilter(where, args, 2, "   ")
	assert.Equal(t, where, gotWhere)
	assert.Equal(t, args, gotArgs)
	assert.Equal(t, 2, gotN)
}

func TestEscapeILIKEPattern(t *testing.T) {
	assert.Equal(t, `100\%`, escapeILIKEPattern("100%"))
	assert.Equal(t, `a\_b`, escapeILIKEPattern("a_b"))
}
