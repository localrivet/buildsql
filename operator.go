package buildsql

type Operator string

const (
	Equal              Operator = "eq"
	NotEqual           Operator = "neq"
	Like               Operator = "like"
	NotLike            Operator = "nlike"
	LessThan           Operator = "lt"
	LessThanOrEqual    Operator = "lte"
	GreaterThan        Operator = "gt"
	GreaterThanOrEqual Operator = "gte"
	Betweeen           Operator = "btw"
)

func (o Operator) Convert() string {
	switch o {
	case Equal:
		return "="
	case NotEqual:
		return "!="
	case Like:
		return "LIKE"
	case NotLike:
		return "NOT LIKE"
	case LessThan:
		return "<"
	case LessThanOrEqual:
		return "<="
	case GreaterThan:
		return ">"
	case GreaterThanOrEqual:
		return ">="
	case Betweeen:
		return "BETWEEN"
	}
	return ""
}

func (o Operator) isLike() bool {
	return o == Like || o == NotLike
}
