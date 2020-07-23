package builder

import (
	"encoding/base64"
	"reflect"

	"github.com/cipriantarta/gogql/pkg/types"
	"github.com/graphql-go/graphql"
)

type relayInfo struct {
	key    string
	method string
}
type PageInfo struct {
	StartCursor string `graphql:"nonull"`
	EndCursor   string `graphql:"nonull"`
	HasMore     bool   `graphql:"nonull"`
}

type Edge struct {
	Node   interface{}
	Cursor string `graphql:"nonull"`
}

type Connection struct {
	PageInfo *PageInfo
	Edges    interface{}
}

func (b *Builder) buildConnection(source reflect.Value, parent reflect.Value) (graphql.Output, error) {
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
	node, err := b.mapObject(el, parent, []*graphql.Interface{b.interfaces["INode"]}, name+"Node")
	if err != nil {
		return nil, err
	}

	if _, err := b.mapObject(reflect.ValueOf(&PageInfo{}), reflect.Value{}, []*graphql.Interface{b.interfaces["IPageInfo"]}, ""); err != nil {
		return nil, err
	}

	edge, err := b.mapObject(reflect.ValueOf(&Edge{}), reflect.Value{}, []*graphql.Interface{b.interfaces["IEdge"]}, name+"Edge")
	if err != nil {
		return nil, err
	}
	edge.AddFieldConfig("node", &graphql.Field{Type: graphql.NewNonNull(node)})
	edges := graphql.NewList(edge)

	connection, err := b.mapObject(reflect.ValueOf(&Connection{}), reflect.Value{}, []*graphql.Interface{b.interfaces["IConnection"]}, name+"Connection")
	if err != nil {
		return nil, err
	}
	connection.AddFieldConfig("edges", &graphql.Field{Type: graphql.NewNonNull(edges)})
	return graphql.NewNonNull(connection), nil
}

func connectionResolver(nodes interface{}, err error, relayInfo *relayInfo) (interface{}, error) {
	n := reflect.ValueOf(nodes)
	if n.Kind() != reflect.Slice {
		panic("Connection result expects a slice")
	}
	edges := make([]interface{}, 0)
	pageInfo := &PageInfo{HasMore: n.Len() > types.PageLimit}
	for i := 0; i < n.Len(); i++ {
		if i >= types.PageLimit {
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
		c := toGlobalID(cursor.Interface().(string))
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
