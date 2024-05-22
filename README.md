# QueryBuilder

`QueryBuilder` is a Go package that provides a flexible and intuitive way to build query strings with filters and sorts. It uses a builder pattern to incrementally construct the query string, ensuring it is correctly formatted and ready to use in HTTP requests.

## Installation

To use `QueryBuilder`, you need to have Go installed on your machine. You can install the package using `go get`:

```sh
go get github.com/localrivet/buildsql
```

## Usage

### Creating a New QueryBuilder

To start building a query string, create a new instance of `QueryBuilder`:

```go
qb := buildsql.NewQueryBuilder()
```

### Adding Filters

Filters are added using the `AddFilter` method. The method takes four parameters:
- `prefix`: The table prefix.
- `fieldName`: The name of the field to filter on.
- `operator`: The operator to use for filtering (e.g., `eq` for equal, `lt` for less than, etc.).
- `value`: The value to filter by.

Example:

```go
qb.AddFilter("r", "user_id", buildsql.Equal, "u7fb0d70550c849")
qb.AddFilter("r", "account_id", buildsql.Equal, "a7fb0d70550c849")
```

### Adding Sorts

Sorts are added using the `AddSort` method. The method takes three parameters:
- `prefix`: The table prefix.
- `fieldName`: The name of the field to sort on.
- `direction`: The sort direction (`-` for descending, empty for ascending).

Example:

```go
qb.AddSort("r", "created_at", "")
qb.AddSort("r", "created_at", "-")
```

### Building the Query String

Once all filters and sorts have been added, call the `Build` method to construct the final query string:

```go
queryString, err := qb.Build()
if err != nil {
    fmt.Println("Error:", err)
    return
}

fmt.Println("Query String:", queryString)
```

### Full Example

```go
package main

import (
	"fmt"
	"buildsql"
)

func main() {
	qb := buildsql.NewQueryBuilder()
	qb.AddFilter("r", "user_id", buildsql.Equal, "u7fb0d70550c849")
	qb.AddFilter("r", "account_id", buildsql.Equal, "a7fb0d70550c849")
	qb.AddSort("r", "created_at", "")
	qb.AddSort("r", "created_at", "-")

	queryString, err := qb.Build()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Query String:", queryString)
}
```

### Output

When you run the above code, the output will be:

```
Query String: filter=r-user_id-eq-u7fb0d70550c849&filter=r-account_id-eq-a7fb0d70550c849&sortOn=r-created_at&sortOn=-r-created_at
```

## Query String Format

### Filters

Filters follow the format: `table prefix` `-` `field name` `-` `operator` `-` `field value`.

Example:
```
filter=r-user_id-eq-u7fb0d70550c849
```
- `r`: Table prefix.
- `user_id`: Field name.
- `eq`: Operator (equal).
- `u7fb0d70550c849`: Field value.

### Sorts

Sorts follow the format: `optional ASC/DESC prefix` `table prefix` `-` `field name`.

Example:
```
sortOn=-r-created_at
```
- `-`: DESC prefix (optional).
- `r`: Table prefix.
- `created_at`: Field name.

## Sample Query String

A complete query string with multiple filters and sorts:

```
https://example.org/?filter=r-user_id-eq-u7fb0d70550c849&filter=r-account_id-eq-a7fb0d70550c849&sortOn=r-created_at&sortOn=-r-created_at
```

## Protips

- The `-` sign prefixing a field in the `sortOn` parameter indicates a DESC sort order. No prefix indicates an ASC sort order.
- Filters are always combined using an `AND` operator.

## Operator Type and Constants

### Operator Type

The `Operator` type defines constants for various operators used in filters.

```go
package buildsql

type Operator string

const (
	Equal              Operator = "eq"
	NotEqual           Operator = "neq"
	Like               Operator = "like"
	ILike              Operator = "ilike"
	OrLike             Operator = "orlike"
	OrILike            Operator = "orilike"
	NotLike            Operator = "nlike"
	NotILike           Operator = "nilike"
	LessThan           Operator = "lt"
	LessThanOrEqual    Operator = "lte"
	GreaterThan        Operator = "gt"
	GreaterThanOrEqual Operator = "gte"
	Between            Operator = "btw"
	Or                 Operator = "or"
	In                 Operator = "in"
	NotIn              Operator = "notin"
	IsNull             Operator = "isnull"
	IsNotNull          Operator = "isnotnull"
)

func (o Operator) Convert() string {
	switch o {
	case Equal:
		return "="
	case NotEqual:
		return "!="
	case Like:
		return "LIKE"
	case ILike:
		return "ILIKE"
	case OrLike:
		return "LIKE"
	case OrILike:
		return "ILIKE"
	case NotLike:
		return "NOT LIKE"
	case NotILike:
		return "NOT ILIKE"
	case LessThan:
		return "<"
	case LessThanOrEqual:
		return "<="
	case GreaterThan:
		return ">"
	case GreaterThanOrEqual:
		return ">="
	case Between:
		return "BETWEEN"
	case In:
		return "IN"
	case NotIn:
		return "NOT IN"
	case IsNull:
		return "IS NULL"
	case IsNotNull:
		return "IS NOT NULL"
	}
	return ""
}

func (o Operator) isLike() bool {
	return o == Like || o == OrLike || o == ILike || o == OrILike || o == NotLike || o == NotILike
}
```

