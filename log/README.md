

# package log

    import "/Users/harlow/code/go/src/github.com/go-kit/kit/log"

Package log provides basic interfaces for structured logging.

The fundamental interface is Logger. Loggers create log events from
key/value data.



## Rationale

TODO

## Usage


#### JSONLogger Example

Code:
```go
&{17004 [0xc2080dce50] 17102}
```

Output:
```
{"meaning_of_life":42}
```


#### PrefixLogger Example

Code:
```go
&{17235 [0xc2080dcfd0] 17395}
```

Output:
```
question=what is the meaning of life? answer=42
```



