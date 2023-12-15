package getuk

import (
	"fmt"
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
	case "in":
		symbol = "IN"
	}

	if symbol == "BETWEEN" {
		whereString = fmt.Sprintf("%s%s%s %s ? AND ?", tableNameEscapeChar, col, tableNameEscapeChar, symbol)
	} else if symbol == "IN" {
		whereString = fmt.Sprintf("%s%s%s %s ?", tableNameEscapeChar, col, tableNameEscapeChar, symbol)
	} else {
		whereString = fmt.Sprintf("%s%s%s %s ?", tableNameEscapeChar, col, tableNameEscapeChar, symbol)
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
