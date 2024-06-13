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

func (o Operator) IsLike() bool {
	return (o == Like || o == OrLike || o == ILike || o == OrILike) || (o == NotLike || o == NotILike)
}

func (o Operator) IsBetween() bool {
	return o == Between
}

func (o Operator) IsIn() bool {
	return o == In
}

func (o Operator) IsNotIn() bool {
	return o == NotIn
}

func (o Operator) IsNull() bool {
	return o == IsNull || o == IsNotNull
}

// to string
func (o Operator) String() string {
	return string(o)
}
