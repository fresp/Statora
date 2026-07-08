package handlers

import "net/http"

type OriginPolicy struct {
	allowed map[string]struct{}
}

func NewOriginPolicy(origins []string) OriginPolicy {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}
	return OriginPolicy{allowed: allowed}
}

func (p OriginPolicy) Allow(origin string) bool {
	_, ok := p.allowed[origin]
	return ok
}

func (p OriginPolicy) AllowRequest(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	return p.Allow(origin)
}
