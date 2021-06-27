---
title: nginx location 路由匹配规则
cover: https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210627210644.jpeg
toc: 1
categories:
  - [Nginx]
date: 2021-06-27 20:54:50
tags:
- nginx
---

<br>

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210627212110.png" />

如上图所示，请求匹配到虚拟主机后，继续由 `location` 指令进行路由。location 定义了路由模式，以及对请求的具体处理方法

-----

<!-- more -->

## location 指令

`location` 可以定义在 `server` 和 `location` 两个 context 中，指令语法

```nginx
location [ modifier ] uri { ... } # 这里我们只关心这种
location @name { ... }
```

从类别上，location 可以分为两大类
- 前缀匹配
- 正则表达式

modifier 有以下几项
- 空，前缀匹配。放在第一步查找
- `=` 完全匹配。放在第一步查找，如果匹配到，则不走第二步
- `^~` 前缀匹配，放在第一步查找，如果匹配到，则不走第二步
- `~` 大小写敏感的正则匹配
- `~*` 大小写不敏感的正则匹配

## 路由匹配方式

- 先找出契合度最高的前缀匹配路由：顺序无关 + 最长前缀匹配。如果该路由 modifier 是 `^~` 或 `=` 则直接用这个 location，否则暂存并继续第二步
- 按配置顺序查找，找到第一个匹配的正则路由。如果没有匹配的正则路由，则使用第一步找到的前缀路由

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/20210627232151.png" />

## 案例

```nginx
server {
    listen 4500;

    # 1
    location = / { 
        return 201;
    }

    # 2
    location / {
        return 202;
    }

    # 3
    location /documents/ {
        return 203;
    }

    # 4
    # 因为 5 的存在，所以 /images/xxx 永远匹配到第五项
    location ~* ^/image./ {
        return 205;
    }

    # 5
    location ^~ /images/ {
        return 204;
    }

    # 6
    location ~* \.(gif|jpg|jpeg)$ {
        return 206;
    }
}
```

- `/` 匹配到配置 1。前缀匹配
- `/index.html` 匹配到配置 2。前缀匹配 -> 正则匹配 -> 前缀匹配项 
- `/documents/document.html` 匹配配置 3。前缀匹配 -> 正则匹配 -> 前缀匹配项
- `/images/1.gif` 匹配到配置 5。前缀匹配
- `/documents/1.jpg` 匹配到配置 6。前缀匹配 -> 正则匹配


## 参考

- [ngx_http_core_module#location](http://nginx.org/en/docs/http/ngx_http_core_module.html#location)