# log-transfer

## Introduction

This is a log transfer project, which can read logs from k8s pod and transfer them to anywhere by websocket.

![image-20231102222957634](https://wanz-bucket.oss-cn-beijing.aliyuncs.com/typora/image-20231102222957634.png)

## Usage

> Run Server

Min go version is **1.21**, since this project uses `slog`.

It's recommended to build this project from source code. The code is extremely simple, so you can adjust it to your own needs.

If you run it in **k8s**, you need to **link a service account** to the pod in order to read the logs of other pods.

If you run it **locally**, you need to **specify the kubeconfig file path** in `pkg/k8s.go`.

In `pkg/model.go`, you can config the namespace **white list** for safety. It's a regular expression list. By default, it's `".*"`, which
means **all namespaces are allowed.**

```go
var allowNamespaceRegList = []string{
        ".*",
}
```

> Connect to Server

You can use any websocket client to connect to the server. The server will send logs to the client.

To get logs from the server, you need to send a message to the server first. The message is a **json string**, which contains the namespace, pod name, mode and tail lines. The mode can be `pod` or `job`.

The tail lines is the number of lines you want to get from the end of the log file for the first time, which cannot be 0 or greater than the
limit set in the server, default is 100.

```json
{
  "namespace": "default",
  "pod": "nginx-7db9fccd9f-4q9q2",
  "mode": "pod",
  "tailLines": 100
}
```

> example client code

```typescript
const socket = new WebSocket("localhost:8080"); // 创建一个 Socket 实例
let data: any = {
  mode: "pod",
  namespace: "default",
  workload: "nginx-7db9fccd9f-4q9q2",
  tailLines: 100,
};

socket.onopen = function () {
  socket.send(JSON.stringify(data)); // 发送数据
};

//
socket.onmessage = function (msg) {
  console.log(msg);
};
```
