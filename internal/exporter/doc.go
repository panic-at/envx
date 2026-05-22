// Package exporter serializes a resolved set of environment variables into
// formats consumable by other tools: dotenv files, JSON objects and POSIX
// shell scripts.
//
// Each format is a named Exporter retrieved from the registry with Get; All
// lists the supported formats. Every exporter writes its keys in ascending
// lexical order, so output is byte-for-byte deterministic and diff-friendly.
package exporter
