#### 自定义封装gorm.DB库

##### 实现数据库自动根据定义的模型进行数据库字段的增加与删除(自动增加原生的gorm.DB已经实现了)

##### 使用方法:
```go
// 定义 struct
/*
    使用方法，定义字段:
    PlatformId    string       `json:"platform_id" gorm:"not_null;not_set;-"`
    定义json和gorm字段, json定义被引用为表字段名,gorm not_set;-定义字段存在表中时删除，
 */

type Game struct {
Platform      string       `json:"platform" gorm:"not_null;not_set;-"`
}
```

```go
// 连接时使用
db, err := gorm.Open("mysql", "xxxxxxxx")
if err != nil {
    log.Fatal("mysql连接失败:%v", err)
}
mydb := lib.NewMyDb(db)


```


