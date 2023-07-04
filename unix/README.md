# Introduction
This executor runs Unix commands. 

## Usage
```console
cat ./ls.json
```

```json
{
    "conditions": {
        "executortype": "unix"
    },
    "funcname": "ls",
    "args": [
        "-al"
    ]
}
```

```console
colonies function submit --spec ./ls.json
```

or altenatively:
```console
colonies function exec --func ls --args "-al" --targettype unix
```
