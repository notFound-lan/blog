---
title: kong 入门
cover: 'https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/20210623002530.png'
toc: 1
categories:
  - [kong]
date: 2021-06-23 22:54:30
tags:
  - kong
---

<br>

Kong 是一个云原生、高效的、可扩展的、分布式的 API 网关，我们使用 Kong，通常是基于以下目的
- 统一流量入口，Kong 集群暴露一个公网 ip，所有外部请求经过这个 ip，打到内部服务
- 安全管控、开发提效。Kong 是一个插件式平台，所有接入 Kong 的应用通过开启对应插件，获得如，日志、认证、限流等功能

这是 Kong 系列文章的第一篇，本篇文章将会介绍 Kong 的基础知识，利用 Docker 搭建一个本地 Kong 服务，以及，通过 Konga 对 Kong 集群进行管理～

-----

<!-- more -->

## 为什么需要 Kong

早期我们的服务可能是这样的

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624000110.png" />

两个服务部署在两个虚拟机上，每个虚拟机有一个公网 IP，域名分别解析到这两个 IP，用户通过域名访问。

再具体到功能上，是这样的

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/23/20210623233843.png" />

这种架构存在的问题很明显
- 每个服务都需要公网 IP，每个服务的域名解析都独立进行，各种底层的运维服务复杂且不好管理
- 每个服务都需要各自实现一些基础功能，比如日志、认证、限流、监控
- 当服务达到一定量级时，基本都不可管理、不可维护

