---
title: Nginx 应用场景举例
cover: https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210626131623.png
toc: 1
categories:
- [kong]
date: 2021-06-24 23:37:11
tags:
- kong
- nginx
---

<br>

Nginx 是一个 http 服务器、反向代理服务器、邮件代理服务器、通用的 TCP/UDP 反向代理服务器。特点是开源、轻量级，高性能

Nginx 常见使用场景有
- 静态文件服务器
- 反向代理
- 负载均衡
- 动静分离

-----

<!-- more -->

## 写在前面

官方文档初看会很疑惑，我们只需要关注 [NGINX Plus](https://docs.nginx.com/nginx/admin-guide/installing-nginx/installing-nginx-open-source/) 中非 Plus 的文档。Plus 是收费版本，剩下的大多数产品都是配套 Plus 使用，如 NGINX Controller, NGINX App Protect，不用关注；部分是云原生组件，比如 NGINX Ingress Controller，暂时也不用关注

官方文档大概有以下几个

- [official Nginx Wiki -- 偏向场景介绍](https://www.nginx.com/resources/wiki/start/)
- [official Nginx Doc -- 手册类型](https://nginx.org/en/docs/)
- [official Nginx Development -- 内部实现](http://nginx.org/en/docs/dev/development_guide.html)
- [official admin guide --功能介绍](https://docs.nginx.com/nginx/admin-guide/basic-functionality/runtime-control/)

## 安装

> 虚拟机部署，可以参考 [Installing NGINX Open Source](https://docs.nginx.com/nginx/admin-guide/installing-nginx/installing-nginx-open-source/)

我们通过 docker-compose 安装，在[此目录](https://github.com/notFound-lan/blog/tree/master/tool/kong/nginx-docker)下执行 `make start`，服务启动，接着访问服务

```bash
root@823231eac657:/# curl -i localhost:8001
HTTP/1.1 200 OK
Server: nginx/1.19.6
Date: Fri, 25 Jun 2021 00:37:45 GMT
Content-Type: text/html
Content-Length: 4
Last-Modified: Thu, 24 Jun 2021 16:28:06 GMT
Connection: keep-alive
ETag: "60d4b296-4"
Accept-Ranges: bytes

app2
```

通过 `make exec` 命令，可以进到 nginx 内部，查看服务基本情况

```bash
➜  nginx-docker git:(master) ✗ make exec

# 可以看到启动进程及其启动命令
root@bd251f379733:/# ps aux
USER     TAT START   TIME COMMAND
root     Ss+  07:13   0:00 nginx: master process nginx -g daemon off;
nginx    S+   07:13   0:00 nginx: worker process

# 二进制位置
root@bd251f379733:/# which nginx
/usr/sbin/nginx

# 配置文件
root@bd251f379733:/etc/nginx# tree
.
|-- conf.d
|   `-- default.conf
|-- fastcgi_params
|-- koi-utf
|-- koi-win
|-- mime.types
|-- modules -> /usr/lib/nginx/modules
|-- nginx.conf
|-- scgi_params
|-- uwsgi_params
`-- win-utf
```

我们将 `app/nginx` 下的配置文件挂载到了容器，因此在主机修改这些配置，然后重启 Nginx（`make reload`），即可生效；我们将 `app/html` 下的文件挂载到了容器，因此修改这里的文件，可以让响应即时生效

### 常用命令

通过 `nginx -h` 查看

```bash
nginx -V # 查看 nginx 版本和配置信息
nginx -t # 配置文件语法校验
nginx -s reload # 重新加载配置文件。-s 表示给 nginx 发送 signal，包括 stop, quit, reopen, reload
```

## 配置文件

> 完整配置：[Full Example Configuration](https://www.nginx.com/resources/wiki/start/topics/examples/full/) [Another Full Example](https://www.nginx.com/resources/wiki/start/topics/examples/fullexample2/)

Nginx 是一个声明式配置服务器，基本使用方式：**修改配置文件，nginx 加载配置文件，配置生效，nginx 根据配置对请求进行处理**

配置文件基本结构
```conf
http {
  server {
    listen       80;
    server_name  www.example.com;

    location / {
        root   /usr/share/nginx/html;
        index  index.html index.htm;
    }

    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
  }
}
```

- `http` 代表一个 http 服务器，全局只能有一个
- `server` 代表一个虚拟主机。nginx 首先根据 `listen` 指令和请求 ip/port 进行匹配，得到一组 server，接着用 `server_name` 指令内容和 `Host` 请求头进行匹配，得到最终处理请求的 server，如果后者没有匹配项，则使用 default_server
- `location` 路由，请求命中 server 后，需要继续进行路由匹配，比如用户访问 `www.example.com:80/50x.html`，会命中第二个 location
- location 里面定义具体的 upstream，可以是对静态文件的访问，或者反向代理到指定后端，或者负载均衡到多个后端

以上是一个最基本的结构，除此之外，Nginx 通过内置指令和模块，提供了一个 web 服务器和代理服务器所需的大部分功能

## 场景

> PS: 后面的所有访问默认都在容器内执行。通过 `make exec` 进入容器

### 准备工作

创建两个后端

```nginx
# nginx/conf.d/default.conf

server {
    listen       8000;

    location / {
        root   /usr/share/nginx/html/app1;
        index  index.html index.htm;
    }
}

server {
    listen 8001;
    server_name "localhost";
    location / {
        root   /usr/share/nginx/html/app2;
        index  index.html index.htm;
    } 
}
```

以下每个功能，均会创建 `$feat_name.conf` 文件，存放在 `nginx/conf.d` 文件夹下，如下

```bash
  app git:(master) ✗ tree nginx 
nginx
├── conf.d
│   ├── http
│   │   ├── default.conf
│   │   ├── feat
│   │   │   └── backlog.conf
│   │   ├── lb
│   │   │   └── http.conf
│   │   └── web
│   │       ├── reverse_proxy.conf
│   │       └── static_content.conf
│   └── tcp
│       └── default.conf
└── nginx.conf
```

### 静态文件服务器

nginx 可以将用户请求根据 path 映射到服务器的指定目录，nginx 会将对应文件发送给客户端

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/25/20210625082856.png" />

#### 最简版

```nginx
# web/static_content.conf
server {
    listen 5000;

    root /usr/share/nginx/html/app1;

    location / { 
        index index.html;
    }
}
```

#### 学习版

```nginx
# web/static_content.conf
server {
    listen 5000;

    root /usr/share/nginx/html/app1;

    location / { # 其他无法匹配的路由均匹配到这里，nginx 会到 root 配置的目录下查找文件并发送给客户端
        index index.html;
        autoindex on; # 用户可以在浏览器访问目录
        try_files $uri $uri/ /test.html; # 如果访问的 $uri 或 $uri/ 匹配不到文件，则做一个内部重定向到 /test.html

        # 默认数据传输需要先把数据拷贝到缓存。打开这选项后，可以实现两个 io 的直接数据传输
        sendfile on;
        sendfile_max_chunk 1m; # 限制一次 sendfile() 调用最大的传输数据
    }

    location /images/ {
        try_files $uri $uri/ @backend; # 匹配不到，还可以 proxy 到 @backend 服务
    }

    location @backend {
        proxy_pass http://10.0.0.2;
    }

    location ~ \.(mp3|mp4) { # mp3 和 mp4 会打到 /www/media 目录
        root /www/media;
    }
}
```

访问结果：
```bash
root@012f0d7efee9:~# curl -i  localhost:5000/test.html
HTTP/1.1 200 OK
Server: nginx/1.19.6
Date: Sat, 26 Jun 2021 11:25:10 GMT
Content-Type: text/html
Content-Length: 4
Last-Modified: Thu, 24 Jun 2021 17:07:20 GMT
Connection: keep-alive
ETag: "60d4bbc8-4"
Accept-Ranges: bytes

app1
```

### 反向代理

反向代理，指用户本来要访问指定的服务，但因为我们在前面架了个 nginx（同时域名也解析到了 Nginx 所在的机器），导致用户流量被 nginx 接收，nginx 处理后，再将流量转发到后端服务，最后按原路返回。此时用户并不能感知到中间 Nginx 的存在

反向代理的一般场景
- 在公网和内部服务建立一道屏障，方便服务的管理和实施一些安全措施
- 负载均衡
- 隐藏内部细节，从而以统一的方式给用户提供不同服务的内容
- 协议转换，比如对外提供 https 服务，转换为 http 后转发给后端。比如 nginx 支持将 http 转为 FastCGI、memchached 等协议

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/25/20210625083430.png" />

#### 最简版

```nginx
server {
    listen 4000;
    location / {
        proxy_pass http://127.0.0.1:8001;
    }
}
```

通过 proxy_pass 可以将流量转发到 `127.0.0.1:8001` 

#### 学习版

```nginx
server {
    listen 4000;
    location /path {
        # 如果 proxy 是由后面跟着 uri，，它将替代前面的 /path
        # 比如 localhost/path/t.html -> localhost/child/t.html
        # 如果不以 uri 结尾，则使用用户访问路由打到后端
        proxy_pass http://127.0.0.1:8001/child/; 
    }

    # 协议转换
    location ~ \.php$ {
        # 将请求转发给 FastCGI 服务器
        fastcgi_pass  localhost:9000; 
        fastcgi_param SCRIPT_FILENAME
                      $document_root$fastcgi_script_name;
        include       fastcgi_params;

        # 将请求转发给 memcached
        # memcached_pass localhost:11211;
        # memcached_read_timeout 60s;
    }

    # 静态文件不太好获取 proxy 后的数据，这里我们在主机起一个服务，用来查看请求信息
    # 服务在 nginx-docker/app/backend/app1 -> go run .
    # curl -i  localhost:4000/foo
    location /foo {
        # 1. Host 默认被重置为 $proxy_host，即下面的 docker.for.mac.host.internal:8080
        # 2. Connection 默认被重置为 close
        # 3. 同时会删除值为空字符的 header
        # 可以通过 proxy_set_header 修改这些行为
        proxy_pass http://docker.for.mac.host.internal:8080;

        # 修改转发给后端的 header
        # 重置后变成 localhost
        proxy_set_header Host $host; 
        proxy_set_header X-Real-IP $remote_addr;
        # 如果不想要 header 被转发，可以将其置为空字符串
        proxy_set_header Accept "";

        # 缓存配置
        # 默认情况下，nginx 会缓存后端响应数据，直到获得完整数据再发送给前端
        # 控制缓存是否打开，默认是打开。如果关闭，则后端响应数据会同步发送给客户端
        proxy_buffering on; 
        # number size 分别为缓冲区数据量和缓冲区大小，总大小就是他们的乘积
        proxy_buffers 16 4k;
        # 响应头的缓冲区大小，默认为一个缓冲区的大小，即上面的 size
        proxy_buffer_size 2k;
    }
}
```


### 动静分离

本质讲也是反向代理。动态请求，比如 api 请求匹配到一个 location，反向代理到后端 api 服务；静态请求，一般指静态文件，比如 html、图片，匹配一个 location，打到本地文件，或文件服务器，或 Redis

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/25/20210625083228.png" />

### 负载均衡

后端一个副本容易导致单点故障，或者无法支撑日常流量，因此，线上服务一般是多副本模式，通过 nginx，可以用不同的负载均衡算法将用户请求打到后端某个节点

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/25/20210625083404.png" />

#### HTTP 

##### 最简版

```nginx

upstream backend {
    server 127.0.0.1:8000;
    server 127.0.0.1:8001;
}

server {
    listen 3000;
    location / {
        proxy_pass http://backend;
    }
}
```

##### 学习版

```nginx
upstream backend {
    # 默认使用 Round Robin 算法
    # least_conn; # 请求将会转发给最少活跃连接的服务器
    # ip_hash;
    # hash $request_uri consistent;  
    # random;
    server 127.0.0.1:8000 weight=1; # 权重越高。命中的概率越大。默认是 1
    server 127.0.0.1:8001 weight=5;
    # server 10.0.0.1 backup; # 当上面两个都挂了，打到这个服务器
    # server 10.0.2.1 down;
    # 会话保持，plus 功能。开源版本使用 hash 或 ip_hash
    # sticky cookie srv_id expires=1h domain=.example.com path=/;
}

server {
    listen 3000;
    location / {
        proxy_pass http://backend;
    }
}
```

#### TCP 

采用 tcp 协议，配置上和 http 类似，包括负载均衡。

一些前置工作
- 容器内启动 tcp 服务：`nc -v -l -p 8999`
- 启动 tcp 客户端 `nc 127.0.0.1 9000` 就可以通过 nginx 和内部的 tcp 服务通信了

```最简版
# 现在容器内启动一个 tcp 服务：nc -v -l -p 8999
# 更新当前配置后，通过 9000 端口就可以和上述 tcp 服务通信
# 启动一个客户端连接到 9000 端口：nc 127.0.0.1 9000
server {
    listen 9000;
    proxy_pass 127.0.0.1:8999;
}
```

### 其他重要参数 

#### 压缩

#### backlog

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/25/20210625090901.png" />

- 同步队列默认为 128 `/proc/sys/net/ipv4/tcp_max_syn_backlog`
- 消费队列默认为 128 `/proc/sys/net/core/somaxconn`

## 参考
- [official Nginx Doc -- 手册类型](https://nginx.org/en/docs/)
- [official Nginx Wiki -- 偏向场景介绍](https://www.nginx.com/resources/wiki/start/)
- [agentzh's Nginx Tutorizals](https://openresty.org/download/agentzh-nginx-tutorials-en.html)
- [official Nginx Development -- 实现](http://nginx.org/en/docs/dev/development_guide.html)
- [tengine taobao nginx docs](https://tengine.taobao.org/nginx_docs/cn/docs/http/request_processing.html)
- [official admin guide -- 功能介绍](https://docs.nginx.com/nginx/admin-guide/basic-functionality/runtime-control/)