# get k3s cluster network info

import urllib
import json

AUTH_SERVER = "http://192.168.0.116:8301" # FIX it to ip address of contanier NodeAuthServer

# curl 172.17.0.3:8080/sysinfo
def handle(req):
    # sudo docker inspect 4553499e09de | grep Gateway
    # Use Gateway address to access HOST network
    # can NOT use 127.0.0.1

    # get nodes info
    nodeJson = urllib.urlopen(AUTH_SERVER + "/nodes").read()
    nodeData = json.loads(nodeJson)
    rtn = "--------- Nodes ---------\n"
    rtn += "NAME        IP            STAUTS\n"
    for item in nodeData["items"]:
        metadata = item["metadata"]
        name = metadata["name"]
        ip = metadata["annotations"]["k3s.io/internal-ip"]
        isReady = False
        for info in item["status"]["conditions"]:
            if info["type"] == "Ready" and info["status"] == "True":
                isReady = True
                break
        rtn += '%s %s %s\n' %(name, ip, isReady)

    # get pods info
    podJson = urllib.urlopen(AUTH_SERVER + "/pods").read()
    podData = json.loads(podJson)
    rtn += "--------- Pods ---------\n"
    rtn += "NAME        HostIP        STAUTS\n"
    for item in podData["items"]:
        metadata = item["metadata"]
        name = metadata["name"]
        status = item["status"]
        hostIP = status["hostIP"]
        isReady = False
        for info in status["conditions"]:
            if info["type"] == "Ready" and info["status"] == "True":
                isReady = True
                break
        rtn += '%s %s %s\n' %(name, hostIP, isReady)

    return rtn

if __name__== "__main__":
    handle("test")
