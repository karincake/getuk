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
//
// allowed option, the default is without like and between
//
//	[]string{"=", "lt", "gt", "lte", "gte", "ne", "left", "mid", "right"}
func optionString(col, option, tableNameEscapeChar string, value interface{}) (whereString string, valueFinal interface{}) {
	cols := strings.Split(col, ".")
	for idx := range cols {
		cols[idx] = tableNameEscapeChar + cols[idx] + tableNameEscapeChar
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
		whereString = strings.Join(cols, ".") + " BETWEEN ? AND ?"
	} else if symbol == "IN" {
		whereString = strings.Join(cols, ".") + " IN (?)"
	} else {
		whereString = strings.Join(cols, ".") + " = ?"
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
