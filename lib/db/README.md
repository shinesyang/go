#### 自定义封装gorm.DB库,实现一些自己封装的方法,

##### 已经实现的方法有:
> 1. 实现数据库自动根据定义的模型进行数据库字段的增加与删除(自动增加原生的gorm.DB已经实现了)
> 2. 实现基于id的批量更新方法,(此方法是根据csdn上面的网友修改而来，感谢这位网友,地址: https://blog.csdn.net/m0_38101105/article/details/110007732)
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
DisplayName   string        `json:"display_name"`   
}
```

```go
// 连接时使用
db, err := gorm.Open(mysql.New(mysql.Config{
DSN:                       "xxxxxxxxxxxxx",
SkipInitializeWithVersion: false}), &gorm.Config{
NamingStrategy: schema.NamingStrategy{
SingularTable: true, // 禁止复数表
},
})

if err != nil {
    log.Fatal("mysql连接失败:%v", err)
}
mydb := lib.NewMyDb(db)


//自动更新删除表字段
err := mydb.MyAutoMigrate(&Game{},...)


// 根据id批量更新
game := []game{}
game1 := &game{"xx","yy"}
game2 := &game{"zz","ff"}
game = appned(game,game1,game2)
res,err := mydb.UpdatesById(game)
```


