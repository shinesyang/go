package db

// 默认参数
var (
	DefaultSize = 100
)

// 获取表名
func (d *MyDB) GetName(tableName string) string {
	i := len(tableName) - 1
	sqBrackets := 0
	for i >= 0 && (tableName[i] != '.' || sqBrackets != 0) {
		switch tableName[i] {
		case ']':
			sqBrackets++
		case '[':
			sqBrackets--
		}
		i--
	}
	return tableName[i+1:]
}
