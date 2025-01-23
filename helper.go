package getuk

import (
	"fmt"
	"strings"
)

// just local functions, avoid import if it is simple

// check if stnig availble in an array
func stringInSlice(s string, a []string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}

// create where string e.g "Column" <= ? for where gorm
func optionString(col, option, tableNameEscapeChar string, value interface{}, rawStatus bool) (whereString string, valueFinal interface{}) {
	var refSource string
	if rawStatus {
		refSource = col
	} else {
		cols := strings.Split(col, ".")
		for idx := range cols {
			cols[idx] = tableNameEscapeChar + cols[idx] + tableNameEscapeChar
		}
		refSource = strings.Join(cols, ".")
	}

	symbol := "="
	valueFinal = value
	switch option {
	case "eq":
		symbol = "="
	case "left":
		symbol = "LIKE"
		valueFinal = fmt.Sprintf("%v%%", value)
	case "mid":
		symbol = "LIKE"
		valueFinal = fmt.Sprintf("%%%v%%", value)
	case "right":
		symbol = "LIKE"
		valueFinal = fmt.Sprintf("%%%v", value)
	case "lt":
		symbol = "<"
	case "gt":
		symbol = ">"
	case "lte":
		symbol = "<="
	case "gte":
		symbol = ">="
	case "ne":
		symbol = "<>"
	case "between":
		symbol = "BETWEEN"
	case "in-string":
		symbol = "IN"
	case "in-int":
		symbol = "IN"
	case "in-float":
		symbol = "IN"
	}

	if symbol == "BETWEEN" {
		whereString = refSource + " BETWEEN ? AND ?"
	} else if symbol == "IN" {
		whereString = refSource + " IN (?)"
	} else if symbol == "LIKE" {
		whereString = refSource + " LIKE (?)"
	} else {
		whereString = refSource + " = ?"
	}
	return
}

// reflect field value or nil
// func refOperatorVal(iV reflect.Value, iTF reflect.StructField) string {
// 	vOpt := "="
// 	o := iV.FieldByName(iTF.Name + "_Opt") // option
// 	if o.Interface() != nil {
// 		vOpt = o.String()
// 	}
// 	return vOpt
// }
