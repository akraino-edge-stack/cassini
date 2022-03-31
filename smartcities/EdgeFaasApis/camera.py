# image recognition by triton client

import urllib

CAMERA_CLIENT = "http://192.168.0.118:8303/camera" # FIX it to ip address of contanier CameraClient in Nvidia Nano

# curl 172.17.0.3:8080/camera -o capture.jpg
def handle(req):
    data = urllib.urlopen(CAMERA_CLIENT).read()
    return data

if __name__== "__main__":
    handle("test")
