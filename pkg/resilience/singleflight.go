package resilience

import "golang.org/x/sync/singleflight"

// Group wraps singleflight.Group for convenience.
// Use to deduplicate concurrent identical outbound requests.
//
// Usage:
//
//	val, err, _ := sf.Do("key", func() (any, error) {
//	    return expensiveCall()
//	})
type Group = singleflight.Group
