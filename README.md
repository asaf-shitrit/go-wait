# go-wait
[![Tests](https://github.com/asaf-shitrit/go-wait/actions/workflows/run-tests.yml/badge.svg)](https://github.com/asaf-shitrit/go-wait/actions/workflows/run-tests.yml)

A tiny util library for programs that require a little patience ⏰

## Usage

#### Basic
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
#### With explicit options:
```go
checkFunc := func() (bool, error) {
    // any bool based logic that changes over a 
    // given period of time
}

options := &wait.Options{
    timeout: time.Minute
    interval: time.Second
}

ctx := context.Background() // or pass any ctx you would like
if err := wait.Until(ctx, checkFunc, options); err != nil {
    // handle logical/timeout err
}

// logic that should happen after check is satisfied
```

### Options
##### Jitter
Allows you to set an amount of jitter percentage that will apply
for the calculation of each interval.
