package domain

import (
	"fmt"
	"strings"
	"time"
)

func RequireID(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidState, name)
	}
	if len(value) > 96 {
		return fmt.Errorf("%w: %s is too long", ErrInvalidState, name)
	}
	return nil
}

func ValidateActor(actor Actor) error {
	if err := RequireID("actor.user_id", actor.UserID); err != nil {
		return err
	}
	if !ValidRole(actor.Role) {
		return fmt.Errorf("%w: invalid role", ErrForbidden)
	}
	if err := RequireID("actor.request_id", actor.RequestID); err != nil {
		return err
	}
	return nil
}

func RequireRole(actor Actor, roles ...Role) error {
	if err := ValidateActor(actor); err != nil {
		return err
	}
	for _, role := range roles {
		if actor.Role == role {
			return nil
		}
	}
	return fmt.Errorf("%w: role %s cannot perform operation", ErrForbidden, actor.Role)
}

func ValidateWindow(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return fmt.Errorf("%w: time window is required", ErrInvalidState)
	}
	if !end.After(start) {
		return fmt.Errorf("%w: window end must follow start", ErrInvalidState)
	}
	return nil
}

func ValidatePage(page Page) Page {
	if page.Limit <= 0 {
		page.Limit = 50
	}
	if page.Limit > 200 {
		page.Limit = 200
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	if page.Direction != "asc" && page.Direction != "desc" {
		page.Direction = "asc"
	}
	return page
}
