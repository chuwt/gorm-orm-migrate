package model

import "fmt"


type TestApp struct {
    TableIndex int `gorm:"-"`
    Name string `gorm:"type:varchar(32)"`
    BName string `gorm:"type:varchar(32)"`
}

func (testApp *TestApp) TableName() string {
    return fmt.Sprintf("test_app_%02d", testApp.TableIndex)
}

/* 用于多表
    {表名}_01 - {表名}_10
 */
func (*TestApp) MultiTable() int {
    return 10
}