package gogql_test

import (
	"reflect"
	"testing"

	"github.com/cipriantarta/gogql"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"
)

type VehicleAttributes struct {
	Color string
	Power int
}

type Vehicle struct {
	Make       string
	Model      string
	Attributes *VehicleAttributes
}

type User struct {
	ID       int    `graphql:"readonly"`
	Email    string `graphql:"required"`
	Password string `graphql:"inputonly,required"`
}

func (u *User) ResolveID(p graphql.ResolveParams) (interface{}, error) {
	return 1, nil
}

type Query struct {
	Hello string
	User  *User
}

func (q *Query) ResolveHello(p graphql.ResolveParams) (interface{}, error) {
	return "world", nil
}

func (q *Query) ResolveUser(p graphql.ResolveParams) (interface{}, error) {
	return &User{}, nil
}

type Mutation struct {
	CreateUser    *User
	CreateVehicle *Vehicle
}

func (m *Mutation) ResolveCreateUser(p graphql.ResolveParams, data *User) (*User, error) {
	// do something with the data maybe
	data.ID = 1
	return data, nil
}

func (m *Mutation) ResolveCreateVehicle(p graphql.ResolveParams, data *Vehicle) (*Vehicle, error) {
	return data, nil
}

func TestQuery(t *testing.T) {
	root := &Query{}
	s, err := gogql.New(root, nil, nil, nil, nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	q := `{
                hello
                user {
                        id
                }
        }`
	r := graphql.Do(graphql.Params{Schema: *s, RequestString: q})
	if len(r.Errors) > 0 {
		t.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}

	e := map[string]interface{}{
		"hello": "world",
		"user": map[string]interface{}{
			"id": 1,
		},
	}
	if !reflect.DeepEqual(r.Data, e) {
		t.Fatalf("Bad result, query: %v, result: %v", q, testutil.Diff(e, r.Data))
	}
}

func TestMutation(t *testing.T) {
	q := &Query{}
	m := &Mutation{}
	s, err := gogql.New(q, m, nil, nil, nil, 10)
	if err != nil {
		t.Fatal(err)
	}

	rs := `mutation {createUser(email: "test@email.com", password:"test"){id}}`
	r := graphql.Do(graphql.Params{Schema: *s, RequestString: rs})
	if len(r.Errors) > 0 {
		t.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	e := map[string]interface{}{
		"createUser": map[string]interface{}{
			"id": 1,
		},
	}
	if !reflect.DeepEqual(r.Data, e) {
		t.Fatalf("Bad result, mutation: %v, result: %v", rs, testutil.Diff(e, r.Data))
	}
}

func TestMutationWithNestedStruct(t *testing.T) {
	q := &Query{}
	m := &Mutation{}
	s, err := gogql.New(q, m, nil, nil, nil, 10)
	if err != nil {
		t.Fatal(err)
	}

	rs := `mutation {
            createVehicle(make: "VW", model: "Tiguan", attributes: {color: "red", power: 103}) {
                make
                model
                attributes {
                    color
                    power
                }
            }
        }`
	r := graphql.Do(graphql.Params{Schema: *s, RequestString: rs})
	if len(r.Errors) > 0 {
		t.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	e := map[string]interface{}{
		"createVehicle": map[string]interface{}{
			"make":  "VW",
			"model": "Tiguan",
			"attributes": map[string]interface{}{
				"color": "red",
				"power": 103,
			},
		},
	}
	if !reflect.DeepEqual(r.Data, e) {
		t.Fatalf("Bad result, mutation: %v, result: %v", rs, testutil.Diff(e, r.Data))
	}
}
