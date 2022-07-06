# Build SQL

Dynamically generates `WHERE`, `ORDER BY` AND `NAMED PARAMETER MAP` for queries using the `sqlx` package. Supports both postgres and mysql.

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

	// fmt.Println("sql:", sql)
	// fmt.Println("where:", where)
	// fmt.Println("order by:", orderBy)
	// fmt.Println("limit:", limit)

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
```

### How It Works

Sample query string formats
Delimiter is hyphen: http://www.blooberry.com/indexdot/html/topics/urlencoding.htm

**Filter**: firstName = 'bob' ORDER BY 'id' DESC  
_Protip: the '-' sign prefixing the 'id' field indicates a DESC, no prefix indicates an ASC_

`https://example.org/?filter=u-firstName-bob&sortOn=-u-id`

**filter**: field format is: 'table prefix' 'hyphen' 'fieldname' 'hyphen' 'field value'

_Example_: u-firstName-bob

```
u-firstName-bob = u      		-      	firstName       -			bob
                |				|			|			|			 |
                table prefix	      hyphen    fieldName	  hyphen	 field value
```

**sortOn**: field format is: 'optional ASC/DESC prefix' 'table prefix' 'hyphen' 'fieldname'

_Example_: u-firstName-bob

```
-u-id =		 			    - 						u      		-      		   id
						 	|						|			|				|
 			    optional ASC/DESC prefix		table prefix  hyphen	    fieldName
```

**Filter**: firstName = 'bob' AND lastName = 'philips' ORDER BY 'id' DESC

`https://example.org/?filter=u-firstName-bob&filter=u-lastName-philips&sortOn=u-lastName&sortOn=-u-firstName`

Assume the filter is always an "AND"
check the allowedFields for the fieldnames
return an error if a unknown fieldname

In both AllowedFilterFields and AllowedSortFields
the map[string]string maps to:

```
map     [string]		 string
		  |                 |
     fieldName          table alias
```

Example:

```
map[string]string{
		"id":     "p",  // product alias
		"name":   "p",  // product alias
		"slug":   "p",  // product alias
		"sku":    "v",  // product alias
		"amount": "pr", // price alias
	}
```
