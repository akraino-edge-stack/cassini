# image recognition by triton client

import urllib2

TRITON_CLIENT = "http://192.168.0.118:8302/image" # FIX it to ip address of contanier TritonClient in Nvidia Nano

# curl 172.17.0.3:8080/image -T xxx.png
def handle(req):
    headers = {'Content-Type': 'multipart/form-data'}
    req = urllib2.Request(url=TRITON_CLIENT, data=req, headers=headers)
    return urllib2.urlopen(req).read()

if __name__== "__main__":
    handle("test")
