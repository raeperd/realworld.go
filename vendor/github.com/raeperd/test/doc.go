// Package test is a minimalist test assertion helper library.
//
// # Philosophy
//
// Tests usually should not fail. When they do fail, the failure should be repeatable. Therefore, it doesn't make sense to spend a lot of time writing good test messages. (This is unlike error messages, which should happen fairly often, and in production, irrepeatably.) Package be is designed to simply fail a test quickly and quietly if a condition is not met with a reference to the line number of the failing test. If the reason for having the test is not immediately clear from context, you can write a comment, like normal code. If you do need more extensive reporting to figure out why a test is failing, use [test.DebugLog] or [test.Debug] to capture more information.
//
// Most tests just need simple equality testing, which is handled by [test.Equal] (for comparable types), and [test.DeepEqual] (which relies on [reflect.DeepEqual]). Another common test is that a string or byte slice should contain or not some substring, which is handled by [test.In] and [test.NotIn]. Rather than package be providing every possible test helper, you are encouraged to write your own advanced helpers for use with [test.True], while package be takes away the drudgery of writing yet another simple func nilErr(t *testing.T, err) { ... }.
//
// Every test in the be package requires a [testing.TB] as its first argument. There are various clever ways to get the testing.TB implicitly,[*] but package be is designed to be simple and explicit, so it's easiest to just always pass in a testing.TB the boring way.
//
// [*]: https://dave.cheney.net/201/12/08/dynamically-scoped-variables-in-go
package test
