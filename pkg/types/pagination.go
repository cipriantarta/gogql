package types

var PageLimit = 100

type PageArguments struct {
	First  *int
	Last   *int
	Before string
	After  string
}

func (p *PageArguments) Slice() int {
	limit := PageLimit
	if p.First != nil && *p.First < limit {
		limit = *p.First
	}

	if p.Last != nil && *p.Last < limit {
		limit = *p.Last
	}

	PageLimit = limit
	return limit
}

func (p *PageArguments) PageInfo(total *int) (int, int) {
	skip := 0
	limit := PageLimit

	if p.Last != nil {
		if total != nil && *total > limit {
			skip = *total - limit
		}
	}
	return skip, limit
}
