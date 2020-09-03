package builder

import (
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/cipriantarta/gogql/pkg/types"
	"github.com/graphql-go/graphql"
)

type relayInfo struct {
	key    string
	method string
}

// PageInfo relay pagination info
type PageInfo struct {
	StartCursor string `graphql:"required"`
	EndCursor   string `graphql:"required"`
	HasMore     bool   `graphql:"required"`
}

// Edge relay edge
type Edge struct {
	Node   interface{}
	Cursor string `graphql:"required"`
}

// Connection relay connection
type Connection struct {
	PageInfo *PageInfo `graphql:"required"`
	Edges    interface{}
}

func (b *Builder) buildConnection(source reflect.Value, parent reflect.Value) graphql.Output {
	b.buildInterfaces()

	if isSequence(source) {
		source = reflect.New(source.Type().Elem()).Elem()
	}

	var el reflect.Value
	if isPtr(source) {
		el = source.Elem()
	}
	if !el.IsValid() {
		el = reflect.New(source.Type().Elem()).Elem()
	}
	name := typeName(el.Type())
	node := b.mapObject(el, parent, []*graphql.Interface{b.interfaces["INode"]}, name+"Node")

	_ = b.mapObject(reflect.ValueOf(&PageInfo{}), reflect.Value{}, []*graphql.Interface{b.interfaces["IPageInfo"]}, "")

	edge := b.mapObject(reflect.ValueOf(&Edge{}), reflect.Value{}, []*graphql.Interface{b.interfaces["IEdge"]}, name+"Edge")
	edge.AddFieldConfig("node", &graphql.Field{Type: graphql.NewNonNull(node)})
	edges := graphql.NewList(edge)

	connection := b.mapObject(reflect.ValueOf(&Connection{}), reflect.Value{}, []*graphql.Interface{b.interfaces["IConnection"]}, name+"Connection")
	connection.AddFieldConfig("edges", &graphql.Field{Type: graphql.NewNonNull(edges)})
	return graphql.NewNonNull(connection)
}

func connectionResolver(nodes interface{}, err error, relayInfo *relayInfo, pageArgs *types.PageArguments) (interface{}, error) {
	n := reflect.ValueOf(nodes)
	if n.Kind() != reflect.Slice {
		panic("Connection result expects a slice")
	}
	edges := make([]interface{}, 0)
	pageInfo := &PageInfo{HasMore: n.Len() > pageArgs.Limit}
	for i := 0; i < n.Len(); i++ {
		if i >= pageArgs.Limit {
			pageInfo.EndCursor = edges[i-1].(*Edge).Cursor
			break
		}
		node := n.Index(i)
		cursor := node.Elem().FieldByName(relayInfo.key)
		if reflect.Ptr == cursor.Kind() {
			cursor = cursor.Elem()
		}
		if !cursor.IsValid() {
			continue
		}
		m := cursor.MethodByName(relayInfo.method)
		if m.IsValid() {
			r := m.Call(nil)
			cursor = r[0]
		}
		c := toGlobalID(fmt.Sprintf("%v", cursor.Interface()))
		edges = append(edges, &Edge{
			Cursor: c,
			Node:   node.Interface(),
		})
		if i == 0 {
			pageInfo.StartCursor = c
		}
		if i == n.Len()-1 {
			pageInfo.EndCursor = c
		}
	}
	c := &Connection{
		PageInfo: pageInfo,
		Edges:    edges,
	}
	return c, err
}

func toGlobalID(v string) string {
	return base64.StdEncoding.EncodeToString([]byte(v))
}
