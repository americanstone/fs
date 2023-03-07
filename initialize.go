package fs

import (
	"context"
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/fs/dateTime"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/fs/modules"
	"github.com/farseer-go/fs/net"
	"github.com/farseer-go/fs/parse"
	"github.com/farseer-go/fs/snowflake"
	"github.com/farseer-go/fs/stopwatch"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// StartupAt 应用启动时间
var StartupAt dateTime.DateTime

// AppName 应用名称
var AppName string

// HostName 主机名称
var HostName string

// AppId 应用ID
var AppId int64

// AppIp 应用IP
var AppIp string

// ProcessId 进程Id
var ProcessId int

// Context 最顶层的上下文
var Context context.Context

// 依赖的模块
var dependModules []modules.FarseerModule

var callbackFnList []callbackFn

type callbackFn struct {
	f    func()
	name string
}

// Initialize 初始化框架
func Initialize[TModule modules.FarseerModule](appName string) {
	sw := stopwatch.StartNew()
	Context = context.Background()
	AppName = appName
	ProcessId = os.Getppid()
	HostName, _ = os.Hostname()
	StartupAt = dateTime.Now()
	rand.Seed(time.Now().UnixNano())
	snowflake.Init(parse.HashCode64(HostName), rand.Int63n(32))
	AppId = snowflake.GenerateId()
	AppIp = net.GetIp()

	flog.Println("AppName： ", flog.Colors[2](AppName))
	flog.Println("AppID：   ", flog.Colors[2](AppId))
	flog.Println("AppIP：   ", flog.Colors[2](AppIp))
	flog.Println("HostName：", flog.Colors[2](HostName))
	flog.Println("HostTime：", flog.Colors[2](StartupAt.ToString("yyyy-MM-dd hh:mm:ss")))
	flog.Println("PID：     ", flog.Colors[2](ProcessId))
	showComponentLog()
	flog.Println("---------------------------------------")

	var startupModule TModule
	//flog.Println("Loading Module...")
	dependModules = modules.Distinct(modules.GetDependModule(startupModule))
	flog.Println("Loaded, " + flog.Red(len(dependModules)) + " modules in total")
	flog.Println("---------------------------------------")

	modules.StartModules(dependModules)
	flog.Println("---------------------------------------")
	flog.Println("Initialization completed, total time：" + sw.GetMillisecondsText())

	// 健康检查
	healthChecks := container.ResolveAll[core.IHealthCheck]()
	if len(healthChecks) > 0 {
		flog.Println("Health Check...")
		isSuccess := true
		for _, healthCheck := range healthChecks {
			item, err := healthCheck.Check()
			if err == nil {
				flog.Printf("%s%s\n", flog.Green("【✓】"), item)
			} else {
				flog.Errorf("%s%s：%s", flog.Red("【✕】"), item, flog.Red(err.Error()))
				isSuccess = false
			}
		}
		flog.Println("---------------------------------------")

		if !isSuccess {
			//os.Exit(-1)
			panic("健康检查失败")
		}
	}
	// 加载callbackFnList，启动后才执行的模块
	if len(callbackFnList) > 0 {
		for index, fn := range callbackFnList {
			sw.Restart()
			fn.f()
			flog.Println("Run " + strconv.Itoa(index+1) + "：" + fn.name + "，Use：" + sw.GetMillisecondsText())
		}
		flog.Println("---------------------------------------")
	}
}

// 组件日志
func showComponentLog() {
	err := configure.ReadInConfig()
	if err != nil { // 捕获读取中遇到的error
		_ = flog.Errorf("An error occurred while reading: %s \n", err)
	}

	logConfig := configure.GetSubNodes("Log.Component")
	var logSets []string
	for k, v := range logConfig {
		if v == true {
			logSets = append(logSets, k)
		}
	}
	if len(logSets) > 0 {
		flog.Println("Log Switch：", flog.Colors[2](strings.Join(logSets, " ")))
	}
}

// Exit 应用退出
func Exit(code int) {
	modules.ShutdownModules(dependModules)
	os.Exit(code)
}

// AddInitCallback 添加框架启动完后执行的函数
func AddInitCallback(name string, fn func()) {
	callbackFnList = append(callbackFnList, callbackFn{name: name, f: fn})
}
