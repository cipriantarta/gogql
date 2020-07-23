package builder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cipriantarta/gogql/pkg/types"
	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type nodeType struct {
	source       reflect.Value
	inputOnly    bool
	readOnly     bool
	noNull       bool
	skip         bool
	name         string
	alias        string
	description  string
	resolver     graphql.FieldResolveFn
	resolverArgs graphql.FieldConfigArgument
	isRelay      bool
	relay        *relayInfo
}

type Builder struct {
	scalars         map[string]*graphql.Scalar
	queryTypes      map[string]*graphql.Object
	mutationTypes   map[string]*graphql.InputObject
	interfaces      map[string]*graphql.Interface
	enums           map[string]*graphql.Enum
	PaginationLimit int
}

func New() *Builder {
	return &Builder{
		scalars:         scalars,
		interfaces:      make(map[string]*graphql.Interface),
		queryTypes:      make(map[string]*graphql.Object),
		mutationTypes:   make(map[string]*graphql.InputObject),
		enums:           make(map[string]*graphql.Enum),
		PaginationLimit: 100,
	}
}

// Scalar - Add/Replace an existing scalar type with a custom one
func (b *Builder) Scalar(name string, value *graphql.Scalar) {
	b.scalars[name] = value
}

// Object - Add/Replace an existing object type with a custom one
func (b *Builder) Object(name string, value *graphql.Object) {
	b.queryTypes[name] = value
}

// Enum - Add/Replace an existing enum type with a custom one
func (b *Builder) Enum(name string, value *graphql.Enum) {
	b.enums[name] = value
}

func (b *Builder) QueryFields(source reflect.Value, parent reflect.Value) (graphql.Fields, error) {
	result := make(graphql.Fields, 0)
	if source.IsValid() && source.IsZero() {
		source = reflect.New(source.Type())
	}
	nodes, err := b.buildObject(source, parent)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if node.skip {
			continue
		}
		if !node.source.CanSet() {
			continue
		}
		if node.inputOnly {
			continue
		}
		name := node.alias
		if name == "" {
			name = strcase.ToLowerCamel(node.name)
		}
		var gType graphql.Type
		if node.isRelay {
			gType, err = b.buildConnection(node.source, parent)
			if err != nil {
				return nil, err
			}
		} else {
			gType, err = b.mapOutput(node.source, parent)
			if err != nil {
				return nil, err
			}
		}
		if gType == nil {
			continue
		}
		if node.noNull {
			gType = graphql.NewNonNull(gType)
		}

		field := &graphql.Field{
			Name:        name,
			Type:        gType,
			Description: node.description,
			Resolve:     node.resolver,
			Args:        node.resolverArgs,
		}
		result[name] = field
	}
	return result, nil
}

func (b *Builder) InputFields(source reflect.Value, parent reflect.Value) (graphql.InputObjectConfigFieldMap, error) {
	result := make(graphql.InputObjectConfigFieldMap, 0)
	nodes, err := b.buildObject(source, parent)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if node.skip {
			continue
		}
		if !node.source.CanSet() {
			continue
		}
		if node.readOnly {
			continue
		}

		name := strcase.ToLowerCamel(node.name)
		gType, err := b.mapInput(node.source, parent)
		if err != nil {
			return nil, err
		}

		field := &graphql.InputObjectFieldConfig{
			Type: gType,
		}
		result[name] = field
	}
	return result, nil
}

