package container

import (
	"fmt"
	"github.com/farseer-go/fs/container/eumLifecycle"
	"github.com/farseer-go/fs/flog"
	"os"
	"reflect"
)

// 容器
type container struct {
	name       string
	dependency map[reflect.Type][]componentModel // 依赖
	component  []componentModel                  // 实现类
}

// NewContainer 实例化一个默认容器
func NewContainer() *container {
	return &container{
		name:       "default",
		dependency: make(map[reflect.Type][]componentModel),
		component:  []componentModel{},
	}
}

// 注册实例，添加到依赖列表
func (r *container) addComponent(model componentModel) {
	componentModels, exists := r.dependency[model.interfaceType]
	if !exists {
		r.dependency[model.interfaceType] = []componentModel{model}
	} else {
		for index := 0; index < len(componentModels); index++ {
			if componentModels[index].name == model.name && componentModels[index].instanceType == model.instanceType {
				panic(fmt.Sprintf("container：已存在同样的注册对象,interfaceType=%s,name=%s,instanceType=%s", model.interfaceType.String(), model.name, reflect.TypeOf(model.instanceType).String()))
			}
		}
		r.dependency[model.interfaceType] = append(componentModels, model)
	}
	r.component = append(r.component, model)
}

// 注册构造函数
func (r *container) registerConstructor(constructor any, name string, lifecycle eumLifecycle.Enum) {
	constructorType := reflect.TypeOf(constructor)
	for inIndex := 0; inIndex < constructorType.NumIn(); inIndex++ {
		if name == "" && constructorType.In(inIndex).String() == constructorType.String() {
			panic("container：构造函数注册，当未设置别名时，入参的类型不能与返回的接口类型一样")
		}

		if constructorType.In(inIndex).Kind() != reflect.Interface {
			panic("container：构造函数注册，入参类型必须为interface")
		}
	}
	if constructorType.NumOut() != 1 {
		panic("container：构造函数注册，只能有1个出参")
	}
	interfaceType := constructorType.Out(0)
	if interfaceType.Kind() != reflect.Interface {
		panic("container：构造函数注册，出参类型只能为Interface")
	}
	model := NewComponentModel(name, lifecycle, interfaceType, constructor)
	r.addComponent(model)
}

// 注册实例
func (r *container) registerInstance(interfaceType any, ins any, name string, lifecycle eumLifecycle.Enum) {
	interfaceTypeOf := reflect.TypeOf(interfaceType)
	if interfaceTypeOf.Kind() == reflect.Pointer {
		interfaceTypeOf = interfaceTypeOf.Elem()
	}
	if interfaceTypeOf.Kind() != reflect.Interface {
		flog.Error("container：实例注册，interfaceType类型只能为Interface")
		os.Exit(-1)
	}
	model := NewComponentModelByInstance(name, lifecycle, interfaceTypeOf, ins)
	model.instance = ins
	r.addComponent(model)
}

// 获取对象
func (r *container) resolve(interfaceType reflect.Type, name string) any {
	if interfaceType.Kind() == reflect.Pointer {
		interfaceType = interfaceType.Elem()
	}

	// 通过Interface查找注册过的container
	if interfaceType.Kind() == reflect.Interface {
		componentModels, exists := r.dependency[interfaceType]
		if !exists {
			flog.Errorf("container：%s未注册", interfaceType.String())
			return nil
		}

		for i := 0; i < len(componentModels); i++ {
			// 找到了实现类
			if componentModels[i].name == name {
				return r.getOrCreateIns(interfaceType, i)
			}
		}
		flog.Errorf("container：%s未注册，name=%s", interfaceType.String(), name)

		// 结构对象，直接动态创建
	} else if interfaceType.Kind() == reflect.Struct {
		return r.createIns(componentModel{
			instanceType: interfaceType,
		})
	}
	return nil
}

// 根据lifecycle获取实例
func (r *container) getOrCreateIns(interfaceType reflect.Type, index int) any {
	// 单例
	if r.dependency[interfaceType][index].lifecycle == eumLifecycle.Single {
		if r.dependency[interfaceType][index].instance == nil {
			r.dependency[interfaceType][index].instance = r.createIns(r.dependency[interfaceType][index])
		}
		return r.dependency[interfaceType][index].instance
	} else {
		return r.createIns(r.dependency[interfaceType][index])
	}
}

// 根据类型，动态创建实例
func (r *container) createIns(model componentModel) any {
	if model.instanceType.Kind() == reflect.Func {
		var arr []reflect.Value
		// 构造函数，需要分别取出入参值
		for inIndex := 0; inIndex < model.instanceType.NumIn(); inIndex++ {
			val := reflect.ValueOf(r.resolveDefaultOrFirstComponent(model.instanceType.In(inIndex)))
			arr = append(arr, val)
		}
		if arr == nil {
			arr = []reflect.Value{}
		}
		return r.inject(model.instanceValue.Call(arr)[0].Interface())
	}

	if model.instanceType.Kind() == reflect.Struct || model.instanceType.Kind() == reflect.Pointer {
		if model.instance != nil {
			return r.inject(model.instance)
		} else {
			return r.injectByType(model.instanceType)
		}
	}
	return nil
}

// 获取对象，如果默认别名不存在，则使用第一个注册的实例
func (r *container) resolveDefaultOrFirstComponent(interfaceType reflect.Type) any {
	componentModels, exists := r.dependency[interfaceType]
	if !exists {
		flog.Errorf("container：%s未注册", interfaceType.String())
		return nil
	}

	findIndex := 0
	// 优先找默认实例
	for i := 0; i < len(componentModels); i++ {
		// 找到了实现类
		if componentModels[i].name == "" {
			findIndex = i
		}
	}
	return r.getOrCreateIns(interfaceType, findIndex)
}

// 解析注入
func (r *container) inject(ins any) any {
	if ins == nil {
		return ins
	}
	insVal := reflect.Indirect(reflect.ValueOf(ins))
	for i := 0; i < insVal.NumField(); i++ {
		field := insVal.Type().Field(i)
		if field.IsExported() && field.Type.Kind() == reflect.Interface {
			fieldIns := r.resolve(field.Type, field.Tag.Get("inject"))
			insVal.Field(i).Set(reflect.ValueOf(fieldIns))
		}
	}
	return ins
}

// 解析注入
func (r *container) injectByType(instanceType reflect.Type) any {
	instanceVal := reflect.New(instanceType).Elem()
	for i := 0; i < instanceVal.NumField(); i++ {
		field := instanceVal.Type().Field(i)
		if field.IsExported() && field.Type.Kind() == reflect.Interface {
			fieldIns := r.resolve(field.Type, field.Tag.Get("inject"))
			instanceVal.Field(i).Set(reflect.ValueOf(fieldIns))
		}
	}
	return instanceVal.Interface()
}
