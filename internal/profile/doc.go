// Package profile defines the core domain model for envx: profiles and the
// environment variables they contain, plus operations to merge and diff them.
//
// A Profile is a named set of Var definitions and may inherit from a parent via
// Extends. A Var is either a literal value or a reference URI resolved later by
// the resolver package.
//
// This package performs no I/O: loading, saving and validation live in the
// config package, which imports this one.
package profile
