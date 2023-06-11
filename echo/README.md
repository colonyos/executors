# Introduction
The echo executor just closes with process with its input string. 

## Usage
```console
cat ./echo.json
```

```json
{
    "conditions": {
        "executortype": "echo"
    },
    "funcname": "echo",
    "args": [
        "helloworld"
    ]
}
```

```console
colonies function submit --spec ./echo.json
```

or altenatively:
```console
colonies function exec --func echo --args helloworld --targettype echo
```
