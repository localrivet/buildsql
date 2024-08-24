package buildsql_test

import (
	"database/sql"
	"net/url"
	"testing"
	"time"

	"github.com/localrivet/buildsql"
	"github.com/stretchr/testify/assert"
)

type Product struct {
	ID     int64   `json:"id" db:"id"`         // id
	Name   string  `json:"name" db:"name"`     // name
	Slug   string  `json:"slug" db:"slug"`     // slug
	Sku    string  `json:"sku" db:"sku"`       // sku
	Amount float64 `json:"amount" db:"amount"` // amount
}

type Pricing struct {
	ID        int64   `json:"id" db:"id"`                 // id
	ProductID int64   `json:"product_id" db:"product_id"` // product_id
	Amount    float64 `json:"amount" db:"amount"`         // amount
}

func TestQueryBuilder(t *testing.T) {

	t.Run("should return valid AllowedFilterFields from a string map", func(t *testing.T) {
		builder := buildsql.NewQueryBuilder()
		builder.AllowedFilterFields = map[string]string{
			"id":     "p",  // product alias
			"name":   "p",  // product alias
			"slug":   "p",  // product alias
			"sku":    "p",  // product alias
			"amount": "pr", // price alias
		}
		assert.NotNil(t, builder.AllowedFilterFields)
	})

	t.Run("should correctly parse a param string", func(t *testing.T) {
		builder := buildsql.NewQueryBuilder()
		on := "filter=p-name-eq-Practical Cotton Gloves&filter=p-sku-eq-practical-cotton-gloves&sortOn=p-name&sortOn=-pr-amount"

		err := builder.ParseParamString(on)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(builder.Filters))
		assert.Equal(t, "p", builder.Filters[0].TableAlias)
		assert.Equal(t, "name", builder.Filters[0].FieldName)
		assert.Equal(t, "Practical Cotton Gloves", builder.Filters[0].Value)
		assert.Equal(t, "p", builder.Filters[1].TableAlias)
		assert.Equal(t, "sku", builder.Filters[1].FieldName)
		assert.Equal(t, "practical-cotton-gloves", builder.Filters[1].Value)
		assert.Equal(t, 2, len(builder.Sorts))
		assert.Equal(t, "p", builder.Sorts[0].TableAlias)
		assert.Equal(t, "name", builder.Sorts[0].FieldName)
		assert.Equal(t, buildsql.ASC, builder.Sorts[0].Direction)
		assert.Equal(t, "pr", builder.Sorts[1].TableAlias)
		assert.Equal(t, "amount", builder.Sorts[1].FieldName)
		assert.Equal(t, buildsql.DESC, builder.Sorts[1].Direction)
	})

	t.Run("should error on parsing an invalid param string", func(t *testing.T) {
		builder := buildsql.NewQueryBuilder()
		on := "filter=c_name-Practical Cotton Gloves&filter=v-sku_practical-cotton-gloves&sortOn=p_name&sortOn=-pr_amount"

		err := builder.ParseParamString(on)
		assert.NotNil(t, err)
	})

	t.Run("should correctly build query", func(t *testing.T) {
		builder := buildsql.NewQueryBuilder()
		on := "filter=p-name-like-Practical&filter=p-name-nlike-Cotton&filter=p-sku-eq-practical-cotton-gloves&sortOn=p-name&sortOn=-pr-amount"

		where, orderBy, namedParamMap, err := builder.Build(on, map[string]interface{}{
			"p":  Product{}, // product alias
			"pr": Pricing{}, // pricing alias
		})
		assert.Nil(t, err)
		assert.Contains(t, where, "p.name LIKE :filter_p_name_0")
		assert.Contains(t, where, "p.name NOT LIKE :filter_p_name_1")
		assert.Contains(t, where, "p.sku = :filter_p_sku_0")
		assert.Contains(t, orderBy, "p.name ASC")
		assert.Contains(t, orderBy, "pr.amount DESC")
		assert.Equal(t, "%Practical%", namedParamMap["filter_p_name_0"])
		assert.Equal(t, "%Cotton%", namedParamMap["filter_p_name_1"])
		assert.Equal(t, "practical-cotton-gloves", namedParamMap["filter_p_sku_0"])
	})

	t.Run("should parse raw URL", func(t *testing.T) {
		builder := buildsql.NewQueryBuilder()
		on := "http://example.com/v1/products/0/20?filter=p-name-like-cotton&sortOn=p-id"

		decodedValue, err := url.QueryUnescape(on)
		assert.Nil(t, err)

		where, orderBy, namedParamMap, err := builder.Build(decodedValue, map[string]interface{}{
			"p": Product{}, // product alias
		})
		assert.Nil(t, err)
		assert.Contains(t, where, "p.name LIKE :filter_p_name_0")
		assert.Equal(t, "ORDER BY p.id ASC", orderBy)
		assert.Equal(t, "%cotton%", namedParamMap["filter_p_name_0"])
	})
}

