package taskpolicy

import (
	"fmt"
	"sort"
	"strings"
)

type Explanation struct {
	Subject     string
	Allowed     bool
	Summary     string
	Obligations []string
	Metadata    map[string]string
}

func Explain(decision Decision, metadata map[string]string) Explanation {
	copyMetadata := make(map[string]string, len(metadata))
	for key, value := range metadata {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		copyMetadata[key] = strings.TrimSpace(value)
	}
	obligations := append([]string(nil), decision.Obligations...)
	sort.Strings(obligations)
	summary := "district task operation is allowed"
	if !decision.Allowed {
		summary = fmt.Sprintf("district task operation rejected: %s", decision.Reason)
	}
	return Explanation{Subject: "district task", Allowed: decision.Allowed, Summary: summary, Obligations: obligations, Metadata: copyMetadata}
}

func Merge(primary, secondary Decision) Decision {
	if !primary.Allowed {
		return cloneDecision(primary)
	}
	if !secondary.Allowed {
		return cloneDecision(secondary)
	}
	seen := make(map[string]struct{}, len(primary.Obligations)+len(secondary.Obligations))
	obligations := make([]string, 0, len(primary.Obligations)+len(secondary.Obligations))
	for _, list := range [][]string{primary.Obligations, secondary.Obligations} {
		for _, obligation := range list {
			if _, exists := seen[obligation]; exists {
				continue
			}
			seen[obligation] = struct{}{}
			obligations = append(obligations, obligation)
		}
	}
	sort.Strings(obligations)
	return Decision{Allowed: true, Code: "allowed", Obligations: obligations}
}

func cloneDecision(input Decision) Decision {
	return Decision{Allowed: input.Allowed, Code: input.Code, Reason: input.Reason, Obligations: append([]string(nil), input.Obligations...)}
}

func ObligationSet(decision Decision) map[string]bool {
	result := make(map[string]bool, len(decision.Obligations))
	for _, obligation := range decision.Obligations {
		result[obligation] = true
	}
	return result
}
