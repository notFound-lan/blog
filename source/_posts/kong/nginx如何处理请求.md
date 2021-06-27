---
title: Nginx 如何处理一个请求
cover: 'https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210626131623.png'
toc: 1
categories:
  - [kong]
date: 2021-06-26 11:26:43
tags:
- nginx
---

<br>

> 翻译：[How nginx processes a request](https://nginx.org/en/docs/http/request_processing.html)
> written by Igor Sysoev
> edited by Brian Mercer

这是 Nginx 作者写的一篇文章，主要阐述了 nginx 如何处理用户请求，我阅读过程中顺手给翻译了，提前总结下
- 首先根据 `listen` 指令匹配请求 ip 和 port，可以匹配多个 server
- 接着根据 `server_name` 指令匹配请求头 `host`，如果没有匹配的 server，则使用 `default_server`，`default_server` 可以显示指定，或者由 nginx 默认指定
- 确定 server 后，接着进行路由匹配，`location` 指令和请求 URI 匹配，如果匹配成功，根据 location 内的配置进行处理，否则默认返回 404

-----

<!-- more -->

## 基于名字的虚拟主机

nginx 会先判断将请求交给哪个 server 处理。我们从下面这个简单配置开始，它有三个虚拟主机，均监听 `*.80` 

```nginx
server {
    listen      80;
    server_name example.org www.example.org;
    ...
}

server {
    listen      80;
    server_name example.net www.example.net;
    ...
}

server {
    listen      80;
    server_name example.com www.example.com;
    ...
}
```

由于监听端口都相同，因此只需要根据 server_name 进行匹配。Nginx 将 `server_name` 和请求头 `Host` 进行匹配，如果没有找到匹配的 Server，或者请求就没带 `Host` 头，则 nginx 会将它路由到 default_server。在上述配置中，Nginx 默认会取第一个 server 作为 default_server。当然，我们也可以在 `listen` 指令中使用 `default_name` 参数显示指定：

```nginx
server {
  listen 80 default_server;
  server_name example.net www.example.net;
}
```

> `default_server` 自 0.8.21 版本可用。之前的版本请使用 `deault` 参数 

注意 deault_server 是 `listen` 指令的属性，而非 `server_name`

## 如何避免处理 Host 未定义的请求

如果不允许未携带 `Host` 的请求，可以通过以下配置丢弃这些请求：

```nginx
server {
  listen 80;
  server_name "";
  return 444;
}
```

以上配置将 server_name 设置为空字符串，它将会匹配 80 端口上所有未带 `Host` 请求头的流量，并返回一个特殊的 nginx 非标准状态码 444 来关闭连接。

> 从 0.8.48 开始，这是一个默认配置，无需我们手动配置。在更早的版本，采用主机的 hostname 作为 default server 的 server_name

## 基于名称和ip的混合虚拟服务器

> 译注：简单来说，就是同时配置 listen 和 server_name 指令

下面是一个稍微复杂的配置，这里的 server 监听了不同的地址

```nginx
server {
    listen      192.168.1.1:80;
    server_name example.org www.example.org;
    ...
}

server {
    listen      192.168.1.1:80;
    server_name example.net www.example.net;
    ...
}

server {
    listen      192.168.1.2:80;
    server_name example.com www.example.com;
    ...
}
```

根据以上配置，Nginx 先根据请求 ip 和端口，与 `listen` 指令进行匹配，接着根据 `Host` 请求头和 `server_name` 指令进行匹配。如果无法找到匹配的 `server_name`，则会用 default_server。比如，访问 `www.example.com`，打到 `192.168.1.1:80`，以上第一个 server 和第二个 server 均匹配成功，接着利用 Host 进行匹配，可以看到，两个 server 均不满足要求，因此选择 default_server，即第一个 server 进行处理

如前所述，default_server 是 listen 指令的参数，相同的 listen 只能有一个 deault_server

```nginx 
server {
    listen      192.168.1.1:80;
    server_name example.org www.example.org;
    ...
}

server {
    listen      192.168.1.1:80 default_server;
    server_name example.net www.example.net;
    ...
}

server {
    listen      192.168.1.2:80 default_server;
    server_name example.com www.example.com;
    ...
}
```

## 一个简单的 php 网站配置

下面是一个简单的 php 网站配置，让我们一起看下 Nginx 如何选择正确的 location 来处理网站的请求：

```nginx
server {
    listen      80;
    server_name example.org www.example.org;
    root        /data/www;

    location / {
        index   index.html index.php;
    }

    location ~* \.(gif|jpg|png)$ {
        expires 30d;
    }

    location ~ \.php$ {
        fastcgi_pass  localhost:9000;
        fastcgi_param SCRIPT_FILENAME
                      $document_root$fastcgi_script_name;
        include       fastcgi_params;
    }
}
```

请求进来时，Nginx 会根据以下步骤来判定最终由哪个路由进行处理
- 先用前缀匹配找出匹配度最高的 location，此时不考虑配置顺序。以上配置仅有一个前缀匹配路由：`/`，它可以匹配任何路由，被放到最后进行匹配
- 按配置顺序查找正则匹配项。当命中第一个正则表达式时，查找将会停止，匹配到的 location 将被应用。如果所有正则都不匹配，则会使用前面找出的前缀匹配路由

注意，所有 location 指定的路由，均只会匹配请求的 uri 部分，不包括参数，这么做是因为 query string 格式是不确定的，比如：

```bash
/index.php?user=john&page=1
/index.php?page=1&user=john
```

此外，query string 可能包含任意内容

```bash
/index.php?page=1&something+else&user=john
```

以下例子展示了不同路由是如何被处理的
- 请求 `/logo.gif`，匹配路由规则 `/` 和 `~* \.(gif|jpg|png)$ `。根据上面说的匹配步骤，它会被后者处理。再根据 `root /data/www` 指令，请求会被映射到 `/data/www/logo.gif` 文件，接着该文件被发送给客户端
- 请求 `/index.php`，匹配路由规则 `/` 和 `~ \.php$`。同样的，它会被后者处理，nginx 会将这个请求转发给 `localhost:9000`，这是一个 FastCGI 服务。根据 fastcgi_param 指令，最终请求打到 `/data/www/index.php` 文件，FastCGI 会执行这个脚本。`$document_root` 变量引用 root 指令内容；`$fastcgi_script_name` 变量引用请求 URI，即 `/index.php`
- 请求 `/about.html`，只匹配路由 `/`，因此只能由它处理。根据指令 `root /data/www`，请求会映射到 `/data/www/about.html` 文件，nginx 将这个文件发送给客户端
- 请求 `/` 的处理会复杂一些。它只匹配路由 `/`。`index` 指令结合 root 指令，先判断 `/data/www/index.html` 是否存在，如存在则映射到这个文件，如果不存在，则继续查看 `/data/www/index.php`，如果存在，则做一个内部重定向到 `/index.php`，然后继续执行一次新的路由匹配，根据前面的分析，请求 `/` 最终被 FastCGI 网关处理，并执行 `/data/www/index.php` 文件