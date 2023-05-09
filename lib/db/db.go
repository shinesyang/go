package db

import (
	"fmt"
	"github.com/asim/go-micro/v3/util/log"
	"gorm.io/gorm"
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

// 通过id主键实现批量更新
func (d *MyDB) UpdatesById(values interface{}) (interface{}, error) {
	sqlArray, err := d.buildBatchUpdateSQLArray(values, DefaultSize)
	if err != nil {
		log.Infof("创建sql切片失败: %v", err)
		return nil, err
	}

	// 启动事务
	tx := d.Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.Callback()
		}
	}()

	if err := tx.Error; err != nil {
		msg := fmt.Sprintf("批量更新启动事务时失败: %v", err)
		log.Info(msg)
		return nil, err
	}

	// 遍历sqlArray添加
	for _, value := range sqlArray {
		if err := d.Exec(value).Error; err != nil {
			tx.Callback()
			msg := fmt.Sprintf("批量更新失败: %v", err)
			log.Info(msg)
			return nil, err
		}
	}

	return values, tx.Commit().Error

}
