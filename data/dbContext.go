package data

import (
	"github.com/farseernet/farseer.go/configure"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

// DbContext 数据库上下文
type DbContext struct {
	// 数据库配置
	dbConfig *dbConfig
}

// NewDbContext 初始化上下文
func initConfig(dbName string) *DbContext {
	configString := configure.GetString("Database." + dbName)
	if configString == "" {
		panic("[farseer.yaml]找不到相应的配置：Database.\" + dbName")
	}
	dbConfig := configure.ParseConfig[dbConfig](configString)
	dbContext := &DbContext{
		dbConfig: &dbConfig,
	}
	dbContext.dbConfig.dbName = dbName
	return dbContext
}

// NewContext 数据库上下文初始化
// dbName：数据库配置名称
func NewContext[TDbContext any](dbName string) *TDbContext {
	if dbName == "" {
		panic("dbName入参必须设置有效的值")
	}
	dbConfig := initConfig(dbName) // 嵌入类型
	//var dbName string       // 数据库配置名称
	customContext := new(TDbContext)
	contextValueOf := reflect.ValueOf(customContext).Elem()

	for i := 0; i < contextValueOf.NumField(); i++ {
		field := contextValueOf.Field(i)
		fieldType := field.Type().String()
		if !field.CanSet() || !strings.HasPrefix(fieldType, "data.TableSet[") {
			continue
		}
		data := contextValueOf.Type().Field(i).Tag.Get("data")
		var tableName string
		if strings.HasPrefix(data, "name=") {
			tableName = data[len("name="):]
		}
		if tableName == "" {
			continue
		}
		// 再取tableSet的子属性，并设置值
		field.Addr().MethodByName("Init").Call([]reflect.Value{reflect.ValueOf(dbConfig), reflect.ValueOf(tableName)})
	}
	return customContext
}

// 获取对应驱动
func (dbContext *DbContext) getDriver() gorm.Dialector {
	// 参考：https://gorm.cn/zh_CN/docs/connecting_to_the_database.html
	switch strings.ToLower(dbContext.dbConfig.DataType) {
	case "mysql":
		return mysql.Open(dbContext.dbConfig.ConnectionString)
	case "postgresql":
		return postgres.Open(dbContext.dbConfig.ConnectionString)
	case "sqlite":
		return sqlite.Open(dbContext.dbConfig.ConnectionString)
	case "sqlserver":
		return sqlserver.Open(dbContext.dbConfig.ConnectionString)
	}
	panic("无法识别数据库类型：" + dbContext.dbConfig.DataType)
}
