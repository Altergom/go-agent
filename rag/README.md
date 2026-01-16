# RAG知识库构建

## 一、Milvus启动

### RAG知识库依赖于Milvius服务，启动前请确认Milvus运行正常(注：Milvuis运行于Docker)

### Linux

`curl -sfL https://raw.githubusercontent.com/milvus-io/milvus/master/scripts/standalone_embed.sh -o standalone_embed.sh`

`bash standalone_embed.sh start`

### Windows(注：PowerShell执行)

`Invoke-WebRequest https://raw.githubusercontent.com/milvus-io/milvus/refs/heads/master/scripts/standalone_embed.bat -OutFile standalone.bat`

`.\standalone.bat start`

## 二、召回检索设计

本项目基于框架开发，召回器和检索器使用的是Eino封装后的实例。**请务必自行根据Eino源码手动敲一遍检索召回**，明白如何封装？封装了什么？

## 三、节点编排

本项目所有编排均为图编排，在实践前请自行学习Chain和Graph的区别，了解有向无环图(DAG),了解链式编排局限性。
