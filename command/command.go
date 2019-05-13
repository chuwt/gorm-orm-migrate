package command

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	MigrateDir = "./migrations"
)

var (
	Command     string
	Arg         string
	MigrateList []interface{}
)

func init() {
	if len(os.Args) == 1 {
		Usage()
	}
	Command = os.Args[1]
	if len(os.Args) == 3 {
		Arg = os.Args[2]
	} else {
		Arg = ""
	}
}

func AddTable(values ...interface{}) {
	for _, value := range values {
		MigrateList = append(MigrateList, value)
	}

}

func AddCommand(db *gorm.DB) {
	if Command == "init" {
		Init(db)
	} else {
		if ok := InitCheck(db); !ok {
			fmt.Println("plz init first")
			return
		}
		switch Command {
		case "migrate":
			Migrate(db)
			break
		case "upgrade":
			Upgrade(db)
			break
		case "downgrade":
			Downgrade(db)
			break
		case "head":
			version := Head(db)
			dbVersion, _ := GetVersion(db)
			fmt.Println(fmt.Sprintf("head: [%s]\ncurrent: [%s]", version, dbVersion.Version))
			break
		default:
			Usage()
		}
	}
}

func Usage() {
	fmt.Println("usage")
}

func Init(db *gorm.DB) {
	err := os.Mkdir(MigrateDir, os.ModePerm)
	if err != nil {
		fmt.Printf("migrations 文件夹创建失败：![%v]\n", err)
		return
	} else {
		// 生成数据库
		fmt.Printf("migrations 文件夹创建成功\n")
		if err := db.AutoMigrate(&SQLVersion{}).Error; err != nil {
			fmt.Printf("创建数据库失败：![%v]\n", err)
		}
	}
}

func InitCheck(db *gorm.DB) bool {
	// 判断文件夹是否存在
	s, err := os.Stat(MigrateDir)
	if err != nil {
		return false
	}
	if ok := s.IsDir(); !ok {
		return false
	}
	// 判断数据库是否存在
	if ok := db.HasTable(&SQLVersion{}); !ok {
		return false
	}
	return true
}

func writeFile(version, downVersion, upString, downString string) {
	// 生成文件
	filename := fmt.Sprintf("%s.sql", version)
	f, err := os.OpenFile(fmt.Sprintf("./%s/%s", MigrateDir, filename), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("生成文件错误：![%v]\n", err)
		return
	}
	f.WriteString(fmt.Sprintf(Template, time.Now().String(), version, downVersion, upString, downString))
	f.Close()
}

type multitable interface {
	MultiTable() int
}

func Migrate(db *gorm.DB) {

	head := Head(db)

	// 获取数据库的version
	var downVersion string
	var upString, downString []string

	dbVersion, _ := GetVersion(db)
	if dbVersion == nil {
		downVersion = ""
	} else {
		downVersion = dbVersion.Version
	}
	if head != downVersion {
		fmt.Println("plz upgrade first")
		return
	}
	version := generateVersionHash()

	for _, value := range MigrateList {
		scope := db.NewScope(value)
		tableName := scope.TableName()
		tableNameList := make([]string, 0)
		if z, ok := scope.Value.(multitable); ok {
			for i:=0 ; i<z.MultiTable(); i++ {
				tableNameList = append(tableNameList, fmt.Sprintf("%s_%02d", tableName, i))
			}
		} else {
			tableNameList = append(tableNameList, tableName)
		}
		fmt.Println(tableNameList)
		for _, tableName := range tableNameList {
			if !scope.Dialect().HasTable(tableName) {
				// 表不存在
				// upgrade
				upString = append(upString, createTable(scope, tableName))
				// downgrade
				downString = append(downString, dropTable(scope, tableName))
			} else {
				// 表存在
				//var is
				for _, field := range db.NewScope(value).GetModelStruct().StructFields {
					// 去除relationship
					if !scope.Dialect().HasColumn(tableName, field.DBName) {
						if field.IsNormal {
							sqlTag := scope.Dialect().DataTypeOf(field)
							sqlString := fmt.Sprintf("ALTER TABLE %v ADD %v %v", tableName, scope.Quote(field.DBName), sqlTag)
							downSqlString := fmt.Sprintf("ALTER TABLE %v DROP COLUMN %v", tableName, scope.Quote(field.DBName))
							upString = append(upString, sqlString)
							downString = append(downString, downSqlString)
						}
					}
					// 建立依赖表
					//scope.createJoinTable(field)
				}
				// index
				indexSqlString, indexDownSqlString := getIndex(scope)
				upString = append(upString, indexSqlString...)
				downString = append(downString, indexDownSqlString...)
			}
			//
		}
	}
	if upString != nil {
		fmt.Println(fmt.Sprintf("head: [%s]", version))
		writeFile(version, downVersion, strings.Join(upString, ";\n"), strings.Join(downString, "\n"))
	} else {
		fmt.Println("nothing changed")
	}
}

