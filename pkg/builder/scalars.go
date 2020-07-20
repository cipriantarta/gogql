package builder

import (
	"reflect"
	"time"

	"github.com/graphql-go/graphql"
)

var scalars = map[string]*graphql.Scalar{
	reflect.TypeOf(bool(false)).Name(): graphql.Boolean,
	reflect.TypeOf(int(0)).Name():      graphql.Int,
	reflect.TypeOf(int8(0)).Name():     graphql.Int,
	reflect.TypeOf(int16(0)).Name():    graphql.Int,
	reflect.TypeOf(int32(0)).Name():    graphql.Int,
	reflect.TypeOf(int64(0)).Name():    graphql.Int,
	reflect.TypeOf(uint(0)).Name():     graphql.Int,
	reflect.TypeOf(uint8(0)).Name():    graphql.Int,
	reflect.TypeOf(uint16(0)).Name():   graphql.Int,
	reflect.TypeOf(uint32(0)).Name():   graphql.Int,
	reflect.TypeOf(uint64(0)).Name():   graphql.Int,
	reflect.TypeOf(float32(0)).Name():  graphql.Float,
	reflect.TypeOf(float64(0)).Name():  graphql.Float,
	reflect.TypeOf(string("")).Name():  graphql.String,
	reflect.TypeOf(time.Time{}).Name(): graphql.DateTime,
	reflect.TypeOf(byte(0)).Name(): graphql.NewScalar(graphql.ScalarConfig{
		Name: "Byte",
		Serialize: func(value interface{}) interface{} {
			if v, ok := value.(byte); ok {
				return string(rune(v))
			}
			return nil
		},
	}),
}
