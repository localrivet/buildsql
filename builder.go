// This is from the open source package K8sCommerce managed by Alma Tuck
// It is expressly authorized to be used within this software without limitation
package buildsql

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

//
// Sample query string formats
// Delimiter is hyphen: http://www.blooberry.com/indexdot/html/topics/urlencoding.htm
//
// Filter: firstName = 'bob' ORDER BY 'id' DESC
// Protip: the '-' sign prefixing the 'id' field indicates a DESC
// no prefix indicates an ASC
//
// https://example.org/?filter=u-firstName-eq-bob&sortOn=-u-id
//
// filter: field format is: 'table prefix' 'hyphen' 'fieldname' 'hyphen' 'operator' 'hyphen' 'field value'
// Example: u-firstName-eq-bob
// u-firstName-eq-bob =		 u      		-      	firstName       -		  eq       -		  bob
// 						 	 |				|			|			   |          |         |
//  					table prefix	 hyphen    fieldName		hyphen	 operator	 hyphen	  field value
//
//
// sortOn: field format is: 'optional ASC/DESC prefix' 'table prefix' 'hyphen' 'fieldname'
// Example: -u-id
// -u-id =		 			- 						u      		-      		   id
// 						 	|						|			|				|
//  			optional ASC/DESC prefix		table prefix  hyphen	fieldName
//
//
//
// Filter: firstName = 'bob' AND lastName = 'philips' ORDER BY 'id' DESC
// https://example.org/?filter=u-firstName-eq-bob&filter=u-lastName-eq-philips&sortOn=u-lastName&sortOn=-u-firstName
//
// Assume the filter is always an "AND"
// check the allowedFields for the fieldnames
// return an error if an unknown fieldname
//
// Example of how allowed fields maps work:
// map[string]string{
//		"id":     "p",  // product alias
//		"name":   "p",  // product alias
//		"slug":   "p",  // product alias
//		"sku":    "v",  // product alias
//		"amount": "pr", // price alias
//	}

var Delimiter string = "-"

type SortDirection string

const (
	ASC  SortDirection = "ASC"
	DESC SortDirection = "DESC"
)

type FilterField struct {
	TableAlias string
	FieldName  string
	Operator   Operator
	Value      interface{}
	Values     []string
}
type SortField struct {
	TableAlias string
	FieldName  string
	Direction  SortDirection
}

type Where struct {
	CombinedName string
	SqlString    string
	Named        string
	Operator     Operator
}

func NewQueryBuilder() QueryBuilder {
	return QueryBuilder{}
}

type QueryBuilder struct {
	AllowedFilterFields map[string]string
	AllowedSortFields   map[string]string
	Filters             []FilterField
	Sorts               []SortField
	SearchTables        map[string]int
}

