package builder

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

//Intersection - GraphQL types used either as Output or Input
type Intersection interface {
	Name() string
	Description() string
	String() string
	Error() error
}

var _ Intersection = (*graphql.Scalar)(nil)
var _ Intersection = (*graphql.Enum)(nil)
var _ Intersection = (*graphql.List)(nil)
var _ Intersection = (*graphql.NonNull)(nil)

func (b *Builder) mapOutput(source reflect.Value, parent reflect.Value) graphql.Output {
	if intersection := b.mapIntersection(source, false); intersection != nil {
		return intersection
	}
	if ptr := b.mapPointer(source, false); ptr != nil {
		return ptr
	}
	if source.Kind() == reflect.Struct {
		return b.mapObject(source, parent, nil, "")
	}
	return nil
}

func (b *Builder) mapInput(source reflect.Value, parent reflect.Value) graphql.Input {
	if intersection := b.mapIntersection(source, true); intersection != nil {
		return intersection
	}
	if ptr := b.mapPointer(source, true); ptr != nil {
		return ptr
	}
	name := fmt.Sprintf("%sInput", typeName(source.Type()))
	in, ok := b.mutationTypes[name]
	if ok {
		return in
	}

	fields := b.InputFields(source, parent)
	o := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   name,
		Fields: fields,
	})
	b.mutationTypes[name] = o
	return o
}

func (b *Builder) mapObject(source reflect.Value, parent reflect.Value, interfaces []*graphql.Interface, alias string) graphql.Output {
	name := alias
	if name == "" {
		name = typeName(source.Type())
	}
	obj, ok := b.queryTypes[name]
	if ok {
		return obj
	}
	fields, err := b.QueryFields(source, parent)
	if err != nil {
		panic(err)
	}
	obj = graphql.NewObject(graphql.ObjectConfig{
		Name:       name,
		Fields:     fields,
		Interfaces: interfaces,
	})
	b.queryTypes[name] = obj
	return obj
}

func (b *Builder) mapIntersection(source reflect.Value, isInput bool) Intersection {
	if scalar := b.mapScalar(source); scalar != nil {
		return scalar
	}
	if enum := b.mapEnum(source); enum != nil {
		return enum
	}
	if sequence := b.mapSequence(source, isInput); sequence != nil {
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

func (b *Builder) mapPointer(source reflect.Value, isInput bool) graphql.Type {
	if !isPtr(source) {
		return nil
	}
	var el = source.Elem()
	if !el.IsValid() {
		el = reflect.New(source.Type().Elem()).Elem()
	}
	if isInput {
		return b.mapInput(el, source)
	}
	return b.mapOutput(el, source)
}

func (b *Builder) mapSequence(source reflect.Value, isInput bool) *graphql.List {
	if !isSequence(source) {
		return nil
	}
	el := reflect.New(source.Type().Elem())
	var inner graphql.Type
	if isInput {
		inner = b.mapInput(el.Elem(), el)
	} else {
		inner = b.mapOutput(el.Elem(), el)
	}
	return graphql.NewList(inner)
}
