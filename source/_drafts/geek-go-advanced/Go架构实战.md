---
title: 微服务可用性设计
cover: https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210627104907.png
toc: 1
categories:
  - [Go 进阶训练营]
date: 2021-06-27 10:44:00
tags:
- 可用性设计
---

<br>


-----

<!-- more -->

## 隔离

本质是对系统或资源进行分隔，从而实现当系统发生故障时能限定传播范围和影响范围，即发生故障后只有出问题的服务不可用，保证其他服务仍然可用

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210627115610.png" />

### 服务隔离

- 动静隔离
  - api 等动态资源打到后端服务器，图片等静态资源存放到 oss 等静态存储，并通过 cdn 等加速
  - 表设计，把不常更新的字段和经常更新的字段拆分成多个表。比如用户基本信息+用户统计信息进行拆分
- 读写隔离
  - 主从：数据库一般采用主从模式，读请求打到从库，写请求打到主库
  - CQRS：在编程上，将查询操作和写操作分开。前者只负责返回数据，不做更新，后者只做更新，不返回数据

### 轻重隔离

- 核心隔离
  - 业务按照重要性进行资源池划分。比如订单、账号这种核心业务单独划分资源池，由单独的节点承载。其它业务使用共享资源池
- 快慢隔离
  - 如果对服务统一对待，那么最终的处理速度一般取决于最慢的那个服务
  - 比如日志消费，所有服务日志都放到 Kafka 的一个 topic，当某个消费方本身较慢或资源不足，会导致其他服务的日志消费也受到影响。通常需要结合重要性、部门、业务等对服务进行隔离，减少服务间速率不同带来的负面影响
- 热点隔离
  - 对热点数据进行提前预热，或者广播通知更新的模式，这样用户访问热点数据时，可以通过缓存等方式直接响应，其他数据还是走正常的业务流程

动静隔离
- api 等动态资源打到后端服务器
- 图片、html 等静态资源放到对象存储


### 物理隔离

- 线程隔离
  - 主要通过线程池进行隔离，把业务进行分类交给不同的线程池进行处理，当某个线程池处理一种业务请求发生问题时，不会影响其他线程池的业务
  - 在 Go 中，只需要控制 Goroutine 的总量。Go 协程阻塞并不会阻塞线程
- 进程隔离
  - 通常是容器化的做法。不同服务在不同容器运行，实现进程、网络等的隔离。同时也可以对每个服务的资源进行限制
- 集群隔离
  - 物理上部署多套集群。比如生产集群、测试集群分开部署。甚至可以对核心业务单独分配一个集群

## 超时

超时控制，目的是让组件能够快速失效

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210627122712.png" />

- 应该有公共的默认超时
- 应该有全局的超时传递策略
  - 当前服务的剩余 quota 传递到下游服务，继承超时策略，即超时应该是一个全局策略，需要上下游配合
  - 需要有全局的超时策略，比如上游访问 504 后，下游所有请求应该终止
- 进程内超时控制
  - 一个请求在开始处理前，都要检查是否有足够的 quota，不足则应该 cacel，go 中通过 `context.WithTimeout` 实现
- 服务间超时控制
  - 需要在协议层定义好全局的超时传递策略，比如可以在 headers 传递 context，并带上 quota 等信息，并在基础库层面进行处理

## 限流

限流是定义某个客户或应用可以接收或处理多少个请求的技术。目的是确保应用在自动扩容失效前不会出现过载的情况

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210627122827.png" />

- 单机限流
  - 常用的限流算法：令牌桶、漏桶算法。
  - 常用的限流方式：针对接口限流，针对用户限流（比如 ip），全局限流（对某一类路由进行全局限流）
  - 常见的场景
    - 服务自身对 api 限流，比如 K8s 的 apiServer
    - 通过 api gateway 配置限流，比如 Openresty 提供的 [lua-resty-limit-traffic](https://github.com/openresty/lua-resty-limit-traffic)，或者 Kong 提供的限流插件
  - 单机限流运维很困难，比如我们针对某个接口配置 80/rps 的限流策略，假设是在 kong 配置，如果我们的 kong 有两个节点，则需要配置成 `80/2=40`，不好维护、容易出错（kong 扩缩容）。同时，具体指标无法预知，并且可能会随着迭代而发生变更
- 动态流控
  - 单机限流需要显式配置限流指标，而动态流控控制的是过载计算策略，即触发限流的实际数据是不确定的。比如我们后端有 1 台机器，可能是在 100/rps 时触发限流，而有 2 台机器时，可能是在 150/rps 触发限流。在单机情况下，可能一直都是 90/rps（压测得出的指标）
  - 相对来说，会更加准确，限流配置也更好维护，不受服务迭代影响，也不受节点数影响
  - 实现相对比较复杂，需要主动计算性能数据，计算指标（这一步计算也可能有误差），通常只用于接口限流
- 分布式限流
  - 单机限流需要根据节点手动维护单机配额，分布式限流将这些工作独立成一个模块，用户设置一个全局限额后，由这个模块对限额进行分配。比如设置限流为 100/rps，则当 kong 有一个节点时，全部配额都分配给这个节点，如果有 2 个节点，则各分配 50/rps 的配额
  - 可以对客户端进行抽象，比如 a, b, c 服务都需要某项资源，可以按重要程度，对总资源进行百分比配置
- 配置
  - 接口级别配额分配工作流大，不好维护，可以对接口打上重要性标签，根据标签进行资源分配。比如重要性分为 3 类，则我们只需要维护这三个类比的资源比例
- 其他
  - 客户端一般也要配合流控，比如限制用户的访问频率

## 熔断

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210627133647.png" />

## 降级

降级是有损服务，是最小化影响的一种自我保护策略

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210627133953.png" />

- 自动降级
  - 根据系统情况，触发预定义策略，进行降级。比如可以通过采集动态指标，决定是否触发降级，或者暴力一点，某些情况发生后，手动设置降级
  - 降级模式，降级可以是，直接拒绝某些请求；只提供核心服务，比如双十一支付宝关闭一些配置的修改入口；提供上一次缓存页，老的数据比没有数据好；提供一个降级页，将流量倒到新的服务，比如我们提供一个电商服务，如果确定了服务不可用，可以降级到京东等旗舰店；提前设置默认值，比如热门内容推荐，当推荐服务挂了，可以返回预定义的数据

## 重试

失败重试是一种提高服务稳定性的常见策略，由于网络问题，或后端由于过载等原因返回错误，默认策略一般都是马上重试，这时需要留意流量放大的问题

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210627130441.png" />

通常需要采取以下措施
- 规定全局错误码，当服务端明确失败时，不应该继续重试，并且下游也应该直接放行。需要避免级联重试
- 限制重试次数
- 重试时间最好随机化，并呈指数型递增。比如第一次重试是马上重试，第二次是 2秒后，第三次重试是 4秒后。一般采用退让算法
- 为了应对客户端可能的重试机制，服务应该尽量提供幂等接口
  - 全局唯一 ID，可以定义接口，服务端可以判断请求是否重复处理
  - 去重表，需要根据业务定义，在数据库层面添加依赖，重复则异常
  - 多版本并发控制，在更新的接口中增加一个版本号来做幂等性控制

## 负载均衡

负载均衡主要有两个目的
- 负载分摊，提高服务的能力
- 可用性，避免单点故障

常用方式
- Nginx，最常用的负载均衡网关，配置简单
- K8s，K8s 天然支持负载均衡。并且提供了服务存活检查等完整的配套措施。相对 Nginx，在负载均衡这块，可以说是 几乎不用配置