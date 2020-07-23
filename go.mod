module github.com/cipriantarta/gogql

go 1.14

require (
	github.com/graphql-go/graphql v0.7.9
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/mitchellh/mapstructure v1.3.3
	github.com/pkg/errors v0.9.1
	github.com/sanity-io/litter v1.2.0
)

replace github.com/cipriantarta/gogql/pkg/builder => ./pkg/builder
