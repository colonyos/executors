from pycolonies import Colonies
from pycolonies import func_spec
from pycolonies import Workflow

colonies = Colonies("colonies.colonyos.io", 443, True)

colonyid = "4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4"
executorid = "3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac" 
executor_prvkey = "ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05"

wf = Workflow(colonyid)
pods = 1 
executors_per_pod = 2 
container_image = "colonyos/sleepexecutor:v0.0.1"
f = func_spec(func="deploy", 
              args=["sleep-executor", pods, executors_per_pod, False, container_image], 
              colonyid=colonyid, 
              executortype="k8s",
              priority=1,
              maxexectime=-1,
              maxretries=-1,
              maxwaittime=-1)
wf.add(f, nodename="deploy", dependencies=[])

deps = []
for i in range(10):
    f = func_spec(func="sleep",
                  args=["1000"],
                  colonyid=colonyid,
                  executortype="sleep",
                  priority=1,
                  maxexectime=-1,
                  maxretries=-1,
                  maxwaittime=-1)
    nodename = "sleep-"+str(i)
    deps.append(nodename)
    wf.add(f, nodename=nodename, dependencies=["deploy"])

f = func_spec(func="undeploy",
              args=["sleep-executor"],
              colonyid=colonyid,
              executortype="k8s",
              priority=1,
              maxexectime=-1,
              maxretries=-1,
              maxwaittime=-1)
wf.add(f, nodename="undeploy", dependencies=deps)

processgraph = colonies.submit(wf, executor_prvkey)
print("Workflow", processgraph["processgraphid"], "submitted")
