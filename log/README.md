

# package log

    import "/Users/harlow/code/go/src/github.com/go-kit/kit/log"

## Rationale

Package log provides basic interfaces for structured logging.

The fundamental interface is Logger. Loggers create log events from
key/value data.



## Usage


#### JSONLogger Example

Code:
```go
&{17004 [0xc2080dd060 0xc2080df5c0 0xc2080dd120 0xc2080dd160] 17224}
```

Output:
```
{"answer":42,"question":"what is the meaning of life?"}
```


#### PrefixLogger Example

Code:
```go
&{17357 [0xc2080dd280 0xc2080df880 0xc2080dd340 0xc2080dd380] 17571}
```

Output:
```
question=what is the meaning of life? answer=42
```



