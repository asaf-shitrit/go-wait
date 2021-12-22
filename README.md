# go-wait
[![Tests](https://github.com/asaf-shitrit/go-wait/actions/workflows/run-tests.yml/badge.svg)](https://github.com/asaf-shitrit/go-wait/actions/workflows/run-tests.yml)

A tiny util library for programs that require a little patience ⏰

## Usage
### Wait
Useful for simple cases where a predictable `bool` check should be ran
in a set interval and continue when it is satisfied.
#### Basic usage
```go
checkFunc := func() (bool, error) {
    // any bool based logic that changes over a 
    // given period of time
}

ctx := context.Background() // or pass any ctx you would like
if err := wait.Until(ctx, checkFunc); err != nil {
    // handle logical/timeout err
}

// logic that should happen after check is satisfied
```
#### With explicit options
```go
checkFunc := func() (bool, error) {
    // any bool based logic that changes over a 
    // given period of time
}

options := &wait.UntilOptions{
    timeout: time.Minute
    interval: time.Second
}

ctx := context.Background() // or pass any ctx you would like
if err := wait.Until(ctx, checkFunc, options); err != nil {
    // handle logical/timeout err
}

// logic that should happen after check is satisfied
```

### Backoff
Really useful in cases that low CPU overhead a constraint and the check intervals 
should be backed off after each run.

It was inspired by Go's own `http.Server` `Shutdown` implementation ❤️

#### Basic usage
```go
checkFunc := func() (bool, error) {
    // any bool based logic that changes over a 
    // given period of time
}

ctx := context.Background() // or pass any ctx you would like
if err := wait.Backoff(ctx, checkFunc); err != nil {
    // handle logical/timeout err
}

// logic that should happen after check is satisfied
```

#### With explicit options
```go
checkFunc := func() (bool, error) {
    // any bool based logic that changes over a 
    // given period of time
}

options := &wait.BackoffOptions{
	baselineDuration: time.Millisecond,
	limit:            500 * time.Millisecond,
	multiplier:       2,
}

ctx := context.Background() // or pass any ctx you would like
if err := wait.Backoff(ctx, checkFunc, options); err != nil {
    // handle logical/timeout err
}

// logic that should happen after check is satisfied
```

### Capabilities
### Timeout & Cancel ⏰
It is aligned with Golang concept of context so explicit cancels & timeout will work
out of the box.
#### Jitter
Allows you to set an amount of jitter percentage that will apply
for the calculation of each interval.
