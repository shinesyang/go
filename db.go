package lib

import (
	"fmt"
	"github.com/asim/go-micro/v3/util/log"
	"github.com/jinzhu/gorm"
	"reflect"
	"strings"
)

type MyDB struct {
	*gorm.DB
}

func NewMyDb(db *gorm.DB) *MyDB {
	return &MyDB{db}
}

func (d *MyDB) MyAutoMigrate(values ...interface{}) *gorm.DB {
	db := d.Unscoped()
	d.remove(values...)
	// 调用gorm.DB内置AutoMigrate方法
	db = d.AutoMigrate(values...)
	return db
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
func (d *MyDB) remove(values ...interface{}) {
	/*
		自定义过滤.通过struct(gorm:"not_set;-")属性来删除不需要的表字段
		not_set;-必须同时存在,
		not_set在自定义方法中用于删除表中存在的字段.
		-在内置方法AutoMigrate中用于过滤不处理的字段,不然使用not_set删除的字段会重新创建
	*/
	for _, value := range values {
		scope := d.NewScope(value)
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
			if scope.Dialect().HasTable(tableName) { // 判断表存在
				if scope.Dialect().HasColumn(tableName, filedName) { // 判断字段存在
					tagName, ok := reflectType.Field(i).Tag.Lookup("gorm")
					if ok {
						if strings.Contains(tagName, "not_set") {
							scope.Raw(fmt.Sprintf("ALTER TABLE %v DROP %v;", tableName, filedName)).Exec()
						}
					}
				}
			}
		}

		// 注释老方法

		//scope := db.NewScope(value)                // 实例Scope加载value对应的表
		//tableName := scope.TableName()             // 获取表名
		//quotedTableName := scope.QuotedTableName() // 引用到的表名
		//// 当表存在时,判断字段是否存在not_set tag(标签),存在则删除该字段
		//if scope.Dialect().HasTable(tableName) {
		//	for _, field := range scope.GetModelStruct().StructFields {
		//
		//		if scope.Dialect().HasColumn(tableName, field.DBName) {
		//			if field.IsNormal {
		//				log.Infof("获取到字段的标签:%v", field.Tag.Get("gorm"), field.DBName)
		//				tagName, ok := field.Tag.Lookup("gorm")
		//				if ok {
		//					if strings.Contains(tagName, "not_set") {
		//						scope.Raw(fmt.Sprintf("ALTER TABLE %v DROP %v;", quotedTableName, field.DBName)).Exec()
		//						//tagName = "-"
		//						//d.SetTagToStruct(value, field.Name, tagName)
		//					}
		//				}
		//			}
		//		}
		//	}
		//}
	}
}
