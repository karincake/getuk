package getuk

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// Filters based on query parameters
func Filter(input interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// rt := reflect.TypeOf(input)
		// if rt.Kind() != reflect.Struct {
		// 	return db
		// }

		iV := reflect.ValueOf(input) // input value
		for iV.Kind() == reflect.Pointer {
			iV = iV.Elem()
		}
		if iV.Kind() != reflect.Struct {
			panic("input must be a struct")
		}

		tableNameEscapeChar := "`"
		dialectorName := db.Dialector.Name()
		if dialectorName == "postgres" {
			tableNameEscapeChar = "\""
		}

		iT := iV.Type() // input type
		for i := 0; i < iV.NumField(); i++ {
			iTF := iT.Field(i) // input type of the current field
			opt := iTF.Name
			if len(iTF.Name) >= 4 {
				opt = iTF.Name[len(iTF.Name)-4:]
			}

			// skip option and pagination related
			if opt == "_Opt" || iTF.Name == "PageNumber" || iTF.Name == "PageSize" || iTF.Name == "PageNoLimit" {
				continue
			}

			// skip
			raw := false
			skip := false
			refSource := iTF.Name
			ghTagsRaw := iTF.Tag.Get("gormhelper")
			ghTags := strings.Split(ghTagsRaw, ";")
			for idx := range ghTags {
				vals := strings.Split(ghTags[idx], "=")
				if len(vals) == 2 {
					if vals[0] == "refsource" {
						refSource = vals[1]
						break
					}
				} else {
					if vals[0] == "skip" {
						skip = true
						break
					} else if vals[0] == "raw" {
						raw = true
					}
				}
			}
			if skip {
				continue
			}

			// proceed value
			iVF := iV.Field(i) // input value of the current field
			for iVF.Kind() == reflect.Ptr {
				iVF = iVF.Elem()
			}

			// check field value
			if !iVF.IsValid() || iVF.IsZero() {
				continue
			}

			// check opt
			vOpt := "eq"
			o := iV.FieldByName(iTF.Name + "_Opt") // option
			if o.IsValid() {
				if o.Kind() == reflect.Ptr && o.Elem().IsValid() {
					o = o.Elem()
					vOpt = o.Interface().(string)
				} else if o.Kind() != reflect.Ptr {
					vOpt = o.Interface().(string)
				} else {
					// nothing to do if its invalid pointer
				}
			}
			opts := []string{"eq", "lt", "gt", "lte", "gte", "ne", "left", "mid", "right", "between", "in-string", "in-int", "in-float"}
			if ok := stringInSlice(vOpt, opts); !ok {
				db.AddError(fmt.Errorf("field %s: opt undefined", iTF.Name))
			}

			if iTF.Type.String() == "*[]string" {
				vOpt = "in"
			}

			// add where query
			whereString, value := optionString(refSource, vOpt, tableNameEscapeChar, iVF.Interface(), raw)
			if vOpt == "between" {
				theType := iVF.Type().String()
				if theType == "string" {
					valueString := iVF.String()
					values := strings.Split(valueString, "|")
					if len(values) == 2 {
						db.Where(whereString, values[0], values[1])
					}
				} else if iVF.Kind() == reflect.Slice {
					if iVF.Len() == 2 {
						db.Where(whereString, iVF.Index(0), iVF.Index(1))
					}
				}
			} else if vOpt == "in-string" {
				db.Where(whereString, strings.Split(value.(string), ","))
			} else if vOpt == "in-int" {
				strNumbers := strings.Split(value.(string), ",")
				numbers := make([]int, len(strNumbers))
				for idx := range strNumbers {
					number, err := strconv.Atoi(strNumbers[idx])
					if err != nil {
						panic("input must be a struct")
					}
					numbers = append(numbers, number)
				}
				db.Where(whereString, numbers)
			} else if vOpt == "in-float" {
				strNumbers := strings.Split(value.(string), ",")
				numbers := make([]float64, len(strNumbers))
				for idx := range strNumbers {
					number, err := strconv.ParseFloat(strNumbers[idx], 64)
					if err != nil {
						panic("input must be a struct")
					}
					numbers = append(numbers, number)
				}
				db.Where(whereString, numbers)
			} else {
				db.Where(whereString, value)
			}
		}

		return db
	}
}

// Paginate based on query parameters
func Paginate(input interface{}, p *Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		iV := reflect.ValueOf(input) // input value
		for iV.Kind() == reflect.Pointer {
			iV = iV.Elem()
		}
		if iV.Kind() != reflect.Struct {
			panic("input must be a struct")
		}
		// field pagination
		fP := iV.FieldByName("PageNumber")
		fPS := iV.FieldByName("PageSize")
		fNP := iV.FieldByName("PageNoLimit")
		if fNP.IsValid() {
			if fNP.Type().Kind() == reflect.Bool && bool(fNP.Interface().(bool)) {
				return db
			}
		}
		if fP.IsValid() {
			myKind := fP.Type().Kind()
			if myKind == reflect.Int {
				p.PageNumber = fP.Interface().(int)
			} else {
				panic("property 'PageNumber' must have int type ")
			}
			if p.PageNumber <= 0 {
				p.PageNumber = 1
			}
		} else {
			p.PageNumber = 1
		}
		if fPS.IsValid() {
			myKind := fPS.Type().Kind()
			if myKind == reflect.Int {
				p.PageSize = fPS.Interface().(int)
			} else {
				panic("property 'PageSize' must have int type ")
			}
			if p.PageSize >= 1000 {
				p.PageSize = 1000
			}
			if p.PageSize <= 0 {
				p.PageSize = 10
			}
		} else {
			p.PageSize = 10
		}
		offset := (p.PageNumber - 1) * int(p.PageSize)
		return db.Offset(offset).Limit(p.PageSize)
	}
}

