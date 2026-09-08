package runtime

import (
	"errors"
	"fmt"
	"sort"
)

// ErrNoAdapter reports a provider token this build cannot execute.
var ErrNoAdapter = errors.New("no runtime adapter for provider")

// adapters maps a provider token to its adapter. This map is the entire cost of
// adding a provider: one entry, and nothing above it changes.
var adapters = map[string]func() Adapter{
	"claude-subscription": func() Adapter { return &claudeAdapter{} },
}

// New returns the adapter for a provider token, or an error wrapping ErrNoAdapter.
func New(provider string) (Adapter, error) {
	factory, ok := adapters[provider]
	if !ok {
		return nil, fmt.Errorf("%w %q", ErrNoAdapter, provider)
	}
	return factory(), nil
}

// Providers returns every provider token with an adapter, sorted.
func Providers() []string {
	names := make([]string, 0, len(adapters))
	for name := range adapters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
