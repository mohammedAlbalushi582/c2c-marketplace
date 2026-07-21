// Package presence tracks who is currently on the site, in memory. Each visitor
// browser sends a heartbeat (session id + current path) every ~20s; an entry is
// "online" until its TTL lapses without another heartbeat. Nothing is persisted —
// this answers "who's here right now", not historical analytics.
package presence

import (
	"sort"
	"sync"
	"time"
)

// Visitor is one active browser session.
type Visitor struct {
	SessionID string
	UserID    *int64
	Name      string
	IsUser    bool
	Path      string
	IP        string
	UserAgent string
	LastSeen  time.Time
}

// Tracker holds live sessions keyed by session id, pruning stale ones on read.
type Tracker struct {
	mu       sync.Mutex
	visitors map[string]Visitor
	ttl      time.Duration
}

func NewTracker(ttl time.Duration) *Tracker {
	return &Tracker{visitors: make(map[string]Visitor), ttl: ttl}
}

// Touch records/refreshes a heartbeat. Identity fields (UserID/Name/IsUser) are
// carried from the incoming beat so a user who logs in mid-session is reflected.
func (t *Tracker) Touch(v Visitor) {
	if v.SessionID == "" {
		return
	}
	v.LastSeen = time.Now()
	t.mu.Lock()
	t.visitors[v.SessionID] = v
	t.mu.Unlock()
}

// Active returns the online visitors, most-recent first, pruning expired ones.
func (t *Tracker) Active() []Visitor {
	cutoff := time.Now().Add(-t.ttl)
	t.mu.Lock()
	out := make([]Visitor, 0, len(t.visitors))
	for id, v := range t.visitors {
		if v.LastSeen.Before(cutoff) {
			delete(t.visitors, id)
			continue
		}
		out = append(out, v)
	}
	t.mu.Unlock()
	sort.Slice(out, func(i, j int) bool { return out[i].LastSeen.After(out[j].LastSeen) })
	return out
}

// Count is the number of online visitors.
func (t *Tracker) Count() int {
	return len(t.Active())
}
