package modules

import (
	"github.com/farseer-go/fs/flog"
	"os"
	"reflect"
)

var moduleMap = make(map[string]int64)

// IsLoad 模块是否加载
func IsLoad(module FarseerModule) bool {
	moduleName := reflect.TypeOf(module).String()
	_, isExists := moduleMap[moduleName]
	return isExists
}

// ThrowIfNotLoad 如果没加载模块时，退出应用
func ThrowIfNotLoad(module FarseerModule) {
	load := IsLoad(module)
	if !load {
		moduleName := reflect.TypeOf(module).String()
		flog.Errorf("使用%s模块时，需要在启动模块中依赖%s模块，", flog.Colors[4](moduleName), flog.Colors[4](moduleName))
		os.Exit(1)
	}
}
