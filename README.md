# Base Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![golang](https://img.shields.io/badge/golang-v1.17.13-brightgreen)](https://go.dev/doc/install)
[![version](https://img.shields.io/badge/version-v0.3.2-green)]()

Base Operator 项目可以处理产品实体和权限实体从提供者到目标服务的同步。

对于产品实体，Base Operator 提供了一组用于调谐 Product Provider 资源事件的 Controller，调谐内容主要是监听 Product Provider 资源声明的产品数据提供者（目前只支持 GitLab）中的产品实体的变化，并将变化内容同步至租户管理集群。

对于权限实体，Base Operator 提供了一组用于调谐 Base Provider 资源事件的 Controller，调谐内容主要是监听 Base Provider 资源声明的基础数据提供者（目前只支持 GitLab）中的用户、成员等实体的变化，并将变化内容同步至目标服务（如 Nexus）中。

## 功能简介

### 同步产品实体

由于 Product Provider 是 Nautes 所必须的资源，因此 Controller 只在 Product Provider 资源存在的情况下才可以正常工作。

Nautes 会为每个产品实体创建一个元数据代码库（默认名称为 default.project），并在此代码库中维护该产品的所有子实体的资源。

当 Controller 监听到产品数据提供者新增了产品实体，同时监听到该产品实体存在有效的元数据代码库时，会将产品实体的声明转为 Product 资源、将元数据代码库的声明转为 CodeRepo 资源并写入租户管理集群的管理命名空间（默认名称为 nautes ）内，同时根据 Product 资源名称在租户管理集群中创建对应的产品命名空间，然后创建一个源为元数据代码库、目标为产品命名空间的 ArgoCD Application 资源，该 ArgoCD Application 资源会触发将元数据代码库中的所有子实体的资源向产品命名空间进行同步的动作。所有被同步至租户管理集群中的产品实体相关的资源，会被其他管理组件监听并调谐，用于后续的代码库、制品库、运行时的维护。

当 Controller 监听到产品数据提供者删除了某个存在有效元数据代码库的产品实体或删除了某个产品实体的元数据代码库时，会在租户管理集群中删除产品实体对应的 ArgoCD Application 资源、 Product 资源、CodeRepo 资源、以及产品命名空间。

### 同步权限实体

开发中……

## 快速开始

### 准备

安装以下工具，并配置 GOBIN 环境变量：

- [go](https://golang.org/dl/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

准备一个 kubernetes 实例，复制 kubeconfig 文件到 {$HOME}/.kube/config

### 构建

```shell
go mod tidy -go=1.16 && go mod tidy -go=1.17
go build -o manager main.go
```

### 运行

```shell
./manager
```

### 单元测试

安装 Envtest

```shell
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
setup-envtest use 1.21.x
export KUBEBUILDER_ASSETS=$HOME/.local/share/kubebuilder-envtest/k8s/1.21.4-linux-amd64
```

安装 Vault

```shell
wget https://releases.hashicorp.com/vault/1.10.4/vault_1.10.4_linux_amd64.zip
unzip vault_1.10.4_linux_amd64.zip
sudo mv vault /usr/local/bin/
```

安装 Ginkgo

```shell
go install github.com/onsi/ginkgo/v2/ginkgo@v2.3.1
```

执行单元测试

```shell
ginkgo -r
```
