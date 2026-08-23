package tourism

import (
	"fmt"
	"time"
)

func transition(current, target string, allowed map[string][]string) error {
	if current == target {
		return nil
	}
	for _, next := range allowed[current] {
		if next == target {
			return nil
		}
	}
	return fmt.Errorf("%w: %s -> %s", ErrInvalidState, current, target)
}
func activateAgency(a Agency) (Agency, error) {
	if a.Name == "" || a.Country == "" {
		return a, ErrInvalidState
	}
	a.Active = true
	a.Version++
	return a, nil
}
func publishCampaign(c Campaign) (Campaign, error) {
	if e := transition(c.Status, "published", map[string][]string{"draft": {"published", "canceled"}}); e != nil {
		return c, e
	}
	c.Status = "published"
	c.Version++
	return c, nil
}
func openPackage(p Package) (Package, error) {
	if e := transition(p.Status, "draft", map[string][]string{"": []string{"draft"}}); e != nil {
		return p, e
	}
	if p.Quota <= 0 {
		return p, ErrInvalidState
	}
	p.Status = "draft"
	p.Version++
	return p, nil
}
func signAgreement(a Agreement, at time.Time) (Agreement, error) {
	if e := transition(a.Status, "proposed", map[string][]string{"": []string{"proposed"}}); e != nil {
		return a, e
	}
	a.Status = "proposed"
	a.SignedAt = &at
	a.Version++
	return a, nil
}
