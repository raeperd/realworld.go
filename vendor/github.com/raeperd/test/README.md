# test.go
[![Go Reference](https://pkg.go.dev/badge/github.com/raeperd/test.svg)](https://pkg.go.dev/github.com/raeperd/test) [![.github/workflows/build.yaml](https://github.com/raeperd/test/actions/workflows/build.yaml/badge.svg)](https://github.com/raeperd/test/actions/workflows/build.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/raeperd/test)](https://goreportcard.com/report/github.com/raeperd/test) [![codecov](https://codecov.io/gh/raeperd/test/graph/badge.svg?token=UCZDF4EIXD)](https://codecov.io/gh/raeperd/test)  
Package test is the minimalist testing helper for Go.

Forked from [earthboundkid/be](https://github.com/earthboundkid/be), Inspired by [Mat Ryer](https://github.com/matryer/is) and [Alex Edwards](https://www.alexedwards.net/blog/easy-test-assertions-with-go-generics).

## Features
- Simple, readable and typesafe test assertions using generics
- Single file without dependencies for simple copy and paste into your project
- Fail fast by default but easily switch to relaxed with `test.Relaxed(t)`
- Extend project specific test functions in `test` module when using copy-paste approach

## Installation
You can use this package in two ways:

### 1. go module
```sh
go get github.com/raeperd/test
```

### 2. Copy and Paste
1. Copy the contents of [test.go](./test.go) into your project. (e.g. internal/test/test.go)
2. This file contains all the necessary code for the package and can be used without any dependencies.
3. (optional)[relaxed.go](./relaxed.go) and [debug.go](./debug.go) is optional for niche use cases.

## Usage

### If installed as module
```go
import "github.com/raeperd/test"

func TestExample(t *testing.T) {
    want := 1
    test.Equal(t, want, 2-1)
}
```

### If copied as internal test package
```go
import "yourproject/internal/test"

func TestExample(t *testing.T) {
    want := 1
    test.Equal(t, want, 2-1)
}
```

## Available Assertions

### Equal/NotEqual
```go
test.Equal(t, "hello", "world")     // bad
test.Equal(t, "goodbye", "goodbye") // good

test.NotEqual(t, "hello", "world")     // good
test.NotEqual(t, "goodbye", "goodbye") // bad
```

### DeepEqual
```go
test.DeepEqual(t, map[int]bool{1: true, 2: false}, map[int]bool{1: true, 2: false}) // good
test.DeepEqual(t, nil, []int{})                                                     // bad

s := []int{1, 2, 3}
test.DeepEqual(t, []int{1, 2, 3}, s) // good
test.DeepEqual(t, []int{3, 2, 1}, s) // bad
```

### Nil/NotNil
```go
var err error
test.Nil(t, err)    // good
test.NotNil(t, err) // bad

err = errors.New("(O_o)")
test.Nil(t, err)    // bad
test.NotNil(t, err) // good
```

### Contains/NotContains
```go
type mytype string
var mystring mytype = "hello, world"
test.Contains(t, "hello, world", "world") // good
test.Contains(t, mystring, "world")       // good
test.Contains(t, mystring, "World")       // bad
test.Contains(t, []int{1, 2, 3, 4, 5}, 3) // good
test.Contains(t, []int{1, 2, 3, 4, 5}, 6) // bad

test.NotContains(t, "hello, world", "World") // good
test.NotContains(t, mystring, "World")       // good
test.NotContains(t, mystring, "world")       // bad
test.NotContains(t, []int{1, 2, 3, 4, 5}, 6) // good
test.NotContains(t, []int{1, 2, 3, 4, 5}, 3) // bad
```

### Test anything else:
```go
test.True(t, o.IsValid())
test.True(t, len(pages) >= 20)
```

## Philosophy
Tests usually should not fail. When they do fail, the failure should be repeatable. Therefore, it doesn't make sense to spend a lot of time writing good test messages. (This is unlike error messages, which should happen fairly often, and in production, irrepeatably.) Package test is designed to simply fail a test quickly and quietly if a condition is not met with a reference to the line number of the failing test. If the reason for having the test is not immediately clear from context, you can write a comment, just like in normal code. If you do need more extensive reporting to figure out why a test is failing, use `test.DebugLog` or `test.Debug` to capture more information.

Most tests just need simple equality testing, which is handled by `test.Equal` (for comparable types), and `test.DeepEqual` (which relies on `reflect.DeepEqual`). Another common test is that a string or byte slice should contain or not some substring, which is handled by `test.In` and `test.NotIn`. Rather than package be providing every possible test helper, you are encouraged to write your own advanced helpers for use with `test.True`, while package be takes away the drudgery of writing yet another simple `func nilErr(t *testing.T, err) { ... }`.

Every tool in the be module requires a `testing.TB` as its first argument. There are various [clever ways to get the testing.TB implicitly](https://dave.cheney.net/2019/12/08/dynamically-scoped-variables-in-go), but package be is designed to be simple and explicit, so it's easiest to just always pass in a testing.TB the boring way.
