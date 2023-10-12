import os

from pycolonies import Colonies
from pycolonies import colonies_client
from pycolonies import func_spec

colonies, colonyid, colony_prvkey, executorid, executor_prvkey = colonies_client()

colonies_server = os.getenv("COLONIES_SERVER_HOST")
colonies_port = os.getenv("COLONIES_SERVER_PORT")
colonies_tls = os.getenv("COLONIES_SERVER_TLS")
processid = os.getenv("COLONIES_PROCESS_ID")
print("colonies_server:", colonies_server)
print("colonies_port:", colonies_port)
print("colonies_tls:", colonies_tls)
print("colonyid:", colonyid)
print("colony_prvkey:", colony_prvkey)
print("processid:", processid)
print("executor_prvkey:", executor_prvkey)
colonies.set_output(processid, ["hello"], executor_prvkey)
