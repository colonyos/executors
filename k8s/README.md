# Introduction
The sleep executor simply sleeps for a certain amount of time. The sleep function takes a string as argument stating how long time to sleep in milliseconds. 

## Usage
```console
cat ./sleep.json
```

```json
{
    "conditions": {
        "executortype": "sleep"
    },
    "funcname": "sleep",
    "args": [
        "1000"
    ]
}
```

```console
colonies function submit --spec ./sleep.json
```

or altenatively to sleep for 20 seconds:
```console
colonies function exec --func sleep --args 20000 --targettype sleep 
```