// AllowedFiltersFieldsFromMap
// resets AllowedFilterFields
// example:
//
//	map[string]string{
//			"id":     "p",  // product alias
//			"name":   "p",  // product alias
//			"slug":   "p",  // product alias
//			"sku":    "v",  // product alias
//			"amount": "pr", // price alias
//		}
//
//	func (b *QueryBuilder) AllowedFiltersFieldsFromStringMap(allowed map[string]string) {
//		b.AllowedFilterFields = allowed
//	}
func (b *QueryBuilder) ParseParamString(paramString string) error {
	if paramString == "" {
		paramString = "?"
	}
	// fmt.Println("paramString: ", paramString)

	b.SearchTables = make(map[string]int)

	if strings.Index(paramString, "?") != 0 {
		pathParts := strings.Split(paramString, "?")

		// fmt.Println("pathParts", pathParts)
		if len(pathParts) > 1 {
			paramString = pathParts[1]
		}
	}

	prefix := paramString[0:1]
	if prefix != "?" {
		paramString = "?" + paramString
	}

	// let's let the url parser do the work
	u, err := url.Parse(paramString)
	if err != nil {
		return err
	}
	q := u.Query()
	// fmt.Println(q)

	// fmt.Println("Q:", q)

	// parse filters
	if filters, ok := q["filter"]; ok {
		var count int // Initialize count
		for _, filter := range filters {
			filter = strings.TrimSpace(filter)
			parts := strings.SplitN(filter, Delimiter, 4)

			if len(parts) < 3 {
				return fmt.Errorf("filter: %s has too few params", filter)
			}

			var filterField FilterField
			filterField.TableAlias = parts[0]
			filterField.FieldName = parts[1]

			// Handling different operator scenarios
			operatorPart := parts[2]
			var valuePart string

			if len(parts) > 3 {
				// Assuming the operator is one of eq, lt, gt, etc., and the next part is the value
				filterField.Operator = Operator(operatorPart)
				valuePart = parts[3]

				if filterField.Operator.IsBetween() || filterField.Operator.IsIn() || filterField.Operator.IsNotIn() {
					sp := strings.Split(valuePart, ",")
					filterField.Values = sp
				}
			} else {
				// Handling scenarios where the operator might include the value (e.g., isnull, isnotnull)
				if operatorPart == "isnull" || operatorPart == "isnotnull" {
					filterField.Operator = Operator(operatorPart)
				} else {
					// Splitting the operator and the value
					opAndValue := strings.SplitN(operatorPart, "-", 2)
					if len(opAndValue) != 2 {
						return fmt.Errorf("invalid operator and value combination: %s", operatorPart)
					}
					filterField.Operator = Operator(opAndValue[0])
					valuePart = opAndValue[1]
				}
			}

			// Assigning the value
			if filterField.Operator.IsLike() {
				filterField.Value = "%" + valuePart + "%"
			} else {
				filterField.Value = valuePart
			}

			b.Filters = append(b.Filters, filterField)
			b.SearchTables[filterField.TableAlias] = count + 1
		}
	}

	// parse sorts
	if sortOns, ok := q["sortOn"]; ok {
		count := 0
		for _, sort := range sortOns {
			// check for the direction first
			// since the delimiter is the same as the
			// sort direction prefix
			sort := strings.TrimSpace(sort)
			dir := ASC
			if isDesc := strings.HasPrefix(sort, "-"); isDesc {
				dir = DESC
				sort = sort[1:]
			}

			parts := strings.Split(sort, Delimiter)
			if len(parts) < 1 {
				return fmt.Errorf("sortOn: %s has too few params", sort)
			}

			sortField := SortField{
				TableAlias: parts[0],
				FieldName:  parts[1],
				Direction:  dir,
			}
			b.SearchTables[sortField.TableAlias] = count + 1
			b.Sorts = append(b.Sorts, sortField)
		}
	}

	// fmt.Printf("\n#%+v", b.Filters)
	// fmt.Printf("\n#%+v\n\n", b.Sorts)
	return nil
}

