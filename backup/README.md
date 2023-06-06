# PostgreSQL Backup Executor

The **backup executor** provides functionality to backup a PostgreSQL database to a S3 bucket. The executor may run as a Kubernetes sidecar.

## Usage
cat ./backup.json
```json
{
    "conditions": {
        "executortype": "backup"
    },
    "funcname": "backup"
}
```

```console
colonies function submit --spec ./backup.json
```

or altenatively:
```console
colonies function exec --func backup --targettype backup 
```

### Listening backup processes
```console
colonies process pss --type backup --count 5 

+------------------------------------------------------------------+----------+------+---------------------+---------------+
|                                ID                                | FUNCNAME | ARGS |      END TIME       | EXECUTOR TYPE |
+------------------------------------------------------------------+----------+------+---------------------+---------------+
| c0de2617bd2a09e9ac913443cec6dac1b9fbc8d1fa665b91fb46ba5e79dff4e8 | backup   |      | 2023-06-04 17:24:04 | backup        |
| 2b1e857a4fe561f727f27282059e49f5cef69d95fc24a0c6ccedca106dda8c49 | backup   |      | 2023-06-04 17:24:38 | backup        |
| 4367b24faa690374a18f901d567b667164d21d94858688979d259b147c892aea | backup   |      | 2023-06-04 17:25:25 | backup        |
| 9a8cf91e19ac86f4bc316708b2c8d19420a137359b73fef7658d3c386b70925b | backup   |      | 2023-06-04 17:53:50 | backup        |
| 0e92ed5da7fbf36d1f8dc3408cda5afb68b4d010bbdcb09a7b65da2055525ff1 | backup   |      | 2023-06-04 18:45:57 | backup        |
| 57d7694f26d887db297486b3b70b7139127e5a5f2ac02c40ab548e70849b8fa9 | backup   |      | 2023-06-04 18:56:11 | backup        |
+------------------------------------------------------------------+----------+------+---------------------+---------------+
```

### Get info about a particular process
```console
colonies process get -p  57d7694f26d887db297486b3b70b7139127e5a5f2ac02c40ab548e70849b8fa9 --out
```

```json
{
    "filename": "backup_1685897746.tar.gz",
    "bucket": "backup",
    "filepath_s3": "/backups/backup_1685897746.tar.gz",
    "exectime_backup": 19,
    "size": 171446794,
    "size_s3": 171446794,
    "exectime_upload_s3": 6
}
```

### Schedule cron processes 
```console
cat ./cron.json
```
```json
[
    {
        "nodename": "run_backup",
        "funcname": "backup",
        "conditions": {
            "executortype": "backup"
        }
    }
]
```

Run backup every night at 03:00.

```console
colonies cron add --name backup_cron --cron "0 0 3 * * *" --spec cron.json 
```

## Starting a backup executor
```console
 source devenv
./bin/backup_executor
```

## Configuration
The configuration are done using environmental variables. This should be self-explanatory, except for *FULL_BACKUPS*, which specifies how many (full) backup files should be stored on S3 before being deleted, e.g if set to 3, only 3 backup files will be stored. The oldest files are automatically purged. 

If *COLONIES_COLONY_PRVKEY*, then the executor will self-register to the Colonies server. Otherwise, it will use *COLONIES_EXECUTOR_ID* and *COLONIES_EXECUTOR_PRVKEY* to connect the colony.

```console
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8
export LC_ALL=en_US.UTF-8
export LC_CTYPE=UTF-8
export TZ=Europe/Stockholm
export COLONIES_TLS="false"
export COLONIES_SERVER_HOST="localhost"
export COLONIES_SERVER_PORT="50080"
export COLONIES_COLONY_ID="4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4"
export COLONIES_COLONY_PRVKEY="ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514"
export COLONIES_EXECUTOR_ID=""
export COLONIES_EXECUTOR_PRVKEY=""
export FULL_BACKUPS="3"
export BACKUP_PATH="/tmp"
export AWS_S3_SECURE="true"
export AWS_S3_INSECURE_SKIP_VERIFY="false"
export AWS_S3_ENDPOINT="s3.host
export AWS_S3_REGION=""
export AWS_S3_ACCESS_KEY="accesskey"
export AWS_S3_SECRET_ACCESS_KEY="secret"
export AWS_S3_BUCKET_NAME="backup"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_DATABASE="postgres"
export DB_USER="postgres"
export DB_PASSWORD="password"
```
