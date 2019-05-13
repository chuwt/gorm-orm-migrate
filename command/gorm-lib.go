package command

import (
	"crypto/md5"
	"fmt"
	"github.com/jinzhu/gorm"
	"gopkg.in/satori/go.uuid.v1"
	"strings"
)

func dropTable(scope *gorm.Scope, tableName string) string {
	return fmt.Sprintf("DROP TABLE %v%s;", tableName, getScopeOption(scope))
}

func getScopeOption(scope *gorm.Scope) string {
	tableOptions, ok := scope.Get("gorm:table_options")
	if !ok {
		return ""
	} else {
		return " " + tableOptions.(string)
	}
}

func getIndex(scope *gorm.Scope) ([]string, []string) {
	var sqlString, downSqlString []string
	var indexes = map[string][]string{}
	var uniqueIndexes = map[string][]string{}
	for _, field := range scope.GetModelStruct().StructFields {
		// index
		if name, ok := field.TagSettingsGet("INDEX"); ok {
			names := strings.Split(name, ",")

			for _, name := range names {
				if name == "INDEX" || name == "" {
					name = scope.Dialect().BuildKeyName("idx", scope.TableName(), field.DBName)
				}
				if !scope.Dialect().HasIndex(scope.TableName(), name) {
					indexes[name] = append(indexes[name], field.DBName)
				}
			}
		}
		// unique index
		if name, ok := field.TagSettingsGet("UNIQUE_INDEX"); ok {
			names := strings.Split(name, ",")

			for _, name := range names {
				if name == "UNIQUE_INDEX" || name == "" {
					name = scope.Dialect().BuildKeyName("uix", scope.TableName(), field.DBName)
				}
				uniqueIndexes[name] = append(uniqueIndexes[name], field.DBName)
			}
		}
	}
	// index
	for name, columns := range indexes {
		if scope.Dialect().HasIndex(scope.TableName(), name) {
			continue
		}
		sqlString = append(sqlString, fmt.Sprintf("%s %v ON %v(%v);", "CREATE INDEX", name, scope.QuotedTableName(), strings.Join(columns, ", ")))
		downSqlString = append(downSqlString, fmt.Sprintf("DROP INDEX %v;", name))
	}

	for name, columns := range uniqueIndexes {
		if scope.Dialect().HasIndex(scope.TableName(), name) {
			continue
		}
		sqlString = append(sqlString, fmt.Sprintf("%s %v ON %v(%v)", "CREATE UNIQUE INDEX;", name, scope.QuotedTableName(), strings.Join(columns, ", ")))
		downSqlString = append(downSqlString, fmt.Sprintf("DROP UNIQUE INDEX %v;", name))
	}
	return sqlString, downSqlString
}

func createTable(scope *gorm.Scope, tableName string) string {
	var sqlString []string
	var tags []string
	var primaryKeys []string
	var primaryKeyInColumnType = false

	for _, field := range scope.GetModelStruct().StructFields {
		if field.IsNormal {
			sqlTag := scope.Dialect().DataTypeOf(field)
			if strings.Contains(strings.ToLower(sqlTag), "primary key") {
				primaryKeyInColumnType = true
			}

			tags = append(tags, scope.Quote(field.DBName)+" "+sqlTag)
		}

		if field.IsPrimaryKey {
			primaryKeys = append(primaryKeys, scope.Quote(field.DBName))
		}
	}
	var primaryKeyStr string
	if len(primaryKeys) > 0 && !primaryKeyInColumnType {
		primaryKeyStr = fmt.Sprintf(", PRIMARY KEY (%v)", strings.Join(primaryKeys, ","))
	}

	//scope.Raw(fmt.Sprintf("CREATE TABLE %v (%v %v)%s", scope.QuotedTableName(), strings.Join(tags, ","), primaryKeyStr, scope.getTableOptions())).Exec()
	sqlString = append(sqlString, fmt.Sprintf("CREATE TABLE %v (%v %v)%s", tableName, strings.Join(tags, ","), primaryKeyStr, getScopeOption(scope)))

	indexSqlString, _ := getIndex(scope)
	sqlString = append(sqlString, indexSqlString...)
	return strings.Join(sqlString, ";\n")
}

func generateVersionHash() string {
	uuidString := uuid.NewV4().String()
	Md5Inst := md5.New()
	Md5Inst.Write([]byte(uuidString))
	Result := Md5Inst.Sum([]byte(""))
	return fmt.Sprintf("%x", Result)
}
