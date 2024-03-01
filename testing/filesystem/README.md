# Testing things that work with filesystems

This section is all about how we can test things that interact with filesystems!

## General principals

- Rather than test against an actual filesystem, we want to test against something that mimics a
  filesystem.
- We should use [dependency
  injection](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/dependency-injection) to
  write code that works both against a real filesystem or against something that looks like a
  filesystem.

## Libraries used

- https://github.com/spf13/afero - To act as a shim for our filesystem.
- https://github.com/stretchr/testify - A wrapper that makes testing a little easier.

## Notes

The way I did dependency injection isn't super idiomatic. Ideally I'd have a struct that has a field
of type `afero.Fs`, then a method on the struct that does the appending, plus a wrapper method that
defaults the field with an `afero.OsFs` for ease of use.