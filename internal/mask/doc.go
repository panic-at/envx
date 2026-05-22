// Package mask renders secret values in a partially-hidden form suitable for
// human-facing output.
//
// A masked value reveals just enough — a short prefix and the length — for a
// person to recognise it, without disclosing the secret itself. The package is
// used by the show and diff commands when printing sensitive variables.
package mask