func manager(db *gorm.DB, op string) {
	// 获取上一次upgrade的版本
	var dbVersion string
	var head, downVersion string
	dbVersionMeta, _ := GetVersion(db)
	if dbVersionMeta == nil {
		dbVersion = ""
	} else {
		dbVersion = dbVersionMeta.Version
	}
	// 读取文件夹下的所有version
	for _, version := range getMigrateList() {
		// 查看文件
		fileBytes := getFileString(fmt.Sprintf("%s/%s.sql", MigrateDir, version))
		if fileBytes == nil {
			continue
		}
		// 解析
		versionRe, _ := regexp.Compile(`reversion:[\w]+`)
		downVersionRe, _ := regexp.Compile(`down_revision:[\w]+`)

		// 最新版本
		headList := string(versionRe.Find(fileBytes))
		// 当前版本
		downVersionList := string(downVersionRe.Find(fileBytes))
		head = strings.Split(headList, ":")[1]
		if head == dbVersion {
			fmt.Println("has upgraded")
			return
		}
		if dbVersion == "" {
			// run
			var sqlRe *regexp.Regexp
			if op == "upgrade" {
				sqlRe, _ = regexp.Compile(`-- upgrade\n[\w\W]+\n-- end upgrade`)
			} else {
				sqlRe, _ = regexp.Compile(`-- downgrade\n[\w\W]+\n-- end downgrade`)
			}

			sqlBytes := sqlRe.Find(fileBytes)
			db.Exec(string(sqlBytes))
			UpdateVersion(db, head)
		} else {
			if downVersionList != "" {
				downVersion = strings.Split(downVersionList, ":")[1]
				if downVersion == dbVersion {
					// run
					var sqlRe *regexp.Regexp
					if op == "upgrade" {
						sqlRe, _ = regexp.Compile(`-- upgrade\n[\w\W]+\n-- end upgrade`)
					} else {
						sqlRe, _ = regexp.Compile(`-- downgrade\n[\w\W]+\n-- end downgrade`)
					}
					sqlBytes := sqlRe.Find(fileBytes)
					db.Exec(string(sqlBytes))
					UpdateVersion(db, head)
				}
			} else {
				fmt.Println("has upgraded")
			}
		}
	}
}

func Upgrade(db *gorm.DB) {
	manager(db, "upgrade")
}

func Downgrade(db *gorm.DB) {
	var downVersion string
	// 获取当前版本
	dbVersion, _ := GetVersion(db)
	if dbVersion.Version == "" {
		fmt.Println("plz migrate first")
		return
	}
	fileBytes := getFileString(fmt.Sprintf("%s/%s.sql", MigrateDir, dbVersion.Version))

	downVersionRe, _ := regexp.Compile(`down_revision:[\w]+`)
	// 当前版本
	downVersionList := string(downVersionRe.Find(fileBytes))

	sqlRe, _ := regexp.Compile(`-- downgrade\n[\w\W]+\n-- end downgrade`)
	sqlBytes := sqlRe.Find(fileBytes)
	db.Exec(string(sqlBytes))

	if downVersionList == "" {
		downVersion = ""
	} else {
		downVersion = strings.Split(downVersionList, ":")[1]
	}
	UpdateVersion(db, downVersion)
}

func getFileString(fileName string) []byte {
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("读取文件错误：![%v]\n", err)
		return nil
	}
	return fileBytes
}

func getMigrateList() []string {
	versionList := make([]string, 0)
	// 读取文件夹下的所有version
	files, _ := ioutil.ReadDir(MigrateDir)
	for _, f := range files {
		fileNameList := strings.Split(f.Name(), ".")
		if len(fileNameList) == 1 {
			continue
		} else if fileNameList[1] != "sql" {
			continue
		}
		versionList = append(versionList, fileNameList[0])
	}
	return versionList
}

func Head(db *gorm.DB) string {
	// 获取上一次upgrade的版本
	var dbVersion string
	var head, downVersion string
	dbVersionMeta, _ := GetVersion(db)
	if dbVersionMeta == nil {
		dbVersion = ""
	} else {
		dbVersion = dbVersionMeta.Version
	}
	// 读取文件夹下的所有version
	for _, version := range getMigrateList() {
		// 查看文件
		fileBytes := getFileString(fmt.Sprintf("%s/%s.sql", MigrateDir, version))
		if fileBytes == nil {
			continue
		}
		// 解析
		versionRe, _ := regexp.Compile(`reversion:[\w]+`)
		downVersionRe, _ := regexp.Compile(`down_revision:[\w]+`)

		// 最新版本
		headList := string(versionRe.Find(fileBytes))
		// 当前版本
		downVersionList := string(downVersionRe.Find(fileBytes))
		head = strings.Split(headList, ":")[1]
		if dbVersion == "" {
			return head
		} else {
			if downVersionList != "" {
				downVersion = strings.Split(downVersionList, ":")[1]
				if downVersion == dbVersion {
					return head
				}
			}
		}
	}
	if dbVersion != "" {
		return head
	}
	return ""
}

type SQLVersion struct {
	Version string `gorm:"type:varchar(128)"`
}

func (*SQLVersion) TableName() string {
	return "sql_version"
}

func GetVersion(db *gorm.DB) (*SQLVersion, error) {
	version := new(SQLVersion)
	if err := db.First(version).Error; err != nil {
		return nil, err
	}
	return version, nil
}

func UpdateVersion(db *gorm.DB, versionString string) bool {
	version := new(SQLVersion)
	if err := db.First(version, "version != ?", "").Error; err != nil && err != gorm.ErrRecordNotFound {
		return false
	}
	var oldVersion string
	if version == nil || version.Version == ""{
		oldVersion = ""
		if err := db.Create(&SQLVersion{Version: versionString}).Error; err != nil {
			fmt.Println("create version error")
			return false
		}
	} else {
		oldVersion = version.Version
		if err := db.Model(version).UpdateColumn("version", versionString).Error; err != nil {
			fmt.Println("update version error")
			return false
		}
	}

	fmt.Println(fmt.Sprintf("head: [%s] -> [%s]", oldVersion, versionString))
	return true
}
