# exec_until

Simple CLI tool that executes the given command until the given timeout or a pattern is matched. When the timeout is reached or the pattern is matched, the process running the command will be terminated.

# Example usage

## Basic explanation

```
❯ ./exec_until -h
usage: ./exec_until [flags] <command>

flags:
  -p string
        Pattern to match (required)
  -t duration
        Timeout duration in seconds. Set to 0 for no timeout. (default 3s)
```

## Minimal examples

```
❯ ./exec_until -p "foo bar" echo foo bar
pattern matched: foo bar

❯ ./exec_until -p "bar foo" echo foo bar
Error: command completed without a match

❯ ./exec_until -p "bar foo" sleep 5
Error: timeout reached, pattern not matched
```

## Real-life example

I use [`way-displays`](https://github.com/alex-courtis/way-displays) to manage my displays, but I don't want the program to keep running in the background. I only launch it on-demand with a specific configuration. Therefore I created this command to help me terminate the program once it's correctly configured my current setup.

```
./exec_until -p="Changes successful" way-displays -c ~/.config/way-displays/cfg.dock.yaml
```

# TODO

- (fix) Currently known issue, sometimes tests error, concurrency issue(?) 
- (feat) Implement verbose flag and hide some logging
- (fix) Currently some issues with passing arguments, i.e. if you pass `exec_until -p "hi" mycommand | grep something`, command `mycommand | grep something` is not correctly executed. Current workaround: `exec_until -p "hi" "mycommand | grep something"`
- (fix) Currently user lose zsh suggestions and syntax highlighting on command passed to CLI app
