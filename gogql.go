package gogql

import (
	"reflect"

	"github.com/cipriantarta/gogql/pkg/builder"
	"github.com/graphql-go/graphql"
)

//New builds a new graphl Schema
func New(
	query interface{},
	mutation interface{},
	subscription interface{},
	scalars map[string]*graphql.Scalar,
	enums map[string]*graphql.Enum,
	objectTypes map[string]*graphql.Object, paginationLimit int) (*graphql.Schema, error) {

	b := builder.New()
	b.PaginationLimit = paginationLimit
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

	var mutationObj *graphql.Object
	if mutation != nil {
		mf, err := b.QueryFields(reflect.ValueOf(mutation), reflect.Value{})
		if err != nil {
			return nil, err
		}
		mutationObj = graphql.NewObject(
			graphql.ObjectConfig{
				Name:   "Mutation",
				Fields: mf,
			})
	}

	var subscriptionObj *graphql.Object
	if subscription != nil {
		sf, err := b.QueryFields(reflect.ValueOf(subscription), reflect.Value{})
		if err != nil {
			return nil, err
		}
		subscriptionObj = graphql.NewObject(
			graphql.ObjectConfig{
				Name:   "Subscription",
				Fields: sf,
			})
	}

	s, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(
			graphql.ObjectConfig{
				Name:   "Query",
				Fields: qf,
			}),
		Mutation:     mutationObj,
		Subscription: subscriptionObj,
	})
	if err != nil {
		return nil, err
	}
	return &s, nil
}
