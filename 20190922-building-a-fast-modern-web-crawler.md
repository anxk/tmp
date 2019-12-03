# 构建一个快速的现代化网络爬虫

很久以来，我一直对网络爬虫充满热情，着迷于它们背后的理论。我曾经使用过许多语言来构建它，例如：C++、JavaScript（Node.JS）、Python 等。

但是首先，什么是网络爬虫？

## 什么是网络爬虫？

网络爬虫是一个计算机程序，它通过浏览互联网将现有的网页、图像、PDF等编入索引，并允许用户使用[搜索引擎](https://en.wikipedia.org/wiki/Web_search_engine)来检索这些内容。 这基本上就是著名的[谷歌搜索引擎](https://google.com/)背后的技术了。

通常，一个高效的网络爬虫被设计成分布式的，即不是一个运行在专用服务器上的独立程序，而是在多个服务器上（例如：在云上）运行一些程序的多个实例。这样使得任务能够得到更合适的重新拆分，从而达到提高性能、增加带宽的效果。

但是，分布式软件并非没有缺点，有一些因素可能会给程序增加额外的延迟，从而引起性能上的降低。这些因素例如：网络延迟、同步问题、设计不良的通讯协议等。

为了提高效率，分布式网络爬虫必须得到精心的设计，尽可能的消除瓶颈，正如法国海军上将 Olivier Lajous 所说：
> 最薄弱的链接决定了整个链条的强度。

## Trandoshan: 一个暗网爬虫

您也许知道，已经有一些比较成功的爬虫正在网络上运行，例如 google bot，所以这次我不打算再做一个类似的，而是要做一个专门用于暗网的网络爬虫。

## 什么是暗网？

在这里我将不会从技术的角度来阐述什么是暗网，因为有专门的文章来描述它。

互联网由三层组成，我们可以把它想象成一个冰山：

* 表层网或者透明网是我们每天浏览的网络部分，它由 Google，Qwant，Duckduckgo 等流行的网络爬虫索引。
* 深网是未经索引的网络的一部分，这意味着您无法使用搜索引擎找到这些网站，但是您可以通过 URL 或者 IP 地址来访问它们。
* 暗网是网络的另一部分，您无法使用常规浏览器去访问，而是需要使用特殊的应用程序或特代理才能访问。 最著名的暗网是建立在 Tor 网络上，它可以使用以 .onion 结尾的特殊 URL 来访问它。

![Existing web layers](https://creekorful.me/content/images/2019/09/image-1.png)


## Trandoshan 是怎样被设计的？

![Big picture of Trandoshan](https://creekorful.me/content/images/2019/09/Trandoshan-1.png)

在谈论每个进程的职责之前，了解它们之间如何通信是很重要的。

进程间通信（IPC）主要是使用 [NATS](https://nats.io/) 的消息传递协议（上图中的黄线）基于生产者 / 消费者模型来实现的。 NATS 中的每个消息都有一个主题（类似于电子邮件），该主题允许其他进程对消息进行识别，以便只能读取它们感兴趣的消息。NATS 允许扩展：例如，可以有 10 个爬虫进程从消息服务器中同时读取 URL。 每个进程都将获取唯一的 URL 进行爬取。 这允许进程并发（许多实例可以同时运行而没有任何错误），因此可以提高性能。

Trandoshan 被拆分为 4 个主要进程：

* **爬虫**：爬虫进程负责爬取页面，它从 NATS 中读取 URL（由主题 **"todoUrls"** 标识的消息），然后对页面进行爬取，并提取页面中含有的所有 URL。这些提取到的 URL 将被以 **"crawledUrls"** 为主题发给 NATS，而对应的页面正文（整个内容）则会被以 **"content"** 为主题发往 NATS。
* **调度器**：调度器负责对 URL 进行审批，它读取主题为 **"crawledUrls"** 的消息，然后决定是否要爬取对应的 URL（如果是没有爬取过的 URL），如果需要爬取，则将 URL 以 **"todoUrls"** 为主题发给 NATS。
* **持久器**：持久器负责将内容归档，它读取页面内容（由主题 **"content"** 标识的消息）并将其存储到 NoSQL 数据库（MongoDB）中。
* **API**：API负责收集信息供其它进程使用。例如，**调度器**通过它来确定某个页面是否已被爬取过。调度程序并不直接调用数据库接口来检查 URL 是否存在（这将增加与数据库技术之间的耦合），而是使用API，这使得数据库和进程之间有了某种抽象。

不同的进程是使用 Go 语言编写的，因为它提供了很多性能（因为它是作为二进制文件编译的），并且具有很多库的支持。 Go 语言经过精心设计，可以构建高性能的分布式系统。

Trandoshan 的源代码可在以下 github 上找到：[https://github.com/trandoshan-io](https://github.com/trandoshan-io)。

## 怎样运行 Trandoshan？

如前所述，Trandoshan 被设计为运行在分布式系统之上，并且可以通过 docker 镜像来获取，这使得它成为云环境下的理想选择。 实际上，有一个仓库包含了在 Kubernetes 集群上部署 Trandoshan 生产实例所需的所有配置文件，这些文件位于 [https://github.com/trandoshan-io/k8s](https://github.com/trandoshan-io/k8s)，对应的容器镜像位于 [docker hub](https://hub.docker.com/u/trandoshanio) 上。

如果您正确配置了 kubectl，就可以可以通过下面简单的命令来部署 Trandoshan：

```bash
./bootstrap.sh
```

除此之外，您可以使用 docker 和 docker-compose 在本地运行 Trandoshan。 在 [trandoshan-parent](https://github.com/trandoshan-io/trandoshan-parent) 仓库中，借助一个 docker compose 文件和一个 shell 脚本，使用以下命令便可运行此应用程序：

```bash
./deploy.sh
```

## 怎样使用 Trandoshan？

目前，有一个精巧的 Angular 应用程序可以用于检索被编入索引的内容，这个程序使用 **API** 进程来完成对数据库中内容的检索。

![Screenshot of the dashboard](https://creekorful.me/content/images/2019/09/Screenshot-from-2019-09-22-17-09-49.png)

## 总结

目前仅此而已。尽管 Trandoshan 已经可以被用于生产环境，但是仍有许多优化工作需要和新功能需要完成。由于它是一个开源项目，所以每个人都可以通过对相应的项目发起 PR 来做出贡献。

Happy hacking!

---

via: https://creekorful.me/building-fast-modern-web-crawler/

作者：[Aloïs Micard](https://creekorful.me/author/creekorful/)
译者：[Anxk](https://github.com/Anxk)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出