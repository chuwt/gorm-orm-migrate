package model

import (
    "github.com/jinzhu/gorm"
    "gopkg.in/satori/go.uuid.v1"
    _ "github.com/jinzhu/gorm/dialects/postgres"
    )

var (
    DB *gorm.DB
)

func InitPostgresDB() {
    var err error
    db, err := gorm.Open("postgres", "postgres://postgres@127.0.0.1:5432/test?sslmode=disable")

    if err != nil {
        panic(err)
    }
    DB = db.Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false)
    DB.LogMode(true)
    DB.SingularTable(true)

    DB.DB().SetMaxIdleConns(10)
    DB.DB().SetMaxOpenConns(10)
}

func init() {

    InitPostgresDB()
    gorm.DefaultCallback.Create().Before("gorm:before_create").Register("generate_uuid", GenerateUUIDForRecord)
}

func GenerateUUIDForRecord(scope *gorm.Scope) {
    if !scope.HasError() && scope.PrimaryKeyZero() {
        scope.SetColumn("ID", uuid.NewV4().String())
    }
}