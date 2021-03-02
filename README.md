# gpu_process_exporter

在研究Prometheus + Grafana监控主机GPU时，发现nvidia官方提供的exporter只有总的数据，并没有每个进程占用的显存数据，之前也没有用过go语言，就自己边摸索边用go写了个gpu进程的exporter

主要是解析 **_nvidia-smi -q -x_** 命令生成的xml格式，转换输出自定义的Prometheus格式，端口号：**9102**
GPU_EXPORTER{GPU="GPU0",Pid="1818",Type="C",ProcessName="python"} 5097
GPU_EXPORTER{GPU="GPU0",Pid="",Type="",ProcessName="Total"} 15079
GPU_EXPORTER{GPU="GPU0",Pid="",Type="",ProcessName="Used"} 12435
GPU_EXPORTER{GPU="GPU0",Pid="",Type="",ProcessName="Free"} 2644

编译构建
**_go build -v gpu_process_exporter.go_**

执行
**_./gpu_process_exporter_**

查看
**http://ip:9102/metrics**

集成进Grafana的效果图
![image](https://user-images.githubusercontent.com/12092975/109594336-91939a80-7b4d-11eb-9c2c-f5e9c5d14b5b.png)

