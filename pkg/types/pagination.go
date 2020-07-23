package types

type PageArguments struct {
	First  *int
	Last   *int
	Before string
	After  string
	Limit  int
}

func (p *PageArguments) Slice() int {
	limit := p.Limit
	if p.First != nil && *p.First < limit {
		limit = *p.First
	}

	if p.Last != nil && *p.Last < limit {
		limit = *p.Last
	}

	p.Limit = limit
	return limit
}

func (p *PageArguments) PageInfo(total *int) (int, int) {
	skip := 0
	limit := p.Limit

	if p.Last != nil {
		if total != nil && *total > limit {
			skip = *total - limit
		}
	}
	return skip, limit
}
