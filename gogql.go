package gogql

import (
	"reflect"

	"github.com/cipriantarta/gogql/pkg/builder"
	"github.com/graphql-go/graphql"
)

// Scalars - graphql map for scalar types
type Scalars map[string]*graphql.Scalar

// Enums - graphql map for enum types
type Enums map[string]*graphql.Enum

// Interfaces - graphql map for interfaces
type Interfaces map[string]*graphql.Interface

// InputObjects - graphql map for input types
type InputObjects map[string]graphql.Input

// OutputObjects - graphql map for output types
type OutputObjects map[string]graphql.Output

//New builds a new graphl Schema
func New(
	query interface{},
	mutation interface{},
	subscription interface{},
	scalars Scalars,
	enums Enums,
	interfaces Interfaces,
	queryTypes OutputObjects,
	mutationTypes InputObjects,
	paginationLimit int) (*graphql.Schema, error) {

	b := builder.New()
	b.PaginationLimit = paginationLimit
	for k, v := range scalars {
		b.Scalar(k, v)
	}
	for k, v := range queryTypes {
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
