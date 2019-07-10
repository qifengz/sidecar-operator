# 简介

sidecar-operator是基于[kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)脚手架，利用Kubernetes里的CRD+MutatingWebhookConfiguration机制对指定pod注入任意数量的sidecar容器的一种实现方案。

# 基本思路
假定pod受某个deployment或statefulset controller控制，CR的提交会对选定的某个或某组pod进行delete操作，从而触发了pod重建过程，MutatingWebhookConfiguration在pod recreate前起作用，kube-apiserver调用sidecar-operator webhook server对pod spec进行修改，sidecar的数量和模板都从configmap里读取，最终实现sidecar容器的注入。

主要目录介绍：
- 配置相关
config/crds: CRD的定义
config/cr: 一个CR的例子
config/rbac: rbac相关的yaml文件
config/sample: 一个测试用例，包含测试deployement, sidecar模板所在configmap
config/ssl: kube-apiserver请求sidecar-operator webhook server的SSL证书及生成脚本
config/manager: sidecar-operator本身部署相关的Service, Statefulset, MutatingWebhookConfiguration定义

- 功能模块
pkg/api: CRD接口定义
pkg/controller: CR提交后触发的Reconcile具体逻辑，包括对选定pod进行delete，sidecar configmap的校验，sidecar数量的更新，configmap name的传递等
pkg/webhook: webhook server对pod spec进行修改的具体逻辑  






