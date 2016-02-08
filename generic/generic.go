// Package generic defines a placeholder `Generic` type that is used to ensure type-safety for
// packages using gogen.
package generic

// Generic represents an arbitrary generic type. Intended to be underly arbitrary generic types:
//  type T generic.Generic
type Generic interface{}

// Type-checker needs this to get the Generic type.
var _ Generic
