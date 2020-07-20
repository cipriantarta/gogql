package builder

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type Intersaction interface {
	Name() string
	Description() string
	String() string
	Error() error
}

var _ Intersaction = (*graphql.Scalar)(nil)
var _ Intersaction = (*graphql.Enum)(nil)
var _ Intersaction = (*graphql.List)(nil)
var _ Intersaction = (*graphql.NonNull)(nil)

func (b *Builder) mapOutput(source reflect.Value, parent reflect.Value) (graphql.Output, error) {
	if intersection := b.mapIntersection(source, false); intersection != nil {
		return intersection, nil
	}
	if ptr := b.mapPointer(source); ptr != nil {
		return ptr, nil
	}
	if source.Kind() == reflect.Struct {
		return b.mapObject(source, parent, nil, "")
	}
	return nil, nil
}

func (b *Builder) mapInput(source reflect.Value, parent reflect.Value) (graphql.Input, error) {
	if intersection := b.mapIntersection(source, false); intersection != nil {
		return intersection, nil
	}
	if ptr := b.mapPointer(source); ptr != nil {
		return ptr, nil
	}
	name := fmt.Sprintf("%sInput", typeName(source.Type()))
	in, ok := b.mutationTypes[name]
	if ok {
		return in, nil
	}

	fields, err := b.InputFields(source, parent)
	if err != nil {
		return nil, err
	}
	o := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   name,
		Fields: fields,
	})
	b.mutationTypes[name] = o
	return o, nil
}

func (b *Builder) mapObject(source reflect.Value, parent reflect.Value, interfaces []*graphql.Interface, alias string) (*graphql.Object, error) {
	name := alias
	if name == "" {
		name = typeName(source.Type())
	}
	obj, ok := b.queryTypes[name]
	if ok {
		return obj, nil
	}
	fields, err := b.QueryFields(source, parent)
	if err != nil {
		return nil, err
	}
	obj = graphql.NewObject(graphql.ObjectConfig{
		Name:       name,
		Fields:     fields,
		Interfaces: interfaces,
	})
	b.queryTypes[name] = obj
	return obj, nil
}

func (b *Builder) mapIntersection(source reflect.Value, isInput bool) Intersaction {
	if scalar := b.mapScalar(source); scalar != nil {
		return scalar
	}
	if enum := b.mapEnum(source); enum != nil {
		return enum
	}
	if sequence := b.mapSequence(source); sequence != nil {
		return sequence
	}
	return nil
}

func (b *Builder) mapScalar(source reflect.Value) *graphql.Scalar {
	name := typeName(source.Type())
	if s, ok := b.scalars[name]; ok {
		return s
	}
	return nil
}

func (b *Builder) mapEnum(source reflect.Value) *graphql.Enum {
	name := typeName(source.Type())
	if e, ok := b.enums[name]; ok {
		return e
	}
	return nil
}

func (b *Builder) mapPointer(source reflect.Value) graphql.Output {
	if !isPtr(source) {
		return nil
	}
	var el = source.Elem()
	if !el.IsValid() {
		el = reflect.New(source.Type().Elem()).Elem()
	}
	r, _ := b.mapOutput(el, source)
	return r
}

func (b *Builder) mapSequence(source reflect.Value) *graphql.List {
	if !isSequence(source) {
		return nil
	}
	el := reflect.New(source.Type().Elem()).Elem()

	inner, err := b.mapOutput(el, reflect.Value{})
	if err != nil {
		panic(err)
	}
	return graphql.NewList(inner)
}
