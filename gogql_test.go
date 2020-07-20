package gogql_test

import (
	"reflect"
	"testing"

	"github.com/cipriantarta/gogql"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"
)

type User struct {
	ID int
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

func TestSchema(t *testing.T) {
	root := &Query{}
	s, err := gogql.New(root, nil, nil, nil, nil)
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