// Flatten a reference with tricky workaround due to Gorm's nature
// that doesn't allow multiple times call for Select() function.
// The trick is to use a string pointer that contains `SELECT` clause
// for main table source. It will passed to the function to be updated.
// In case an error occured while calling `COUNT`, please avoid calling `SELECT`
// prior the `COUNT`
func FlatJoin(selectStr *string, opt FlatJoinOpt) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if selectStr == nil {
			panic("variable selectStr must be a valid pointer address")
		}
		if opt.Ref == "" {
			panic("opt's 'Ref' is for reference table to be joined, can not be empty")
		}
		if opt.Src == "" {
			panic("opt's 'Src' is for source table, can not be empty")
		}
		if len(opt.Cols) == 0 {
			panic("opt's `Cols` is for columns to be selected from reference table, can not be empty")
		}

		dialector := db.Dialector.Name()
		qt := "`"
		if dialector == "postgres" {
			qt = "\""
		}

		if opt.Mode == "" {
			opt.Mode = JMInner
		}

		fRefCol := "Id"
		if opt.RefCol != "" {
			fRefCol = opt.RefCol
		}

		fSrcFkCol := opt.Ref + "_" + fRefCol
		if opt.SrcFkCol != "" {
			fSrcFkCol = opt.SrcFkCol
		}

		fPrefix := opt.Ref + "_"
		if opt.Prefix != "" && opt.Prefix != "{NOPREFIX}" {
			fPrefix = opt.Prefix
		} else if opt.Prefix == "{NOPREFIX}" {
			fPrefix = ""
		}

		fSelectStr := ""
		for idx := range opt.Cols {
			fSelectStr += fSelectStr + fmt.Sprintf("%v%v%v.%v%v%v %v%v%v, ", qt, opt.Ref, qt, qt, opt.Cols[idx], qt, qt, fPrefix+opt.Cols[idx], qt)
		}
		fSelectStr = fSelectStr[:len(fSelectStr)-2]

		if *selectStr == "" {
			*selectStr = fSelectStr
		} else {
			fSelectStr = *selectStr + ", " + fSelectStr
			*selectStr = fSelectStr
		}

		clause := ""
		if opt.Clause != "" {
			clause = " AND " + opt.Clause
		}

		return db.Joins(fmt.Sprintf("%v %v%v%v ON %v%v%v.%v%v%v = %v%v%v.%v%v%v %v", opt.Mode, qt, opt.Ref, qt, qt, opt.Src, qt, qt, fSrcFkCol, qt, qt, opt.Ref, qt, qt, fRefCol, qt, clause))
	}
}

// Procedural version of FlatJoinProc
func FlatJoinProc(db *gorm.DB, selectStr *string, opt FlatJoinOpt) *gorm.DB { // *gorm.DB
	if selectStr == nil {
		panic("variable selectStr must be a valid pointer address")
	}
	if opt.Ref == "" {
		panic("opt's 'Ref' can not be an empty string")
	}
	if opt.Src == "" {
		panic("opt's 'Src' can not be an empty string")
	}
	if len(opt.Cols) == 0 {
		panic("opt's `Cols` can not be an empty array")
	}
	dialector := db.Dialector.Name()

	qt := "`"
	if dialector == "postgres" {
		qt = "\""
	}

	fRefCol := "Id"
	if opt.RefCol != "" {
		fRefCol = opt.RefCol
	}

	fSrcFkCol := opt.Ref + "_" + fRefCol
	if opt.SrcFkCol != "" {
		fSrcFkCol = opt.SrcFkCol
	}

	fPrefix := opt.Ref + "_"
	if opt.Prefix != "" && opt.Prefix != "{NOPREFIX}" {
		fPrefix = opt.Prefix
	} else if opt.Prefix == "{NOPREFIX}" {
		fPrefix = ""
	}

	fSelectStr := ""
	for idx := range opt.Cols {
		fSelectStr += fSelectStr + fmt.Sprintf("%v%v%v.%v%v%v %v%v%v, ", qt, opt.Ref, qt, qt, opt.Cols[idx], qt, qt, fPrefix+opt.Cols[idx], qt)
	}
	fSelectStr = fSelectStr[:len(fSelectStr)-2]

	if *selectStr == "" {
		*selectStr = fSelectStr
	} else {
		fSelectStr = *selectStr + ", " + fSelectStr
		*selectStr = fSelectStr
	}

	clause := ""
	if opt.Clause != "" {
		clause = " AND " + clause
	}

	return db.Joins(fmt.Sprintf("%v %v%v%v ON %v%v%v.%v%v%v = %v%v%v.%v%v%v %v", opt.Mode, qt, opt.Ref, qt, qt, opt.Src, qt, qt, fSrcFkCol, qt, qt, opt.Ref, qt, qt, fRefCol, qt, clause))
}
