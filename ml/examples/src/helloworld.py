import os

path = "/tmp/helloworld/result"
isExist = os.path.exists(path)
if not isExist:
   os.makedirs(path)

with open("/tmp/helloworld/result/helloworld.txt", 'w') as f:
    f.write("helloworld 4")

print("Helloworld 4!!!\n")
