# exec_until

Simple CLI tool that executes the given command until the given timeout or a pattern is matched. When the timeout is reached or the pattern is matched, the process running the command will be terminated.

# Example usage

## Basic explanation

```
❯ ./exec_until
Error: Pattern flag is required.
  -p string
        Pattern to match (required)
  -t duration
       Timeout duration in seconds (default 3s)
```

## Minimal examples

```
❯ ./exec_until -p="hi" echo hi
pattern matched: hi

❯ ./exec_until -p="ha" echo hi
command completed without a match

❯ ./exec_until -p="hi" sleep 10
Error: timeout reached, pattern not matched
```

## Real-life example

I use `way-displays` to manage my displays, but I don't want the program to keep running in the background. I only launch it on-demand with a specific configuration. Therefore I created this command to help me terminate the program once it's correctly configured my current setup.

```
./exec_until -p="Changes successful" -t=2s way-displays -c ~/.config/way-displays/cfg.dock.yaml
```
