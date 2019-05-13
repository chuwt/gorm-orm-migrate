package model

type TestApp struct {
    Name string `gorm:"type:varchar(32)"`
    BName string `gorm:"type:varchar(32)"`
}

func (*TestApp) TableName() string {
    return "test_app"
}

/* 用于多表
    {表名}_01 - {表名}_10
 */
func (*TestApp) MultiTable() int {
    return 10
}