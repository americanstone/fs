package trace

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/asyncLocal"
	"github.com/farseer-go/fs/path"
	"github.com/farseer-go/linkTrace/eumCallType"
	"runtime"
	"strings"
	"time"
)

// ScopeLevel 层级列表
var ScopeLevel = asyncLocal.New[collections.List[BaseTraceDetail]]()

// BaseTraceDetail 埋点明细（基类）
type BaseTraceDetail struct {
	DetailId       int64            // 明细ID
	ParentDetailId int64            // 父级明细ID
	Level          int              // 当前层级（入口为0层）
	MethodName     string           // 调用方法
	CallType       eumCallType.Enum // 调用类型
	Timeline       time.Duration    // 从入口开始统计
	UnTraceTs      time.Duration    // 上一次结束到现在开始之间未Trace的时间
	StartTs        int64            // 调用开始时间戳
	EndTs          int64            // 调用停止时间戳
	UseTs          time.Duration    // 总共使用时间毫秒
	ignore         bool             // 忽略这次的链路追踪
	Exception      ExceptionStack   // 异常信息
}

type ExceptionStack struct {
	CallFile         string // 调用者文件路径
	CallLine         int    // 调用者行号
	CallFuncName     string // 调用者函数名称
	IsException      bool   // 是否执行异常
	ExceptionMessage string // 异常信息
}

func (receiver *BaseTraceDetail) SetSql(DbName string, tableName string, sql string) {}

// End 链路明细执行完后，统计用时
func (receiver *BaseTraceDetail) End(err error) {
	receiver.EndTs = time.Now().UnixMicro()
	receiver.UseTs = time.Duration(receiver.EndTs-receiver.StartTs) * time.Microsecond

	if err != nil {
		receiver.Exception.IsException = true
		receiver.Exception.ExceptionMessage = err.Error()
		// 调用者
		receiver.Exception.CallFile, receiver.Exception.CallFuncName, receiver.Exception.CallLine = GetCallerInfo()
	}

	// 移除层级
	lstScope := ScopeLevel.Get()
	if !lstScope.IsNil() {
		lstScope.RemoveAt(lstScope.Count() - 1)
		ScopeLevel.Set(lstScope)
	}
}

func (receiver *BaseTraceDetail) Ignore() {
	receiver.ignore = true
}
func (receiver *BaseTraceDetail) IsIgnore() bool {
	return receiver.ignore
}
func (receiver *BaseTraceDetail) GetLevel() int {
	return receiver.Level
}

var ComNames = []string{"/farseer-go/async/", "/farseer-go/cache/", "/farseer-go/cacheMemory/", "/farseer-go/collections/", "/farseer-go/data/", "/farseer-go/elasticSearch/", "/farseer-go/etcd/", "/farseer-go/eventBus/", "/farseer-go/fs/", "/farseer-go/linkTrace/", "/farseer-go/mapper/", "/farseer-go/queue/", "/farseer-go/rabbit/", "/farseer-go/redis/", "/farseer-go/redisStream/", "/farseer-go/tasks/", "/farseer-go/utils/", "/farseer-go/webapi/", "/src/reflect/", "/usr/local/go/src/", "gorm.io/"}

func IsSysCom(file string) bool {
	for _, comName := range ComNames {
		if strings.Contains(file, comName) {
			return true
		}
	}
	return false
}

func GetCallerInfo() (string, string, int) {
	// 获取调用栈信息
	pc := make([]uintptr, 15) // 假设最多获取 10 层调用栈
	n := runtime.Callers(0, pc)
	frames := runtime.CallersFrames(pc[:n])

	// 遍历调用栈帧
	for {
		frame, more := frames.Next()
		if !strings.HasSuffix(frame.File, "_test.go") && (!IsSysCom(frame.File) || strings.HasSuffix(frame.File, "healthCheck.go")) { // !strings.HasPrefix(file, gormSourceDir) ||
			// 移除绝对路径
			prefixFunc := frame.Function[0 : strings.Index(frame.Function, path.PathSymbol)+len(path.PathSymbol)]
			packageIndex := strings.Index(frame.File, prefixFunc)
			file := frame.File[packageIndex:]

			// 只要最后的方法名
			funcName := frame.Function[strings.LastIndex(frame.Function, path.PathSymbol)+len(path.PathSymbol):] + "()"
			return file, funcName, frame.Line
		}
		if !more {
			break
		}
	}
	return "", "", 0
}