func (b *Builder) buildObject(source reflect.Value, parent reflect.Value) ([]*nodeType, error) {
	nodes := make([]*nodeType, 0)
	if source.Kind() == reflect.Ptr {
		return b.buildObject(source.Elem(), source)
	}
	if source.Kind() != reflect.Struct {
		return nil, errors.Errorf("Expected a struct for %s", source)
	}

	for i := 0; i < source.NumField(); i++ {
		fv := source.Field(i)
		ft := source.Type().Field(i)
		node := &nodeType{
			source: fv,
			name:   ft.Name,
		}
		if tag, ok := ft.Tag.Lookup("graphql"); ok {
			for _, v := range strings.Split(tag, ",") {
				switch v {
				case "inputonly":
					node.inputOnly = true
				case "readonly":
					node.readOnly = true
				case "nonull":
					node.noNull = true
				case "-":
					node.skip = true
				}
				if strings.HasPrefix(v, "alias") {
					alias := strings.TrimPrefix(v, "alias=")
					if alias != v {
						node.alias = alias
					}
				}
				if strings.HasPrefix(v, "description") {
					d := strings.TrimPrefix(v, "description=")
					if d != v {
						node.description = strings.Trim(d, "\"")
					}
				}
			}
		}
		if tag, ok := ft.Tag.Lookup("relay"); ok {
			node.isRelay = true
			node.relay = &relayInfo{
				key:    "ID",
				method: "String",
			}
			for _, v := range strings.Split(tag, ",") {
				relay := strings.Split(v, "=")
				if len(relay) != 2 {
					continue
				}
				switch relay[0] {
				case "key":
					node.relay.key = relay[1]
				case "method":
					node.relay.method = relay[1]
				}
			}
		}
		if parent.IsValid() {
			node.resolver, node.resolverArgs = b.resolver(parent, ft.Name, node.isRelay, node.relay)
		} else {
			node.resolver, node.resolverArgs = b.resolver(source, ft.Name, node.isRelay, node.relay)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (b *Builder) resolver(source reflect.Value, fieldName string, isRelay bool, relay *relayInfo) (graphql.FieldResolveFn, graphql.FieldConfigArgument) {
	if !source.IsValid() {
		return nil, nil
	}

	name := "Resolve" + strings.Title(fieldName)
	method := source.MethodByName(name)
	if !method.IsValid() {
		return nil, nil
	}
	methodType := method.Type()
	nIn := methodType.NumIn()
	nOut := methodType.NumOut()
	if nOut != 2 {
		panic(fmt.Sprintf("%s expected two output params. Got %d", name, nOut))
	}
	if !methodType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		panic(fmt.Sprintf("%s second output parameter must be of type error", name))
	}

	in := make([]reflect.Value, nIn)
	args := make(graphql.FieldConfigArgument)
	var arg interface{}
	if nIn > 0 {
		p := methodType.In(0)
		if p.Kind() == reflect.Ptr {
			p = p.Elem()
		}
		if p != reflect.TypeOf(graphql.ResolveParams{}) {
			panic(fmt.Sprintf("First argument to %s must be `ResolveParams`", name))
		}
	}
	if nIn > 1 {
		p := methodType.In(1)
		if p.Kind() == reflect.Ptr {
			p = p.Elem()
		}
		if isRelay {
			relayArgs := reflect.TypeOf(types.PageArguments{})
			if p != relayArgs {
				panic(fmt.Sprintf("Second argument to %s must be `PageArguments`", name))
			}
			b.arguments(reflect.TypeOf(types.PageArguments{}), args, name)
		} else {
			if p.Kind() != reflect.Struct {
				panic(fmt.Sprintf("Second argument to %s must be a struct", name))
			}
			b.arguments(p, args, name)
			arg = reflect.Zero(methodType.In(1)).Interface()
		}
	}
	if nIn > 2 {
		if !isRelay {
			panic(fmt.Sprintf("%s must have maximum 2 arguments when not using relay", name))
		}
		if nIn > 3 {
			panic(fmt.Sprintf("%s must have maximum 3 arguments when using relay", name))
		}
		p := methodType.In(2)
		if p.Kind() == reflect.Ptr {
			p = p.Elem()
		}
		if p.Kind() != reflect.Struct {
			panic(fmt.Sprintf("Third argument to %s must be a struct", name))
		}
		b.arguments(methodType.In(2), args, name)
		arg = reflect.Zero(methodType.In(2)).Interface()
	}
	m := func(p graphql.ResolveParams) (interface{}, error) {
		var pageArgs *types.PageArguments
		v := reflect.ValueOf(p.Source)
		if v.IsValid() {
			m := v.MethodByName(name)
			if m.IsValid() {
				method = m
			}
		}
		if nIn > 0 {
			in[0] = reflect.ValueOf(p)
		}
		if nIn > 1 {
			if isRelay {
				pageArgs = &types.PageArguments{Limit: b.PaginationLimit}
				if err := mapstructure.Decode(p.Args, pageArgs); err != nil {
					panic(err)
				}
				in[1] = reflect.ValueOf(pageArgs)
			} else {
				if err := mapstructure.Decode(p.Args, &arg); err != nil {
					panic(err)
				}
				in[1] = reflect.ValueOf(arg)
			}

		}
		if nIn > 2 {
			if err := mapstructure.Decode(p.Args, &arg); err != nil {
				panic(err)
			}
			in[2] = reflect.ValueOf(arg)
		}
		r := method.Call(in)
		var err error = nil
		if e, ok := r[1].Interface().(error); ok {
			err = e
		}
		if isRelay {
			return connectionResolver(r[0].Interface(), err, relay, pageArgs)
		}
		return r[0].Interface(), err
	}
	return m, args
}

func (b *Builder) arguments(t reflect.Type, args graphql.FieldConfigArgument, resolverName string) {
	parent := reflect.Value{}
	source := reflect.New(t).Elem()
	valid := false
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ft := source.FieldByName(f.Name)
		if !ft.CanSet() {
			continue
		}

		description := ""
		name := strcase.ToLowerCamel(f.Name)
		el := f.Type
		if el.Kind() == reflect.Ptr {
			el = el.Elem()
		}

		argVal := reflect.New(el).Elem()
		v, err := b.mapInput(argVal, parent)
		if err != nil {
			panic(err)
		}

		if tag, ok := f.Tag.Lookup("graphql"); ok {
			skip := false
			for _, t := range strings.Split(tag, ",") {
				switch t {
				case "required":
					v = graphql.NewNonNull(v)
				case "readonly", "-":
					skip = true
				}
				if strings.HasPrefix(t, "alias") {
					alias := strings.TrimPrefix(t, "alias=")
					if alias != t {
						name = alias
					}
				}
				if strings.HasPrefix(t, "description") {
					d := strings.TrimPrefix(t, "description=")
					if d != t {
						description = strings.Trim(d, "\"")
					}
				}
			}
			if skip {
				continue
			}
		}

		arg := &graphql.ArgumentConfig{
			Type:        v,
			Description: description,
		}
		args[name] = arg
		valid = true
	}
	if !valid {
		panic(fmt.Sprintf("%s last argument has no exported fields", resolverName))
	}
}