func TestQueryBuilderOr(t *testing.T) {
	t.Run("should handle OrLike and OrILike conditions correctly", func(t *testing.T) {
		builder := buildsql.NewQueryBuilder()
		on := "filter=u-first_name-orlike-John&filter=u-last_name-orilike-Doe&filter=u-id-eq-123"

		where, _, namedParamMap, err := builder.Build(on, map[string]interface{}{
			"u": User{},
		})
		assert.Nil(t, err)
		assert.Contains(t, where, "u.id = :filter_u_id_0")
		assert.Contains(t, where, "(u.first_name LIKE :filter_u_first_name_0 OR u.last_name ILIKE :filter_u_last_name_0)")
		assert.Equal(t, "%John%", namedParamMap["filter_u_first_name_0"])
		assert.Equal(t, "%Doe%", namedParamMap["filter_u_last_name_0"])
		assert.Equal(t, "123", namedParamMap["filter_u_id_0"])

		// Check the complete where clause
		expectedWhere := " AND u.id = :filter_u_id_0 AND (u.first_name LIKE :filter_u_first_name_0 OR u.last_name ILIKE :filter_u_last_name_0)"
		assert.Equal(t, expectedWhere, where)
	})
}
func TestQueryBuilderBetween(t *testing.T) {
	t.Run("should correctly parse a param string with between", func(t *testing.T) {
		builder := buildsql.NewQueryBuilder()
		on := "filter=r-user_id-eq-u07b4b9def3d3c0&filter=r-account_id-eq-a091321bd573491&filter=r-created_at-btw-2024-06-12 00:00:00,2024-06-12 23:59:59&sortOn=-r-created_at"

		err := builder.ParseParamString(on)
		assert.Nil(t, err)

		// Assert Filters
		assert.Equal(t, 3, len(builder.Filters))
		assert.Equal(t, "r", builder.Filters[0].TableAlias)
		assert.Equal(t, "user_id", builder.Filters[0].FieldName)
		assert.Equal(t, "eq", builder.Filters[0].Operator.String())
		assert.Equal(t, "u07b4b9def3d3c0", builder.Filters[0].Value)

		assert.Equal(t, "r", builder.Filters[1].TableAlias)
		assert.Equal(t, "account_id", builder.Filters[1].FieldName)
		assert.Equal(t, "eq", builder.Filters[1].Operator.String())
		assert.Equal(t, "a091321bd573491", builder.Filters[1].Value)

		assert.Equal(t, "r", builder.Filters[2].TableAlias)
		assert.Equal(t, "created_at", builder.Filters[2].FieldName)
		assert.Equal(t, "btw", builder.Filters[2].Operator.String())
		assert.Equal(t, []string{"2024-06-12 00:00:00", "2024-06-12 23:59:59"}, builder.Filters[2].Values)

		// Assert Sorts
		assert.Equal(t, 1, len(builder.Sorts))
		assert.Equal(t, "r", builder.Sorts[0].TableAlias)
		assert.Equal(t, "created_at", builder.Sorts[0].FieldName)
		assert.Equal(t, buildsql.DESC, builder.Sorts[0].Direction)

		// fmt.Println("builder", builder)
	})
}

type User struct {
	ID                     int64          `json:"id" db:"id" form:"id"`                                                                      // id
	FirstName              string         `json:"first_name" db:"first_name" form:"first_name"`                                              // first_name
	LastName               string         `json:"last_name" db:"last_name" form:"last_name"`                                                 // last_name
	Title                  sql.NullString `json:"title" db:"title" form:"title"`                                                             // title
	Username               string         `json:"username" db:"username" form:"username"`                                                    // username
	Email                  string         `json:"email" db:"email" form:"email"`                                                             // email
	EmailVisibility        bool           `json:"email_visibility" db:"email_visibility" form:"email_visibility"`                            // email_visibility
	RequireReset           bool           `json:"require_reset" db:"require_reset" form:"require_reset"`                                     // require_reset
	LastResetSentAt        sql.NullTime   `json:"last_reset_sent_at" db:"last_reset_sent_at" form:"last_reset_sent_at"`                      // last_reset_sent_at
	LastVerificationSentAt sql.NullTime   `json:"last_verification_sent_at" db:"last_verification_sent_at" form:"last_verification_sent_at"` // last_verification_sent_at
	PasswordHash           string         `json:"password_hash" db:"password_hash" form:"password_hash"`                                     // password_hash
	TokenKey               string         `json:"token_key" db:"token_key" form:"token_key"`                                                 // token_key
	Verified               bool           `json:"verified" db:"verified" form:"verified"`                                                    // verified
	Avatar                 string         `json:"avatar" db:"avatar" form:"avatar"`                                                          // avatar
	TypeID                 int64          `json:"type_id" db:"type_id" form:"type_id"`                                                       // type_id
	CreatedAt              time.Time      `json:"created_at" db:"created_at" form:"created_at"`                                              // created_at
	UpdatedAt              time.Time      `json:"updated_at" db:"updated_at" form:"updated_at"`                                              // updated_at

}