// AllowedFiltersFieldsFromReflectionMap
// resets AllowedFilterFields
// the map takes two fields: string key and an interface
// the key maps to the table alias
// the interface is a struct with 'json', 'db' tags
// it uses reflection to determin the allowed fields
func (b *QueryBuilder) Build(paramString string, allowed map[string]interface{}) (where string, orderBy string, namedParamMap map[string]interface{}, err error) {
	namedParamMap = make(map[string]interface{})
	wheres := make(map[string][]Where)
	sb := []string{}

	if err := b.ParseParamString(paramString); err != nil {
		return "", "", nil, err
	}

	fieldsByTableAlias := make(map[string][]FilterField)
	for _, filter := range b.Filters {
		fieldsByTableAlias[filter.FieldName] = append(fieldsByTableAlias[filter.FieldName], filter)
	}

	sortsByTableAlias := make(map[string][]SortField)
	for _, sort := range b.Sorts {
		sortsByTableAlias[sort.FieldName] = append(sortsByTableAlias[sort.FieldName], sort)
	}

	for tableAlias, tableStruct := range allowed {
		rv := reflect.ValueOf(tableStruct)
		for i := 0; i < rv.NumField(); i++ {
			tag := rv.Type().Field(i).Tag.Get("db")
			if tag == "" {
				continue
			}

			fields, ok := fieldsByTableAlias[tag]
			if ok {
				for i, field := range fields {

					if field.TableAlias == tableAlias {
						switch field.Operator {
						case Between:
							// fmt.Println("field", field, "tag", tag, "tableAlias", tableAlias)
							// fmt.Println("Values", field.Values)

							if len(field.Values) == 2 {
								namedParam0 := fmt.Sprintf("filter_%s_%s_%d_0", field.TableAlias, field.FieldName, i)
								namedParamMap[namedParam0] = field.Values[0]
								namedParam1 := fmt.Sprintf("filter_%s_%s_%d_1", field.TableAlias, field.FieldName, i)
								namedParamMap[namedParam1] = field.Values[1]
								sqlString := fmt.Sprintf("%s.%s %s :%s AND :%s", field.TableAlias, field.FieldName, field.Operator.Convert(), namedParam0, namedParam1)
								combined := fmt.Sprintf("%s.%s", tableAlias, field.FieldName)
								wheres[combined] = append(wheres[combined], Where{
									CombinedName: combined,
									SqlString:    sqlString,
									Named:        namedParam0,
								})
							}

						case In, NotIn:
							var placeholders []string
							for j, val := range field.Values {
								namedParam := fmt.Sprintf("filter_%s_%s_%d_%d", field.TableAlias, field.FieldName, i, j)
								namedParamMap[namedParam] = val
								placeholders = append(placeholders, ":"+namedParam)
							}
							sqlString := fmt.Sprintf("%s.%s %s (%s)", field.TableAlias, field.FieldName, field.Operator.Convert(), strings.Join(placeholders, ", "))
							combined := fmt.Sprintf("%s.%s", tableAlias, field.FieldName)
							wheres[combined] = append(wheres[combined], Where{
								CombinedName: combined,
								SqlString:    sqlString,
							})

						case IsNull, IsNotNull:
							sqlString := fmt.Sprintf("%s.%s %s", field.TableAlias, field.FieldName, field.Operator.Convert())
							combined := fmt.Sprintf("%s.%s", tableAlias, field.FieldName)
							wheres[combined] = append(wheres[combined], Where{
								CombinedName: combined,
								SqlString:    sqlString,
							})

						default:
							namedParam := fmt.Sprintf("filter_%s_%s_%d", field.TableAlias, field.FieldName, i)
							namedParamMap[namedParam] = field.Value
							sqlString := fmt.Sprintf("%s.%s %s :%s", field.TableAlias, field.FieldName, field.Operator.Convert(), namedParam)
							combined := fmt.Sprintf("%s.%s", tableAlias, field.FieldName)
							wheres[combined] = append(wheres[combined], Where{
								CombinedName: combined,
								SqlString:    sqlString,
								Named:        namedParam,
								Operator:     field.Operator,
							})
						}
					}
				}
			}

			sorts, ok := sortsByTableAlias[tag]
			if ok {
				for _, sort := range sorts {
					if sort.TableAlias == tableAlias {
						sb = append(sb, fmt.Sprintf("%s.%s %s", tableAlias, sort.FieldName, sort.Direction))
					}
				}
			}
		}
	}

	where = b.AssembledWheres(wheres)
	orderBy = strings.Join(sb, ", ")
	if orderBy != "" {
		orderBy = fmt.Sprintf("ORDER BY %s", orderBy)
	}

	return where, orderBy, namedParamMap, err
}

func (b *QueryBuilder) AssembledWheres(whereMap map[string][]Where) string {
	where := []string{}
	orWhere := []string{}
	for _, ws := range whereMap {
		if len(ws) > 1 {
			orGroup := []string{}
			for _, w := range ws {
				orGroup = append(orGroup, w.SqlString)
			}
			where = append(where, "("+strings.Join(orGroup, " OR ")+")")
		} else {

			// fmt.Println("ws", ws[0])
			if ws[0].Operator == Or || ws[0].Operator == OrLike {
				orWhere = append(orWhere, ws[0].SqlString)
			} else {
				where = append(where, ws[0].SqlString)
			}
		}
	}

	out := strings.Join(where, " AND ")
	if len(orWhere) > 0 {
		out = out + "(" + strings.Join(orWhere, " OR ") + ")"
	}
	if out != "" {
		return fmt.Sprintf(" AND %s", out)
	}
	return ""
}

func BuildOrderBy(on string, allowedFields map[string]string) (orderBy string, err error) {
	if on == "" {
		return "", nil
	}

	var sb []string
	fields := strings.Split(strings.ToLower(on), ",")

	// fmt.Println("FIELDS: ", fields)

	for _, field := range fields {
		field = strings.TrimSpace(field)
		dir := "ASC"
		fieldName := field
		if isDesc := strings.HasPrefix(field, "-"); isDesc {
			dir = "DESC"
			fieldName = field[1:]
		}

		// fmt.Println("fieldName: ", fieldName)

		allowed := false
		for allowedField, tableName := range allowedFields {
			if fieldName == allowedField {
				sb = append(sb, fmt.Sprintf("%s.%s %s", tableName, fieldName, dir))
				allowed = true
				break
			}
		}

		if !allowed {
			return "", fmt.Errorf("error: %s is not allowed to be sorted on", fieldName)
		}
	}

	if len(sb) == 0 {
		return orderBy, err
	}

	orderBy = strings.Join(sb, ", ")
	if orderBy != "" {
		orderBy = fmt.Sprintf("ORDER BY %s", orderBy)
	}

	return orderBy, err
}
