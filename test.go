package main

import (
    "gorm-orm-migrate/command"
    "gorm-orm-migrate/model"
)

func main() {
    command.AddTable(&model.TestApp{})
    command.AddCommand(model.DB)
}
