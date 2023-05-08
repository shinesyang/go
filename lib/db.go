package lib

import (
	"fmt"
	"github.com/asim/go-micro/v3/util/log"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

type MyDB struct {
	*gorm.DB
}

func NewMyDb(db *gorm.DB) *MyDB {
	return &MyDB{db}
}

func (d *MyDB) MyAutoMigrate(values ...interface{}) error {
	Migrator := d.Migrator()

	// 自封装删除字段
	if err := d.remove(Migrator, values...); err != nil {
		return err
	}

	// 调用gorm.DB内置AutoMigrate方法
	return Migrator.AutoMigrate(values...)
}

func (d *MyDB) GetName(t string) string {
	i := len(t) - 1
	sqBrackets := 0
	for i >= 0 && (t[i] != '.' || sqBrackets != 0) {
		switch t[i] {
		case ']':
			sqBrackets++
		case '[':
			sqBrackets--
		}
		i--
	}
	return t[i+1:]
}

// 自定义过滤.通过struct(gorm:"not_set;-")属性来删除不需要的表字段
func (d *MyDB) remove(migrator gorm.Migrator, values ...interface{}) error {
	/*
		自定义过滤.通过struct(gorm:"not_set;-")属性来删除不需要的表字段
		not_set;-必须同时存在,
		not_set在自定义方法中用于删除表中存在的字段.
		-在内置方法AutoMigrate中用于过滤不处理的字段,不然使用not_set删除的字段会重新创建
	*/
	for _, value := range values {
		reflectType := reflect.ValueOf(value).Type()
		tableNameUpper := d.GetName(reflectType.String())
		tableName := strings.ToLower(tableNameUpper)
		log.Infof("获取当前的table name: %v", tableName)
		// 判断传入的value是否是指针类的类型.是则不能直接取类型.需要使用reflectType.Elem()转换一下
		for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
			reflectType = reflectType.Elem()
		}

		/*
			反射遍历获取字段及标签用于判断标签是否存在not_set.存在则删除该字段
			这里通过json标签来获取filedName,所以在定义struct时,必须要定义json标签
		*/
		for i := 0; i < reflectType.NumField(); i++ {
			log.Info(fmt.Sprintf("当前标签属性: %v,当前字段名:%v", reflectType.Field(i).Tag, reflectType.Field(i).Name))
			filedName := reflectType.Field(i).Tag.Get("json")
			if migrator.HasTable(value) {
				if migrator.HasColumn(value, filedName) {
					tagName, ok := reflectType.Field(i).Tag.Lookup("gorm")
					if ok {
						if strings.Contains(tagName, "not_set") {
							log.Infof("删除字段: %v", filedName)
							return migrator.DropColumn(value, filedName)
						}
					}
				}

			}
		}
	}
	return nil
}
