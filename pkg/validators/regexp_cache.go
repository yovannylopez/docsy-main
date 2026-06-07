package validators

import (
	"fmt"
	"regexp"
	"sync"
)

// compiledRegexpCache memoizes patterns used in validation (e.g. StringWithPattern, custom Phone)
// to avoid recompiling on each request.
var compiledRegexpCache sync.Map // string -> *regexp.Regexp

func cachedRegexp(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty pattern")
	}
	if re, ok := compiledRegexpCache.Load(pattern); ok {
		return re.(*regexp.Regexp), nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	compiledRegexpCache.Store(pattern, re)
	return re, nil
}
