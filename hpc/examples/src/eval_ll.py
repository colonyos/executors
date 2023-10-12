import os

from pycolonies import colonies_client

colonies, colonyid, colony_prvkey, executorid, executor_prvkey = colonies_client()

processid = os.getenv("COLONIES_PROCESS_ID")
process = colonies.get_process(processid, executor_prvkey)
print(process)

colonies.set_output(processid, ["hello"], executor_prvkey)
