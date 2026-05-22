// Package resolver turns a profile.Var into its final string value.
//
// Every value source — inline literals, host environment variables, 1Password,
// AWS Secrets Manager — implements the Resolver interface and is registered in
// a Registry under the URI scheme it handles. DefaultRegistry returns a
// Registry populated with the built-in resolvers.
//
// ResolveAll fans out across an effective profile, resolving every variable
// concurrently. It is fail-soft: an error on one variable is recorded in the
// result without aborting the others.
package resolver