第一个问题可以通过在服务前挂一个 Nginx，然后通过 Nginx 的反向代理来解决，第二个问题也可以通过 Nginx + 模块的模式实现，但实现复杂，维护困难，不具备实践价值。但是，理论上，Nginx 具备解决以上问题的底层能力，因此有了两个演进的产品或工具
- [OpenResty](https://openresty.org/cn/installation.html)：OpenResty 是[春哥](https://github.com/agentzh)主导和开源的项目，主要包括三部分：1. 一个完整的 Nginx 服务器；2. [lua-nginx-module](https://github.com/openresty/lua-nginx-module)，是一个标准的 Nginx模块，它将 lua 嵌入 nginx，提供了 lua 编写 nginx 插件的能力；3. 基于第二项，提供了很多常用的 lua 插件。意味着，当我们安装 OpenResty，它就具备了 Nginx 的所有能力，有很多常用的开箱即用的 lua 插件，并且支持通过编写 lua 脚本对 Nginx 进行扩展
- Kong：Kong 是基于 OpenResty 的产品，更进一步，对 `nginx.conf` 进行抽象，用户可以动态修改，同时通过 adminAPI 对外提供 restful 接口，支持通过 webAPI 动态更新路由。同时，将插件平台化，提供了众多开箱即用的插件，并支持自定义插件

用 Kong 替换前面的 Nginx，变成

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/23/20210623235509.png" />

以上问题得到解决
- Kong 集群对外暴露一个公网 ip，所有服务域名都解析到这个 IP，流量到达 Kong 后，通过反向代理打到对应的后端服务，后端服务没有其他公网入口
- 通过 Kong 插件平台，可以按需开启各种功能插件，获得基础能力

## Kong 安装

我们利用 docker-compose，搭建一个 Kong 服务，并通过 Konga 进行管理。

### 组件

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624002505.png" />

Kong：核心组件，流量入口
Konga：一个非官方但很好用的 Kong 管理后台，通过 adminAPI 和 Kong 进行交互
PG：Kong 数据库，也支持 MySQL 等其他类型数据库。用来存储路由数据，以及其他各种元数据

### 安装

在这个[目录下](https://github.com/notFound-lan/blog.github.io/tree/master/source/_drafts/kong/docker) ，执行以下命令

```bash
make start-db # 启动 pg
make migration # 执行 schema migrate
make start # 启动 kong 和 konga
```

接着就可以访问 konga `http://localhost:1338/register#!/dashboard`

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624004753.png" />

填写密码，并登陆，进入 konga 后，填写 kong 连接配置

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624005007.png" />

连接成功，konga 基础界面

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624005145.png" />

### docker-compose

简单介绍一下安装脚本
```bash
kong:
  ports:
      - 8010:8000 # 流量入口
      - 8011:8001 # adminAPI http 服务
```

`8001` 是 kong 默认的 http adminAPI，我们将其映射到主机的 8011 端口，因此你可以直接在主机访问 `localhost:8011`。`8000` 是默认的流量入口，我们将其映射到主机的 8010 端口，因此你可以直接在主机访问 `localhost:8010`，此时没有配置任何路由，因此将会获得 Kong 的无路由响应

## 配置一个服务

### 目标

通过 konga 配置一个服务，本质上是配置一个 nginx.conf，我们先看一个简单的 nginx.conf 文件

```nginx
http {
  server { # simple reverse-proxy
    listen       80;
    server_name  www.example.com;

    # serve static files
    location /nginx  {
      proxy_pass      https://www.nginx.com;
    }

    # pass requests for dynamic content to rails/turbogears/zope, et al
    location / {
      proxy_pass      https://www.voiddme.cc;
    }
  }
}
```

含义
- 配置了一个 server，当用户访问域名为 `www.example.com` 时，命中这个配置
- 配置了两个 router，当访问路由以 `/nginx` 开头，打到 nginx 官方网站；其他路由，打到我们的博客

### 创建后端服务

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624011541.png" />
<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624011608.png" />

他们是 proxy_pass 的抽象，指定后端地址（这里也可以指定为 upstream，通过 upstream 对应多个 target，实现负载均衡）

### 创建路由 

nginx-service 路由
<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624011734.png" />

blog-demo 路由
<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624011848.png" />

他们是 location 的抽象 

### 访问

注意先配置 hosts `127.0.0.1 www.example.com` 
- 通过访问 `http://www.example.com:8010/nginx`，可以将流量打到 nginx 官网
- 通过访问 `http://www.example.com:8010/` 可以将流量打到我的博客
- 访问 `http://www.example2.com:8010/` 由于没有匹配规则，因此得到 kong 的 404 响应

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624083643.png" />

### upstream 方式

```nginx
http {
  # 提供负载均衡能力
  upstream big_server_com {
    server cn.bing.com weight=5;
    server www.baidu.com weight=5;
  }

  server { # simple load balancing
    listen          80;
    server_name     www.example2.com; 
    access_log      logs/big.server.access.log main;

    location / {
      proxy_pass      http://big_server_com;
    }
  }
}
```

以上配置映射到 kong 配置
- 创建一个 upstream 对象 big_server_com
- `big_server_com` 下，创建 2 个 Target，分别映射到 2 个后端
- 创建一个 service 对象，指向上述 upstream
- 创建默认路由，匹配该域名下的所有流量
- 效果，用户访问 `www.example2.com:8010`，流量会负载到以上两个 target

service
<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624084511.png" />
route
<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624084611.png" />
upstream，其中的 name 配置在 service
<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624084637.png" />
target
<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624084712.png" />

流量走向
<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/24/20210624084831.png" />

### 坑

- 深灰色的配置，代表你填写后需要回车，否则保存将会失败

## 总结

API 网关是应用达到一定规模后的必备基础设施，而 Kong 是性价比较高的一个选项。

学习 Kong 时，需要配合 Nginx，理解当前的修改，最终是利用了 Nginx 的哪项底层能力
- Kong 各种对象，及其配置，可以映射到 nginx.conf
- Kong 的插件实现及自定义开发，依托于 OpenResty 提供的 Lua 能力，而底层是基于 nginx 的请求 phrase
- Kong 的各种功能，比如反向代理、负载均衡，本质也是体现到 nginx.conf

Konga 利用 adminAPI 管理 Kong 集群，也就是我们也可以直接调用 adminAPI，但有 UI 界面明显更利于入门和后续的维护

## 参考

- [Kong Gateway docs](https://docs.konghq.com/gateway-oss/)
- [Kong github](https://github.com/Kong/kong)
- [konga github](https://github.com/pantsel/konga)