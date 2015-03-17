---
RFC: 007
Author: Andrew Gerrand <adg@golang.org>
Status: Accepted
---

# API Stability Policy

## Motivation

The gokit project depends on code maintained by others.
This includes the Go standard library and sub-repositories
and other external libraries.

The Go language and standard library provide stability guarantees, but the
other external libraries typically do not. This RFC proposes a standard policy
for package authors to advertise API stability.

The intention is that the gokit project will require that its dependencies
adhere to the policy, with the greater goal of improving the Go ecosystem
as a whole.

## Scope

This policy is for package authors to provide their users with a promise of API
stability.
This document is similar to and inspired by the [Go 1 compatibility
promise](https://golang.org/doc/go1compat).

An author declaring their package "API Stable" makes the following promise:

> We will not change the package's exported API in backward incompatible ways.
> Future changes to this package will not break dependent code.

### Coverage

The promise of stability includes:

* The package name,
* Exported type declarations and struct fields (names and types),
* Exported function and method names, parameters, and return values,
* Exported constant names and values,
* Exported variable names and values,
* The documented behavior of all exported code.

### Exceptions

* Security. A security issue in the package may come to light whose resolution
  requires breaking compatibility. We reserve the right to address such
  security issues.

* Unspecified behavior. Programs that depend on unspecified
  behavior may break in future releases.

* Bugs. If the package has a bug, a program that depends on the buggy behavior
  may break if the bug is fixed. We reserve the right to fix such bugs.

* Struct literals. For the addition of features it may be necessary to add
  fields to exported structs in the package API. Code that uses unkeyed struct
  literals (such as pkg.T{3, "x"}) to create values of these types would fail
  to compile after such a change. However, code that uses keyed literals
  (pkg.T{A: 3, B: "x"}) will continue to compile after such a change. We will
  update such data structures in a way that allows keyed struct literals to
  remain compatible, although unkeyed literals may fail to compile. (There are
  also more intricate cases involving nested data structures or interfaces, but
  they have the same resolution.) We therefore recommend that composite
  literals whose type is defined in a separate package should use the keyed
  notation.

* Dot imports. If a program imports a package using import . "path", additional
  names later defined in the imported package may conflict with other names
  defined in the program. We do not recommend the use of import . outside of
  tests, and using it may cause a program to fail to compile in the future.

### Breaking compatibility

Should the author wish to break compatibility by redesigning the API the author
should create a new package with a new import path. 

### Awareness

This text should be present in a file named STABILITY in the repository root.

### Enforcement

Tooling may be devised to check the stability of an API as a package evolves,
similar to the api tool used by the Go core.

The [vet](https://godoc.org/golang.org/x/tools/cmd/vet) tool will already
detect "untagged" struct literals; that is, struct literals that will break
when new fields are added to the struct.

## Implementation

To be defined.


