package profile

import "sort"

// ChangeKind classifies how a single variable differs between two profiles.
type ChangeKind int

const (
	// Added marks a variable present in the second profile but not the first.
	Added ChangeKind = iota
	// Removed marks a variable present in the first profile but not the second.
	Removed
	// Modified marks a variable present in both profiles with differing
	// definitions.
	Modified
)

// String returns a short human-readable label for the change kind.
func (k ChangeKind) String() string {
	switch k {
	case Added:
		return "added"
	case Removed:
		return "removed"
	case Modified:
		return "modified"
	default:
		return "unknown"
	}
}

// VarChange describes the difference for one variable key between two
// profiles. For an Added change From is the zero Var; for a Removed change To
// is the zero Var.
type VarChange struct {
	Key  string
	Kind ChangeKind
	From Var
	To   Var
}

// DiffResult is the set of variable changes between two profiles, ordered by
// variable key for deterministic output.
type DiffResult struct {
	Changes []VarChange
}

// Empty reports whether the compared profiles had no variable differences.
func (d DiffResult) Empty() bool {
	return len(d.Changes) == 0
}

// Diff compares profiles a and b and returns the changes that transform a's
// variables into b's. Changes are ordered by variable key.
func Diff(a, b Profile) DiffResult {
	keys := make(map[string]struct{}, len(a.Vars)+len(b.Vars))
	for k := range a.Vars {
		keys[k] = struct{}{}
	}
	for k := range b.Vars {
		keys[k] = struct{}{}
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	var res DiffResult
	for _, k := range sorted {
		av, aok := a.Vars[k]
		bv, bok := b.Vars[k]
		switch {
		case aok && !bok:
			res.Changes = append(res.Changes, VarChange{Key: k, Kind: Removed, From: av})
		case !aok && bok:
			res.Changes = append(res.Changes, VarChange{Key: k, Kind: Added, To: bv})
		case av != bv:
			res.Changes = append(res.Changes, VarChange{Key: k, Kind: Modified, From: av, To: bv})
		}
	}
	return res
}
