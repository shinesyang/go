package db

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

/*
	创建批量更新原生sql切片
	这里所以的反射都是基于json来转换的,所以必须要设置json的标签
*/
func (d *MyDB) buildBatchUpdateSQLArray(dataList interface{}, size int) ([]string, error) {

	fieldValue := reflect.ValueOf(dataList)
	fieldType := reflect.ValueOf(dataList).Type()
	for fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	// 获取表名(结构体名称作为表名)
	tableNameUpper := d.GetName(fieldType.String())
	tableName := strings.ToLower(tableNameUpper)

	// dataList长度sliceLength,字段长度fieldNum
	sliceLength := fieldValue.Len()
	fieldNum := fieldType.NumField()

	/*
		检验结构体标签是否为空和重复
	*/
	verifyTagDuplicate := make(map[string]string)
	idCheck := 0
	for i := 0; i < fieldNum; i++ {
		fieldTag := fieldType.Field(i).Tag.Get("json")

		fieldName := d.GetFieldName(fieldTag)
		if len(strings.TrimSpace(fieldName)) == 0 {
			return nil, errors.New("the structure attribute should have tag in json")
		}

		if strings.HasPrefix(fieldName, "id") {
			idCheck += 1
		}

		_, ok := verifyTagDuplicate[fieldName]
		if !ok {
			verifyTagDuplicate[fieldName] = fieldName
		} else {
			msg := fmt.Sprintf("the structure attribute %v tag is not allow duplication", fieldName)
			return nil, errors.New(msg)
		}
	}

	// 判断id是否唯一
	if idCheck <= 0 {
		return nil, errors.New("the structure attribute should have primary_key")
	} else if idCheck > 1 {
		return nil, errors.New("the structure attribute have exist more primary_key")
	}

	// 获取id/字段类型
	var IDList []string
	updateMap := make(map[string][]string)
	for i := 0; i < sliceLength; i++ {
		structValue := fieldValue.Index(i).Elem()
		for j := 0; j < fieldNum; j++ {
			elem := structValue.Field(j)

			var temp string
			switch elem.Kind() {
			case reflect.Int64:
				temp = strconv.FormatInt(elem.Int(), 10)
			case reflect.String:
				if strings.Contains(elem.String(), "'") {
					temp = fmt.Sprintf("'%v'", strings.ReplaceAll(elem.String(), "'", "\\'"))
				} else {
					temp = fmt.Sprintf("'%v'", elem.String())
				}
			case reflect.Float64:
				temp = strconv.FormatFloat(elem.Float(), 'f', -1, 64)
			case reflect.Bool:
				temp = strconv.FormatBool(elem.Bool())
			default:
				msg := fmt.Sprintf("type conversion error, param is %v", fieldType.Field(j).Tag.Get("json"))
				return nil, errors.New(msg)
			}

			gormTag := fieldType.Field(j).Tag.Get("json")

			fieldTag := d.GetFieldName(gormTag)

			if strings.HasPrefix(fieldTag, "id") {
				id, err := strconv.ParseInt(temp, 10, 64)
				if err != nil {
					return nil, err
				}
				// 判断是否传入了id
				if id < 1 {
					return nil, errors.New("this structure should have a primary key and gt 0")
				}
				IDList = append(IDList, temp)
				continue
			}

			valueList := append(updateMap[fieldTag], temp)
			updateMap[fieldTag] = valueList
		}
	}

	// 生成sql,size切割sql长度
	length := len(IDList)
	SQLQuantity := d.getSQLQuantity(length, size)
	var SQLArray []string
	k := 0

	for i := 0; i < SQLQuantity; i++ {
		count := 0

		var record bytes.Buffer
		record.WriteString("UPDATE " + tableName + " SET ")

		for fieldName, fieldValueList := range updateMap {
			record.WriteString(fieldName)
			record.WriteString(" = CASE " + "id")

			for j := k; j < len(IDList) && j < len(fieldValueList) && j < size+k; j++ {
				record.WriteString(" WHEN " + IDList[j] + " THEN " + fieldValueList[j])
			}
			count++
			if count != fieldNum-1 {
				record.WriteString(" END, ")
			}
		}

		record.WriteString(" END WHERE ")
		record.WriteString("id" + " IN (")
		min := size + k
		if len(IDList) < min {
			min = len(IDList)
		}
		record.WriteString(strings.Join(IDList[k:min], ","))
		record.WriteString(");")

		k += size
		SQLArray = append(SQLArray, record.String())
	}

	return SQLArray, nil
}

func (d *MyDB) getSQLQuantity(length, size int) int {
	SQLQuantity := int(math.Ceil(float64(length) / float64(size)))
	return SQLQuantity
}

func (d *MyDB) GetFieldName(fieldTag string) string {
	fieldTagArr := strings.Split(fieldTag, ":")
	if len(fieldTagArr) == 0 {
		return ""
	}

	fieldName := fieldTagArr[len(fieldTagArr)-1]

	return fieldName
}
