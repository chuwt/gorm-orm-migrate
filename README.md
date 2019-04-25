# Gorm-orm-migrate

### 这是啥
    用于记录数据库版本变更的工具
	
### 注意事项
    - 只支持gorm
    - 只支持index/unique index的注解
	
### 如何使用

##### 1. 创建manage.go
	import (
	    "gorm-orm-migrate/command"
	)
	
	func main() {
	    // add tables that you want to migrate
        command.AddTable(&model.Test{}, &model.Test2{})
        command.AddCommand(db.DB) // your gorm db (type *gorm.DB)
    }

##### 2. 命令行输入一下命令
    // 初始化migrations
    // 会在当前目录生成一个migrations文件
    // 会在数据库生成一个sql_version的表用于记录版本
    go run ./manage.go init
    
    // migrate
    // 会生成迁移文件, hash(uuid).go
    // 此时head变为最新，数据库中的sql_version则不变
    go run ./manage.go migrate
    
    // upgrade
    // 会按照文件中upgrade的进行更新数据库，并把sql_version中的版本变为当前版本
    go run ./manage.go upgrade
    
    // downgrade
    // 会按照当前版本的下一个版本进行回退
    go run ./manage.go downgrade
    
    // head
    // 获取当前head
    go run ./manage.go head
     
### 常见问题说明

    - 重复init
        不会造成影响，当前目录结构不会发生改变
    
    - 重复migrate
        不会造成影响，migrate的时候会检查当前版本和结构变更，分两种情况：
            1. upgrade之后重复migrate
                migrate时会检查当前结构是否变更，没有变更则不生成新版本文件
            2. upgrade之前
                migrate时会检查当前head和数据库head是否一致，如果不一致，则会提醒先upgrade
                
    - 重复upgrade
        不会造成影响，upgrade时会检查当前版本head和数据库版本sql_version,如果相同则不执行
        
    - 不init直接执行migrate或者upgrade
        不会造成影响，会先检查必要的条件是否满足