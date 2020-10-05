package internal

import (
	"context"
	"reflect"
	"strings"
)

type DefaultPatternMatcher struct {
}

func NewDefaultPatternMatcher() Matcher {
	return new(DefaultPatternMatcher)
}

func (dpm *DefaultPatternMatcher) Match(ctx context.Context, event *Event, pattern interface{}) bool {
	if pattern == nil {
		return true
	}
	ispartof, _ := contains(ctx, event.Payload, pattern, 0)
	return ispartof
}

// contains is a helper function to check if b is contained in a (if b is a partial of a)
func contains(ctx context.Context, a interface{}, b interface{}, strength int) (bool, int) {
	if a == nil && b == nil {
		return true, strength
	}
	if a == nil || b == nil {
		return false, strength
	}
	switch b.(type) {
	case string:
		if b.(string) == "*" && a != nil {
			return true, strength
		}
	}
	switch b.(type) {
	case map[string]interface{}:
		switch a.(type) {
		case map[string]interface{}:
			for k, vb := range b.(map[string]interface{}) {
				c, s := contains(ctx, a.(map[string]interface{})[k], vb, 0)
				if !c {
					return false, strength
				}
				strength += s
			}
		case []interface{}:
			return false, strength
		default:
			return false, strength
		}
	case []interface{}:
		switch a.(type) {
		case map[string]interface{}:
			return false, strength
		case []interface{}:
			// check if all fields in b are in a (in any order)
			for _, vb := range b.([]interface{}) {
				present := false
				for _, va := range a.([]interface{}) {
					// this supports partial matches in deeply structured array elements even with wild cards etc.
					c, s := contains(ctx, va, vb, 0)
					if c {
						strength += s
						present = true
						break
					}
				}
				if !present {
					return false, strength
				}
			}
		default:
			return false, strength
		}
	default:
		switch a.(type) {
		case map[string]interface{}:
			return false, strength
		case []interface{}:
			return false, strength
		default:
			switch b.(type) {
			// special treatment of flat string matches supporting boolean or and wild cards
			case string:
				alts := strings.Split(b.(string), "||")
				result := false
				for _, alt := range alts {
					if alt == "*" || alt == a {
						strength++
						result = true
					} else if reflect.TypeOf(a).Kind() == reflect.String && strings.Contains(alt, "*") {
						frags := strings.Split(alt, "*")
						result = true
						partial := a.(string)
						// check corner cases at beginning and end of string
						if !strings.HasPrefix(alt, "*") && !strings.HasPrefix(partial, frags[0]) {
							return false, strength
						}
						if !strings.HasSuffix(alt, "*") && !strings.HasSuffix(partial, frags[len(frags)-1]) {
							return false, strength
						}
						for _, frag := range frags {
							// making sure all fragments occur in the string in correct order without overlaps
							if !strings.Contains(partial, frag) {
								result = false
								break
							} else {
								idx := strings.Index(partial, frag)
								partial = partial[idx+len(frag):]
							}
						}
						if result {
							strength++
						}
					}
				}
				if !result {
					return false, strength
				}
			default:
				equal := a == b
				if equal {
					strength++
				}
				return equal, strength
			}
		}
	}
	return true, strength
}
