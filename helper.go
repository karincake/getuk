package getuk

import (
	"fmt"
	"strings"
	"unicode"
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

	if str, ok := valueFinal.(string); ok {
		valueFinal = strings.TrimSpace(str)
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

func splitString(s string, sep string) []string {
	return strings.Split(s, sep)
}

func searchhQueryBuilder(searchColumns []string, search, dialector string) string {
	switch dialector {
	case "postgres":
		if len(searchColumns) == 1 {
			return fmt.Sprintf("\"%s\" ILIKE '%%%s%%'", searchColumns[0], search)
		}

		var parts []string
		for _, col := range searchColumns {
			parts = append(parts, fmt.Sprintf("\"%s\" ILIKE '%%%s%%'", col, search))
		}
		return strings.Join(parts, " OR ")
	case "mysql":
		search = strings.ToLower(search)
		if len(searchColumns) == 1 {
			return fmt.Sprintf("LOWER(%s) LIKE '%%%s%%'", searchColumns[0], search)
		}

		var parts []string
		for _, col := range searchColumns {
			parts = append(parts, fmt.Sprintf("LOWER(%s) LIKE '%%%s%%'", col, search))
		}
		return strings.Join(parts, " OR ")
	default:
		return ""
	}
}

func normalizeColumnName(input string) string {
	if input == "" {
		return input
	}

	input = strings.ReplaceAll(input, "-", "_")

	runes := []rune(input)
	var out []rune
	upperNext := true
	for _, r := range runes {
		if r == '_' {
			out = append(out, r)
			upperNext = true
			continue
		}
		if upperNext {
			out = append(out, unicode.ToUpper(r))
			upperNext = false
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

func kebabToPascal(input string) string {
	parts := strings.Split(input, "-")
	for i, p := range parts {
		if len(p) > 0 {
			r := []rune(p)
			r[0] = []rune(strings.ToUpper(string(r[0])))[0]
			parts[i] = string(r)
		}
	}
	return strings.Join(parts, "")
}
