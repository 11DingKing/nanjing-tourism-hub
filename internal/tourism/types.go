package tourism

import "time"

type Agency struct {
	ID, Name, Country, City string
	Active                  bool
	Version                 int64
	CreatedAt               time.Time
}
type Campaign struct {
	ID, Name, City   string
	StartsAt, EndsAt time.Time
	Status           string
	Version          int64
}
type Route struct {
	ID, Origin, Destination, Carrier string
	StartsAt                         time.Time
	Seats, Reserved                  int
	Active                           bool
	Version                          int64
}
type Package struct {
	ID, CampaignID, AgencyID, Name          string
	RouteIDs                                []string
	NightCount, PriceCents, Quota, Reserved int
	Status                                  string
	Version                                 int64
}
type Agreement struct {
	ID, AgencyID, PackageID, Status string
	SignedAt                        *time.Time
	Version                         int64
}
type GroupBooking struct {
	ID, PackageID, AgencyID, GroupName string
	Guests                             int
	Status                             string
	IdempotencyKey                     string
	Version                            int64
}
type Voucher struct {
	ID, BookingID, Number, Status string
	ExpiresAt                     time.Time
	RedeemedAt                    *time.Time
	Version                       int64
}
type HostedVisit struct {
	ID, BookingID, Guide                string
	Status                              string
	PlannedAt, CheckedInAt, ConfirmedAt *time.Time
	Version                             int64
}
type Lead struct {
	ID, AgencyID, Contact, Source, Status, Owner string
	Version                                      int64
}
