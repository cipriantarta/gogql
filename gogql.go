package gogql

import (
	"reflect"

	"github.com/cipriantarta/gogql/pkg/builder"
	"github.com/graphql-go/graphql"
)

func New(query interface{},
	mutation interface{},
	scalars map[string]*graphql.Scalar,
	enums map[string]*graphql.Enum,
	objectTypes map[string]*graphql.Object) (*graphql.Schema, error) {

	b := builder.New()
	for k, v := range scalars {
		b.Scalar(k, v)
	}
	for k, v := range objectTypes {
		b.Object(k, v)
	}
	for k, v := range enums {
		b.Enum(k, v)
	}

	qf, err := b.QueryFields(reflect.ValueOf(query), reflect.Value{})
	if err != nil {
		return nil, err
	}

	var mutationObject *graphql.Object
	if mutation != nil {
		mf, err := b.QueryFields(reflect.ValueOf(mutation), reflect.Value{})
		if err != nil {
			return nil, err
		}
		mutationObject = graphql.NewObject(
			graphql.ObjectConfig{
				Name:   "Mutation",
				Fields: mf,
			})
	}

	s, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(
			graphql.ObjectConfig{
				Name:   "Query",
				Fields: qf,
			}),
		Mutation: mutationObject,
	})
	if err != nil {
		return nil, err
	}
	return &s, nil
}

type Connection struct {
	t reflect.Value
}

func NewConnection(i interface{}) *Connection {
	return &Connection{t: reflect.ValueOf(i)}
}

func (c *Connection) Type() reflect.Value {
	return c.t
}
