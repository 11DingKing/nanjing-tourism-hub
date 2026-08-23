package incident

import "time"

type Summary struct {
	Total, Open, Resolved, HighSeverity int
	Oldest                              *time.Time
}

func (r *Registry) Summary(region string) Summary {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := Summary{}
	for _, i := range r.incidents {
		if region != "" && i.RegionID != region {
			continue
		}
		out.Total++
		if i.Status == Resolved {
			out.Resolved++
		} else {
			out.Open++
		}
		if i.Severity >= 4 {
			out.HighSeverity++
		}
		if out.Oldest == nil || i.CreatedAt.Before(*out.Oldest) {
			t := i.CreatedAt
			out.Oldest = &t
		}
	}
	return out
}
