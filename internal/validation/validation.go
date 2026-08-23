package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var idPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{1,63}$`)

func ID(v string) error {
	if !idPattern.MatchString(v) {
		return fmt.Errorf("invalid id")
	}
	return nil
}
func Required(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}
func Length(name, value string, min, max int) error {
	n := utf8.RuneCountInString(value)
	if n < min || n > max {
		return fmt.Errorf("%s length out of range", name)
	}
	return nil
}
func URL(value string) error {
	u, e := url.Parse(value)
	if e != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid url")
	}
	return nil
}
func Enum(name, value string, allowed ...string) error {
	for _, v := range allowed {
		if value == v {
			return nil
		}
	}
	return fmt.Errorf("%s has invalid value", name)
}
func All(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
