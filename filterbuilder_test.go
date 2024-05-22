package buildsql_test

import (
	"fmt"
	"testing"

	"github.com/localrivet/buildsql"
	"github.com/stretchr/testify/assert"
)

func TestFilterBuilder(t *testing.T) {
	t.Run("AddFilter should add a filter", func(t *testing.T) {
		fb := buildsql.NewFilterBuilder()
		fb.AddFilter("r", "user_id", buildsql.Equal, "u7fb0d70550c849")
		expected := "filter=r-user_id-eq-u7fb0d70550c849"
		assert.Contains(t, fb.String(), expected)
	})

	t.Run("AddSort should add a sort in ascending order", func(t *testing.T) {
		fb := buildsql.NewFilterBuilder()
		fb.AddSort("r", "created_at", buildsql.ASC)
		expected := "sortOn=r-created_at"
		assert.Contains(t, fb.String(), expected)
	})

	t.Run("AddSort should add a sort in descending order", func(t *testing.T) {
		fb := buildsql.NewFilterBuilder()
		fb.AddSort("r", "created_at", buildsql.DESC)
		expected := "sortOn=-r-created_at"
		assert.Contains(t, fb.String(), expected)
	})

	t.Run("Build should construct a query string with filters and sorts", func(t *testing.T) {
		fb := buildsql.NewFilterBuilder()
		fb.AddFilter("r", "user_id", buildsql.Equal, "u7fb0d70550c849")
		fb.AddFilter("r", "account_id", buildsql.Equal, "a7fb0d70550c849")
		fb.AddSort("r", "created_at", buildsql.ASC)
		fb.AddSort("r", "created_at", buildsql.DESC)
		expected := "filter=r-user_id-eq-u7fb0d70550c849&filter=r-account_id-eq-a7fb0d70550c849&sortOn=r-created_at&sortOn=-r-created_at"
		assert.Equal(t, expected, fb.String())
	})

	t.Run("Build should return an empty string if no filters or sorts are added", func(t *testing.T) {
		fb := buildsql.NewFilterBuilder()
		expected := ""
		assert.Equal(t, expected, fb.String())
	})

	t.Run("AddFilter should handle invalid filter format", func(t *testing.T) {
		fb := buildsql.NewFilterBuilder()
		fb.AddFilter("r", "user_id", buildsql.Equal, "u7fb0d70550c849")
		fb.AddFilter("invalid", "format", buildsql.Operator(""), "value")
		expected := "filter=r-user_id-eq-u7fb0d70550c849"
		assert.Equal(t, expected, fb.String())
	})

	t.Run("AddSort should handle invalid sort format", func(t *testing.T) {
		fb := buildsql.NewFilterBuilder()
		fb.AddSort("r", "created_at", buildsql.ASC)
		fb.AddSort("r", "id", buildsql.DESC)
		fb.AddSort("r", "updated_at")
		expected := "sortOn=r-created_at&sortOn=-r-id&sortOn=r-updated_at"
		fmt.Println("fb.String()", fb.String())
		assert.Equal(t, expected, fb.String())
	})
}
