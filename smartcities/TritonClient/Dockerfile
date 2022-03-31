# triton python client require python 3.6.9
FROM python:3.6.9

# Copy the entire TritonClient and build it
COPY . /home/TritonClient/
WORKDIR /home/TritonClient/
# install require package of python3
RUN pip3 install -i https://pypi.tuna.tsinghua.edu.cn/simple --upgrade pip \
	&& pip3 install -i https://pypi.tuna.tsinghua.edu.cn/simple Pillow \
	&& pip3 install -i https://pypi.tuna.tsinghua.edu.cn/simple --upgrade wheel setuptools cython \
	&& pip3 install -i https://pypi.tuna.tsinghua.edu.cn/simple --upgrade grpcio-tools numpy==1.19.4 future attrdict \
	&& pip3 install -i https://pypi.tuna.tsinghua.edu.cn/simple --upgrade ./tritonclient-2.17.0-py3-none-manylinux2014_aarch64.whl[all]

CMD ["python", "/home/TritonClient/main.py"]
