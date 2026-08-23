package domain

func ValidCampaignTransition(from, to string) bool {
	return map[string]map[string]bool{"draft": {"active": true}, "active": {"paused": true, "closed": true}, "paused": {"active": true, "closed": true}}[from][to]
}
func ValidTaskTransition(from, to string) bool {
	return map[string]map[string]bool{"queued": {"claimed": true}, "claimed": {"in_progress": true, "queued": true}, "in_progress": {"completed": true, "failed": true, "queued": true}, "failed": {"queued": true}}[from][to]
}
func ValidReportTransition(from, to string) bool {
	return map[string]map[string]bool{"reported": {"triaged": true}, "triaged": {"assigned": true, "rejected": true}, "assigned": {"mitigating": true}, "mitigating": {"resolved": true, "escalated": true}, "escalated": {"mitigating": true}}[from][to]
}
func ValidReservationTransition(from, to string) bool {
	return map[string]map[string]bool{"held": {"confirmed": true, "cancelled": true}, "confirmed": {"fulfilled": true, "cancelled": true}}[from][to]
}
func ValidAidTransition(from, to string) bool {
	return map[string]map[string]bool{"draft": {"review": true}, "review": {"approved": true, "rejected": true}, "approved": {"paid": true}}[from][to]
}
