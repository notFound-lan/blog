---
title: hexo 简易教程
date: 2021-06-22 22:26:33
updated: 2021-06-22 22:29:58
cover: https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/20210623003811.png
toc: 1
tags:
- hexo
categories:
- [hexo]
---

<br>

Hexo 是一个快速、简洁且高效的博客框架，特点
- 支持 markdown，提供友好的写作模式
- 支持主题，有大量开源主题可供选择
- 使用简单，一键生成静态文件，方便部署

本篇文章记录了我的整个搭建过程，供参考～

<!-- more -->

## 安装

依赖

```bash
# https://github.com/nvm-sh/nvm
➜  blog git:(master) ✗ nvm --version
0.30.2

# 设置默认 node 版本：nvm install v16.0.0 && nvm alias default v16.0.0
➜  blog git:(master) ✗ node --version
v16.0.0

➜  blog git:(master) ✗ git version
git version 2.31.0
```

hexo 安装

```bash
npm install -g hexo-cli

➜  blog git:(master) ✗ hexo version
INFO  Validating config
hexo: 5.4.0
hexo-cli: 4.2.0
```

## 创建项目

```bash
hexo init blog
hexo s --debug
```

执行以上命令后，可以在控制台看到如下日志
```bash
00:35:35.041 INFO  Hexo is running at http://localhost:4000 . Press Ctrl+C to stop.
00:35:35.046 DEBUG Database saved
00:35:36.489 DEBUG Rendering HTML index: index.html
```

打开 `localhost:4000` 可以看到博客已经成功部署并运行

## 目录结构

```bash
➜  blog git:(master) ✗ tree -I node_modules
.
├── _config.landscape.yml # 主题配置，默认主题是 landscape
├── _config.yml # hexo 配置
├── db.json
├── package.json
├── scaffolds # 模板文件，创建文章时将会从这里选择指定文件作为模板
│   ├── draft.md
│   ├── page.md
│   └── post.md
├── source # 文章地址
│   └── _posts # 当前文件夹默认作为发布文件夹，里面的文件会被渲染展示
│       └── hello-world.md # 我们打开 `localhost:4000` 展示的文件，可以是 markdown 或其他格式如 pug，只要安装了对应插件
├── themes
└── yarn.lock
```

## 创建新文章

```bash
➜  blog git:(master) ✗ hexo new test-article-1
INFO  Validating config
INFO  Created: /private/tmp/blog/blog/source/_posts/test-article-1.md

➜  blog git:(master) ✗ hexo new page --path _posts/topic1/article1 "主题1的第一篇  文章"
INFO  Validating config
INFO  Created: /private/tmp/blog/blog/source/_posts/topic1/article1.md
```

这里我们创建了两篇文章
- `_posts/test-article-1` 使用默认模版 `scaffolds/post.md`  
- `_posts/topic1/article1.md` 指定 `scaffolds/page.md` 模版，重新指定了路径


如下
```bash
➜  blog git:(master) ✗ tree source
source
└── _posts
    ├── hello-world.md
    ├── test-article-1.md
    └── topic1
        └── article1.md
```

## 静态文件

最终我们将博客部署到服务器，部署的是静态文件，即 `public` 目录

生成静态文件

```bash
hexo g

# 为了简洁，隐去了部分文件
➜  blog git:(master) ✗ tree public
public
├── 2021
│   └── 06
│       └── 22
│           ├── hello-world
│           │   └── index.html
│           └── topic1
│               └── article1
│                   └── index.html
├── archives
│   ├── 2021
│   │   ├── 06
│   │   │   └── index.html
│   │   └── index.html
│   └── index.html
├── css
│   ├── fonts
│   │   ├── FontAwesome.otf
│   ├── images
│   │   └── banner.jpg
│   └── style.css
├── fancybox
│   ├── jquery.fancybox.min.css
├── index.html
└── js
    ├── jquery-3.4.1.min.js
    └── script.js
```

我们在 public 目录起一个服务器

```bash
php -S localhost:8080
```

访问 `http://localhost:8080`，正常访问博客内容

可以发现，public 包含了网站所需的所有内容，因此只要将 public 目录推送到我们的服务器，即完成部署

## 配置文件

根目录下有两个文件
- `_config.yaml`
- `_config.landscape.yml` 

前者是 hexo 默认配置，配置项含义可以查阅 https://hexo.io/zh-cn/docs/configuration ，主要做了两件事
- 配置 cli 的行为，比如模板路径，draft 是否渲染，文件路径等等
- 配置生成的静态文件内容，比如网站的 布局、title、url、分页行为等

后者是主题配置，比如一些主题支持社交组件（如 微信、知乎、github 等），可以在这个配置文件进行自定义配置

## 实践

### 文章添加 tag 和 category

```bash
---
xxx:
tags:
- hexo
categories:
- [hexo]
---
```

### 图片管理

纯编辑器在 markdown 插入图片一直是一件很繁琐的事，这里我决定采用 阿里云OSS+PicGo 的方式，主要解决
- 图片插入困难问题，比如截图需要保存文件，再通过 tag 引用
- 图片需要手动管理，难以维护

oss 购买及图床使用请参考 {% link 阿里云OSS PicGo 配置图床教程 超详细 https://zhuanlan.zhihu.com/p/104152479 external %}

然后就可以简单的上传并插入图片，比如

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/cat-4638664_960_720.jpg" />

ps：发现一个免费图片素材库，且图片非常小 https://pixabay.com/zh/

关于 picGo，添加插件 https://github.com/gclove/picgo-plugin-super-prefix， 并关闭上传重命名，可以获得较规范的路径，`/img/2021/06/23/20210623082611.png`

## 使用模板

我选择的是 https://docs.nexmoe.com/ 文档写的很详细了，跟着走就行，下面列出一些文档没有体现的点

### 添加目录

见 https://github.com/theme-nexmoe/hexo-theme-nexmoe/issues/73

只要在 blog header 添加 `toc: 1`，如下

```bash
---
title: xxx
date: xxx
toc: 1
```

## 部署

如果你是自建服务器，比如放在一个 nginx 后面，参考 https://hexo.io/zh-cn/docs/one-command-deployment

我采用 github page 的方式，见 https://hexo.io/zh-cn/docs/github-pages

提交 master 后，会触发 ci-cd，生成 public 并提交到 gh-pages 分支，再由 github 完成静态文件更新

这时候就可以通过 `xxx.github.io` 的二级域名访问了

### 自定义域名和 https 支持

首先 dig 一下获得 ip `dig xxx.github.io` 

```bash
[root@xxx ~]# dig notfound-lan.github.io

; <<>> DiG 9.9.4-RedHat-9.9.4-74.el7_6.2 <<>> notfound-lan.github.io
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 7643
;; flags: qr rd ra; QUERY: 1, ANSWER: 4, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
;; QUESTION SECTION:
;notfound-lan.github.io.		IN	A

;; ANSWER SECTION:
notfound-lan.github.io.	3600	IN	A	185.199.110.153
notfound-lan.github.io.	3600	IN	A	185.199.111.153
notfound-lan.github.io.	3600	IN	A	185.199.109.153
notfound-lan.github.io.	3600	IN	A	185.199.108.153
```

然后配置四条 A 记录指向上面四个 ip，再建一条 CCA 记录，参考 https://help.aliyun.com/document_detail/65537.html

接着在 github 上用自定义域名替代默认 xxx.github.io 域名


## 总结

<img src="https://voiddme-blog-public.oss-cn-beijing.aliyuncs.com/img/2021/06/23/20210623181937.png" />