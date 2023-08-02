package getuk

import (
	"fmt"
	"reflect"
	"strings"

	// gi "github.com/juliangruber/go-intersect"

	"gorm.io/gorm"
)

// Filters based on query parameters
func Filter(input interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		rt := reflect.TypeOf(input)
		if rt.Kind() != reflect.Struct {
			return db
		}

		iV := reflect.ValueOf(input) // input value
		if iV.Kind() != reflect.Struct {
			panic("input must be a struct")
		}

		iT := iV.Type() // input type
		for i := 0; i < iV.NumField(); i++ {
			iTF := iT.Field(i) // input type of the current field
			opt := iTF.Name
			if len(iTF.Name) >= 4 {
				opt = iTF.Name[len(iTF.Name)-4:]
			}

			// skip option and pagination related
			if opt == "_Opt" || iTF.Name == "Page" || iTF.Name == "PageSize" || iTF.Name == "NoPagination" {
				continue
			}

			// proceed value
			iVF := iV.Field(i) // input value of the current field
			for iVF.Kind() == reflect.Ptr {
				iVF = iVF.Elem()
			}

			// check field value
			lastITF := reflect.TypeOf(iVF)
			if !iVF.IsValid() || iVF.IsZero() || (lastITF.Kind() == reflect.Ptr && iVF.IsNil()) {
				continue
			}

			// check opt
			vOpt := "eq"
			o := iV.FieldByName(iTF.Name + "_Opt") // option
			if o.IsValid() && o.Kind() == reflect.Ptr && o.Elem().IsValid() {
				o = o.Elem()
				vOpt = o.Interface().(string)
			}
			opts := []string{"eq", "lt", "gt", "lte", "gte", "ne", "left", "mid", "right", "between", "in"}
			if ok := stringInSlice(vOpt, opts); !ok {
				db.AddError(fmt.Errorf("field %s: opt undefined", iTF.Name))
			}

			// check source if avaibale
			refSource := iTF.Tag.Get("refsource")
			if refSource != "" {
				refSource = strings.Replace(refSource, ".", "\".\"", -1)
			} else {
				refSource = iTF.Name
			}

			if iTF.Type.String() == "*[]string" {
				vOpt = "in"
			}

			// add where query
			whereString, value := optionString(refSource, vOpt, iVF.Interface())
			if vOpt != "between" {
				db.Where(whereString, value)
			} else {
				valueString := iVF.String()
				values := strings.Split(valueString, ",")
				if len(values) == 2 {
					db.Where(whereString, values[0], values[1])
				}
			}
		}

		return db
	}
}

// Paginate based on query parameters
func Paginate(input interface{}, p *Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		iV := reflect.ValueOf(input) // input value
		if iV.Kind() != reflect.Struct {
			panic("input must be a struct")
		}
		// field pagination
		fP := iV.FieldByName("Page")
		fPS := iV.FieldByName("PageSize")
		fNP := iV.FieldByName("NoPagination")
		if fNP.IsValid() {
			if fNP.Type().Kind() == reflect.Bool && bool(fNP.Interface().(bool)) {
				return db
			}
		}
		if fP.IsValid() {
			if fP.Type().Kind() == reflect.Int {
				p.Page = fP.Interface().(int)
			} else if fP.Type().Kind() == reflect.Int64 {
				p.Page = int(fP.Interface().(int64))
			} else {
				p.Page = 1
			}
			if p.Page <= 0 {
				p.Page = 1
			}
		} else {
			p.Page = 1
		}
		if fPS.IsValid() {
			if fPS.Type().Kind() == reflect.Int {
				p.PageSize = fPS.Interface().(int)
			} else if fPS.Type().Kind() == reflect.Int64 {
				p.PageSize = int(fPS.Interface().(int64))
			} else {
				p.PageSize = 10
			}
			if p.PageSize <= 5 {
				p.PageSize = 10
			}
		} else {
			p.PageSize = 10
		}
		offset := (p.Page - 1) * p.PageSize
		return db.Offset(offset).Limit(p.PageSize)
	}
}
