

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
&{17004 [0xc2080dce20 0xc2080d94c0 0xc2080dcee0 0xc2080dcf20] 17224}
```

Output:
```
{"answer":42,"question":"what is the meaning of life?"}
```


#### PrefixLogger Example

Code:
```go
&{17357 [0xc2080dd040 0xc2080d9780 0xc2080dd100 0xc2080dd140] 17571}
```

Output:
```
question=what is the meaning of life? answer=42
```



