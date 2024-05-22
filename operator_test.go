package buildsql_test

import (
	"testing"

	"github.com/localrivet/buildsql"
	"github.com/stretchr/testify/assert"
)

func TestOperator(t *testing.T) {
	t.Run("Convert should return correct SQL representation", func(t *testing.T) {
		assert.Equal(t, "=", buildsql.Equal.Convert())
		assert.Equal(t, "!=", buildsql.NotEqual.Convert())
		assert.Equal(t, "LIKE", buildsql.Like.Convert())
		assert.Equal(t, "ILIKE", buildsql.ILike.Convert())
		assert.Equal(t, "LIKE", buildsql.OrLike.Convert())
		assert.Equal(t, "ILIKE", buildsql.OrILike.Convert())
		assert.Equal(t, "NOT LIKE", buildsql.NotLike.Convert())
		assert.Equal(t, "NOT ILIKE", buildsql.NotILike.Convert())
		assert.Equal(t, "<", buildsql.LessThan.Convert())
		assert.Equal(t, "<=", buildsql.LessThanOrEqual.Convert())
		assert.Equal(t, ">", buildsql.GreaterThan.Convert())
		assert.Equal(t, ">=", buildsql.GreaterThanOrEqual.Convert())
		assert.Equal(t, "BETWEEN", buildsql.Between.Convert())
		assert.Equal(t, "IN", buildsql.In.Convert())
		assert.Equal(t, "NOT IN", buildsql.NotIn.Convert())
		assert.Equal(t, "IS NULL", buildsql.IsNull.Convert())
		assert.Equal(t, "IS NOT NULL", buildsql.IsNotNull.Convert())
	})

	t.Run("IsLike should return true for like operators", func(t *testing.T) {
		assert.True(t, buildsql.Like.IsLike())
		assert.True(t, buildsql.OrLike.IsLike())
		assert.True(t, buildsql.ILike.IsLike())
		assert.True(t, buildsql.OrILike.IsLike())
		assert.True(t, buildsql.NotLike.IsLike())
		assert.True(t, buildsql.NotILike.IsLike())
	})

	t.Run("IsLike should return false for non-like operators", func(t *testing.T) {
		assert.False(t, buildsql.Equal.IsLike())
		assert.False(t, buildsql.NotEqual.IsLike())
		assert.False(t, buildsql.LessThan.IsLike())
		assert.False(t, buildsql.GreaterThan.IsLike())
		assert.False(t, buildsql.Between.IsLike())
	})
}
