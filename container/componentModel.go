package container

import (
	"github.com/farseer-go/fs/container/eumLifecycle"
	"reflect"
)

// 实现类模型
type componentModel struct {
	name          string            // 别名
	lifecycle     eumLifecycle.Enum // 生命周期
	interfaceType reflect.Type      // 继承的接口
	instanceType  reflect.Type      // 函数类型
	instanceValue reflect.Value     // 函数值
	instance      any               // 实例
}

func NewComponentModel(name string, lifecycle eumLifecycle.Enum, interfaceType reflect.Type, funcIns any) componentModel {
	return componentModel{
		name:          name,
		lifecycle:     lifecycle,
		interfaceType: interfaceType,
		instanceType:  reflect.TypeOf(funcIns),
		instanceValue: reflect.ValueOf(funcIns),
	}
}

func NewComponentModelByInstance(name string, lifecycle eumLifecycle.Enum, interfaceType reflect.Type, instance any) componentModel {
	return componentModel{
		name:          name,
		lifecycle:     lifecycle,
		interfaceType: interfaceType,
		instanceType:  reflect.TypeOf(instance),
		instanceValue: reflect.ValueOf(instance),
		instance:      instance,
	}
}
