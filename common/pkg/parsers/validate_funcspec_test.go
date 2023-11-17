package parsers

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestValidateFuncSpec(t *testing.T) {

	///////////////////////////////////////////////////////////////////
	/// Test case: processes-per-node is missing
	///////////////////////////////////////////////////////////////////

	funcSpecStr := `
{
    "conditions": {
        "executortype": "dev-hpcexecutor",
        "walltime": 60
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "python3",
        "docker-image": "python:3.12-rc-bookworm",
        "rebuild-image": false,
        "args": [
            "/cfs/src/{processid}/helloworld.py"
        ]
    },
    "maxexectime": 10,
    "maxretries": 3
}
`
	funcSpec, err := core.ConvertJSONToFunctionSpec(funcSpecStr)
	assert.Nil(t, err)

	err = ValidateFuncSpec(funcSpec)
	assert.NotNil(t, err) // processes-per-node cannot be 0

	///////////////////////////////////////////////////////////////////
	/// Test case: nodes is missing
	///////////////////////////////////////////////////////////////////

	funcSpecStr = `
{
    "conditions": {
        "executortype": "dev-hpcexecutor",
        "processes-per-node": 1,
        "walltime": 60
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "python3",
        "docker-image": "python:3.12-rc-bookworm",
        "rebuild-image": false,
        "args": [
            "/cfs/src/{processid}/helloworld.py"
        ]
    },
    "maxexectime": 10,
    "maxretries": 3
}
`
	funcSpec, err = core.ConvertJSONToFunctionSpec(funcSpecStr)
	assert.Nil(t, err)

	err = ValidateFuncSpec(funcSpec)
	assert.NotNil(t, err) // nodes is missing

	///////////////////////////////////////////////////////////////////
	/// Test case: mem is missing
	///////////////////////////////////////////////////////////////////

	funcSpecStr = `
{
    "conditions": {
        "executortype": "dev-hpcexecutor",
        "nodes": 1,
        "processes-per-node": 1,
        "walltime": 60
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "python3",
        "docker-image": "python:3.12-rc-bookworm",
        "rebuild-image": false,
        "args": [
            "/cfs/src/{processid}/helloworld.py"
        ]
    },
    "maxexectime": 10,
    "maxretries": 3
}
`
	funcSpec, err = core.ConvertJSONToFunctionSpec(funcSpecStr)
	assert.Nil(t, err)

	err = ValidateFuncSpec(funcSpec)
	assert.NotNil(t, err) // mem is missing

	///////////////////////////////////////////////////////////////////
	// Test case: cpu is missing
	///////////////////////////////////////////////////////////////////

	funcSpecStr = `
{
    "conditions": {
        "executortype": "dev-hpcexecutor",
        "nodes": 1,
        "processes-per-node": 1,
        "mem": "1000Mi", 
        "walltime": 60
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "python3",
        "docker-image": "python:3.12-rc-bookworm",
        "rebuild-image": false,
        "args": [
            "/cfs/src/{processid}/helloworld.py"
        ]
    },
    "maxexectime": 10,
    "maxretries": 3
}
`

	///////////////////////////////////////////////////////////////////
	/// Test case: walltime is missing
	///////////////////////////////////////////////////////////////////

	funcSpec, err = core.ConvertJSONToFunctionSpec(funcSpecStr)
	assert.Nil(t, err)

	err = ValidateFuncSpec(funcSpec)
	assert.NotNil(t, err) // cpu is missing

	funcSpecStr = `
{
    "conditions": {
        "executortype": "dev-hpcexecutor",
        "nodes": 1,
        "processes-per-node": 1,
        "mem": "1000Mi",
        "cpu": "100m"
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "python3",
        "docker-image": "python:3.12-rc-bookworm",
        "rebuild-image": false,
        "args": [
            "/cfs/src/{processid}/helloworld.py"
        ]
    },
    "maxexectime": 10,
    "maxretries": 3
}
`
	funcSpec, err = core.ConvertJSONToFunctionSpec(funcSpecStr)
	assert.Nil(t, err)

	err = ValidateFuncSpec(funcSpec)
	assert.NotNil(t, err) // walltime is missing

	///////////////////////////////////////////////////////////////////
	/// Test case: maxexectime is missing
	///////////////////////////////////////////////////////////////////

	funcSpec, err = core.ConvertJSONToFunctionSpec(funcSpecStr)
	assert.Nil(t, err)

	err = ValidateFuncSpec(funcSpec)
	assert.NotNil(t, err) // cpu is missing

	funcSpecStr = `
{
    "conditions": {
        "executortype": "dev-hpcexecutor",
        "nodes": 1,
        "processes-per-node": 1,
        "mem": "1000Mi",
        "cpu": "100m",
        "walltime": 100
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "python3",
        "docker-image": "python:3.12-rc-bookworm",
        "rebuild-image": false,
        "args": [
            "/cfs/src/{processid}/helloworld.py"
        ]
    },
    "maxexectime": 10,
    "maxretries": 3
}
`
	funcSpec, err = core.ConvertJSONToFunctionSpec(funcSpecStr)
	assert.Nil(t, err)

	err = ValidateFuncSpec(funcSpec)
	assert.Nil(t, err) // maxexectime is missing

	///////////////////////////////////////////////////////////////////
	/// Test case: maxretries is missing
	///////////////////////////////////////////////////////////////////

	funcSpecStr = `
{
    "conditions": {
        "executortype": "dev-hpcexecutor",
        "nodes": 1,
        "processes-per-node": 1,
        "mem": "1000Mi",
        "cpu": "100m",
        "walltime": 100
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "python3",
        "docker-image": "python:3.12-rc-bookworm",
        "rebuild-image": false,
        "args": [
            "/cfs/src/{processid}/helloworld.py"
        ]
    },
    "maxexectime": 10
}
`
	funcSpec, err = core.ConvertJSONToFunctionSpec(funcSpecStr)
	assert.Nil(t, err)

	err = ValidateFuncSpec(funcSpec)
	assert.NotNil(t, err) // maxretries is missing

	///////////////////////////////////////////////////////////////////
	/// Test case: walltime < maxexectime
	///////////////////////////////////////////////////////////////////

	funcSpecStr = `
{
    "conditions": {
        "executortype": "dev-hpcexecutor",
        "nodes": 1,
        "processes-per-node": 1,
        "mem": "1000Mi",
        "cpu": "100m",
        "walltime": 100
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "python3",
        "docker-image": "python:3.12-rc-bookworm",
        "rebuild-image": false,
        "args": [
            "/cfs/src/{processid}/helloworld.py"
        ]
    },
    "maxexectime": 1000,
    "maxretries": 3 
}
`
	funcSpec, err = core.ConvertJSONToFunctionSpec(funcSpecStr)
	assert.Nil(t, err)

	err = ValidateFuncSpec(funcSpec)
	assert.NotNil(t, err) // walltime must be > maxe

	///////////////////////////////////////////////////////////////////
	/// Test case: correctly formated funcspec
	///////////////////////////////////////////////////////////////////

	funcSpecStr = `
{
    "conditions": {
        "executortype": "dev-hpcexecutor",
        "nodes": 1,
        "processes-per-node": 1,
        "mem": "1000Mi",
        "cpu": "100m",
        "walltime": 100
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "python3",
        "docker-image": "python:3.12-rc-bookworm",
        "rebuild-image": false,
        "args": [
            "/cfs/src/{processid}/helloworld.py"
        ]
    },
    "maxexectime": 10,
    "maxretries": 3 
}
`
	funcSpec, err = core.ConvertJSONToFunctionSpec(funcSpecStr)
	assert.Nil(t, err)

	err = ValidateFuncSpec(funcSpec)
	assert.Nil(t, err) // Finally OK!
}
