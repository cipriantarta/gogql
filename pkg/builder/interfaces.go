package builder

import "github.com/graphql-go/graphql"

//IPageInfo pagination information interface
var IPageInfo = graphql.NewInterface(graphql.InterfaceConfig{
	Name: "IPageInfo",
	Fields: graphql.Fields{
		"startCursor": &graphql.Field{
			Type: graphql.String,
		},
		"endCursor": &graphql.Field{
			Type: graphql.String,
		},
		"hasMore": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Boolean),
		},
	},
	Description: "Relay pagination",
})

func (b *Builder) buildInterfaces() {
	if _, ok := b.interfaces["INode"]; !ok {
		b.interfaces["INode"] = graphql.NewInterface(graphql.InterfaceConfig{
			Name: "INode",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.NewNonNull(b.scalars["ID"]),
				},
			},
			Description: "Relay node interface",
		})
	}
	if _, ok := b.interfaces["IEdge"]; !ok {
		b.interfaces["IEdge"] = graphql.NewInterface(graphql.InterfaceConfig{
			Name: "IEdge",
			Fields: graphql.Fields{
				"node": &graphql.Field{
					Type: b.interfaces["INode"],
				},
				"cursor": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Description: "Relay edge interface",
		})
	}
	if _, ok := b.interfaces["IPageInfo"]; !ok {
		b.interfaces["IPageInfo"] = IPageInfo
	}
	if _, ok := b.interfaces["IConnection"]; !ok {
		b.interfaces["IConnection"] = graphql.NewInterface(graphql.InterfaceConfig{
			Name: "IConnection",
			Fields: graphql.Fields{
				"edges": &graphql.Field{
					Type: graphql.NewList(b.interfaces["IEdge"]),
				},
				"pageInfo": &graphql.Field{
					Type: graphql.NewNonNull(b.interfaces["IPageInfo"]),
				},
			},
			Description: "Relay connnection interface",
		})
	}
}
