package main

import "gorm-orm-migrate/model"

func main() {
    //command.AddTable(&model.TestApp{})
    //command.AddCommand(model.DB)
    a := model.TestApp{BName:"123", TableIndex:9}
    model.DB.Table(a.TableName()).Create(a)
}
