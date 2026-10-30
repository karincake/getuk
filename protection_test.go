package getuk

import (
	"strings"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// fakeDialector produces SQL without a database connection (used with DryRun)
type fakeDialector struct{ name string }

func (d fakeDialector) Name() string { return d.name }
func (d fakeDialector) Initialize(db *gorm.DB) error {
	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{})
	return nil
}
func (d fakeDialector) Migrator(db *gorm.DB) gorm.Migrator                          { return nil }
func (d fakeDialector) DataTypeOf(*schema.Field) string                             { return "" }
func (d fakeDialector) DefaultValueOf(*schema.Field) clause.Expression              { return nil }
func (d fakeDialector) BindVarTo(w clause.Writer, _ *gorm.Statement, _ interface{}) { w.WriteByte('?') }
func (d fakeDialector) QuoteTo(w clause.Writer, s string) {
	q := "`"
	if d.name == "postgres" {
		q = `"`
	}
	w.WriteString(q + s + q)
}
func (d fakeDialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `'`, vars...)
}

type row struct{ Id uint }

type searchInput struct {
	Search string `gormhelper:"searchColumns=Patient.Number,Person.Name"`
}

func dryRun(t *testing.T, dialect string, scope func(*gorm.DB) *gorm.DB) *gorm.Statement {
	t.Helper()
	db, err := gorm.Open(fakeDialector{dialect}, &gorm.Config{DryRun: true, Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	var rows []row
	return db.Table("Row").Scopes(scope).Find(&rows).Statement
}

func withProtection(t *testing.T, enabled bool) {
	t.Helper()
	prev := SQLInjectionProtection()
	SetSQLInjectionProtection(enabled)
	t.Cleanup(func() { SetSQLInjectionProtection(prev) })
}

const payload = "da'e' OR SLEEP(5) OR '"

func TestProtectionEnabledByDefault(t *testing.T) {
	if !SQLInjectionProtection() {
		t.Fatal("protection should be enabled by default")
	}
}

func TestSearchBindsKeyword(t *testing.T) {
	withProtection(t, true)
	for _, dialect := range []string{"mysql", "postgres"} {
		stmt := dryRun(t, dialect, Filter(searchInput{Search: payload}))
		sql := stmt.SQL.String()
		if strings.Contains(sql, "SLEEP") || strings.Contains(sql, "'") {
			t.Errorf("%s: keyword leaked into SQL: %s", dialect, sql)
		}
		if len(stmt.Vars) != 2 {
			t.Fatalf("%s: want 2 bound vars, got %v", dialect, stmt.Vars)
		}
		want := "%" + payload + "%"
		if dialect == "mysql" {
			want = strings.ToLower(want)
		}
		for _, v := range stmt.Vars {
			if v != want {
				t.Errorf("%s: bound var = %q, want %q", dialect, v, want)
			}
		}
	}
}

func TestSearchLegacyInterpolates(t *testing.T) {
	withProtection(t, false)
	stmt := dryRun(t, "mysql", Filter(searchInput{Search: "abc"}))
	sql := stmt.SQL.String()
	if !strings.Contains(sql, "LOWER(Patient.Number) LIKE '%abc%' OR LOWER(Person.Name) LIKE '%abc%'") {
		t.Errorf("unexpected legacy SQL: %s", sql)
	}
	if len(stmt.Vars) != 0 {
		t.Errorf("legacy mode should not bind vars, got %v", stmt.Vars)
	}
}

func TestSearchCombinesWithOtherFilters(t *testing.T) {
	withProtection(t, true)
	type input struct {
		Status string
		Search string `gormhelper:"searchColumns=Patient.Number,Person.Name"`
	}
	stmt := dryRun(t, "mysql", Filter(input{Status: "new", Search: "x"}))
	sql := stmt.SQL.String()
	if !strings.Contains(sql, "(LOWER(Patient.Number) LIKE ? OR LOWER(Person.Name) LIKE ?)") {
		t.Errorf("OR group must be parenthesised next to other conditions: %s", sql)
	}
}

func TestSortRejectsUnsafeField(t *testing.T) {
	withProtection(t, true)
	stmt := dryRun(t, "mysql", Sort("name:asc,x`, (SELECT SLEEP(5)), `y:desc"))
	sql := stmt.SQL.String()
	if strings.Contains(sql, "SLEEP") {
		t.Errorf("unsafe sort field reached SQL: %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY `Name` ASC") {
		t.Errorf("safe sort field missing: %s", sql)
	}
}

func TestSortLegacyKeepsField(t *testing.T) {
	withProtection(t, false)
	stmt := dryRun(t, "mysql", Sort("na`me:asc"))
	if !strings.Contains(stmt.SQL.String(), "`Na`me` ASC") {
		t.Errorf("legacy sort should pass the field through: %s", stmt.SQL.String())
	}
}
