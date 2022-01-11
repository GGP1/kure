// Package pb parses protocol buffers with the "proto3" version.
//
// IMPORTANT: When parsing .proto files (at least with "protoc") the output
// generated contains structs with "omitempty" json tags, when Card and Entry
// shouldn't, as empty fields won't be printed when editing them. The easiest
// solution is to change them manually, until a better one is found.
package pb