## How It Works

Dynamically generates `WHERE`, `ORDER BY` AND `NAMED PARAMETER MAP` for queries using the `sqlx` package. Supports both Postgres and MySQL.

### Example

```go
func (m *customerRepo) GetAllCustomers(currentPage, pageSize int64, filter string) (res *getAllCustomersResponse, err error) {
	var builder = buildsql.NewQueryBuilder()
	where, orderBy, namedParamMap, err := builder.Build(filter, map[string]interface{}{
		"c": models.Customer{}, // customer alias
	})
	if err != nil {
		return nil, err
	}

	if where != "" {
		where = fmt.Sprintf("WHERE 1 = 1 %s", where)
	}

	// set a default order by
	if orderBy == "" {
		orderBy = "ORDER BY c.first_name ASC"
	}
	limit := fmt.Sprintf("LIMIT %d, %d", currentPage*pageSize, pageSize)

	sql := fmt.Sprintf(`
		SELECT
			-- customer
			c.id as "customer.id",
			c.first_name as "customer.first_name",
			c.last_name as "customer.last_name",
			c.email as "customer.email",
			-- stats
			COUNT(*) OVER() AS "pagingstats.total_records"
		FROM customer c
		%s
		%s
		%s
	`, where, orderBy, limit)

	nstmt, err := m.db.PrepareNamed(sql)
	if err != nil {
		return nil, fmt.Errorf("error::GetAllCustomers::%s", err.Error())
	}

	var result []*struct {
		Customer    models.Customer   `db:"customer"`
		PagingStats types.PagingStats `db:"pagingstats"`
	}

	namedParamMap["offset"] = currentPage * pageSize
	namedParamMap["limit"] = pageSize

	err = nstmt.Select(&result, namedParamMap)

	results := []models.Customer{}

	var stats *types.PagingStats = &types.PagingStats{}
	for i, r := range result {
		if i == 0 {
			stats = r.PagingStats.Calc(pageSize)
		}
		results = append(results, r.Customer)
	}

	out := &getAllCustomersResponse{
		Results:     results,
		PagingStats: *stats,
	}

	return out, err
}

type getAllCustomersResponse struct {
	PagingStats types.PagingStats `json:"stats"`
	Results     []models.Customer `json:"results"`
}

type PagingStats struct {
	TotalRecords int64 `db:"total_records" json:"total_records"`
	TotalPages   int64 `db:"total_pages" json:"total_pages"`
}

func (s *PagingStats) Calc(pageSize int64) *PagingStats {
	totalPages := float64(s.TotalRecords) / float64(pageSize)
	s.TotalPages = int64(math.Ceil(totalPages))
	return s
}
```

### Sample Query String Formats

Delimiter is hyphen: [http://www.blooberry.com/indexdot/html/topics/urlencoding.htm](http://www.blooberry

.com/indexdot/html/topics/urlencoding.htm)

**Filter**: firstName = 'bob' ORDER BY 'id' DESC  
_Protip: the '-' sign prefixing the 'id' field indicates a DESC, no prefix indicates an ASC_

`https://example.org/?filter=u-firstName-eq-bob&sortOn=-u-id`

**filter**: field format is: 'table prefix' 'hyphen' 'fieldname' 'hyphen' 'operator' 'hyphen' 'field value'

_Example_: u-firstName-eq-bob

```
u-firstName-eq-bob
```
| u            | -      | firstName | -      | eq   | -      | bob         |
| ------------ | ------ | --------- | ------ | ---- | ------ | ----------- |
| table prefix | hyphen | fieldName | hyphen | op   | hyphen | field value |

**sortOn**: field format is: 'optional ASC/DESC prefix' 'table prefix' 'hyphen' 'fieldname'

_Example_: -u-id

```
-u-id
```
| -                        | u            | -      | id          |
| ------------------------ | ------------ | ------ | ----------- |
| optional ASC/DESC prefix | table prefix | hyphen | field value |

**Filter**: firstName = 'bob' AND lastName = 'philips' ORDER BY 'id' DESC

`https://example.org/?filter=u-firstName-eq-bob&filter=u-lastName-eq-philips&sortOn=u-id&sortOn=-u-firstName`

Assume the filter is always an "AND"
check the allowedFields for the fieldnames
return an error if an unknown fieldname

In both `AllowedFilterFields` and `AllowedSortFields`
the map[string]string maps to:

```
map[string]string{
		"id":     "p",  // product alias
		"name":   "p",  // product alias
		"slug":   "p",  // product alias
		"sku":    "v",  // variant alias
		"amount": "pr", // price alias
	}
```

## License

This project is licensed under the MIT License.

## Author

This package is managed by localrivet and is expressly authorized for use without limitation.
