# 构建一个快速的现代网络爬虫

很久以来，我一直对网络爬虫充满热情。 我已经使用许多语言（例如 C++、JavaScript（Node.JS）、Python等）写了一些，我喜欢它们背后的理论。

但是首先，什么是网络爬虫？

## 什么是网络爬虫？

网络爬虫是一个计算机程序，它通过浏览互联网来索引现有的网页、图像、PDF等，并允许用户使用搜索引擎来搜索它们。 这基本上就是著名的谷歌搜搜索擎背后的技术。

通常，高效的网络爬虫的设计是分布式的：不是一个单一的程序运行在在专用服务器上，而是在多个服务器上运行多个实例（例如：在云上），允许更好的任务再分配，提高性能演出和增加带宽。

分布式软件并非没有缺点：有些因素可能会增加程序的延迟，并可能降低性能，例如网络延迟，同步问题，设计不良的通讯协议等。

为了提高效率，必须对分布式Web爬虫进行精心设计：消除尽可能多的瓶颈非常重要：正如法国海军上将Olivier Lajous所说的：最薄弱的环节决定了整个链条的强度。

## Trandoshan: 一个暗网爬虫

您可能知道，有一些成功的网络搜寻器正在网络上运行，例如google bot。 所以我不想再做一个新的。 我这次想做的是为黑暗的网络构建一个网络爬虫。

## 什么是暗网？

我不太会描述暗网是什么，因为它可能需要的是自己的文章。

互联网由三层组成，我们可以把它想象成一个冰山：

* Surface Web或Clear Web是我们每天浏览的部分。 它由Google，Qwant，Duckduckgo等流行的网络爬虫索引。
* Deep Web是未经索引的Web的一部分，这意味着您无法使用搜索引擎找到这些网站，但是您需要通过了解关联的URL / IP地址来访问它们。
* Dark Web是Web的一部分，您无法使用常规浏览器访问。 您需要使用特定的应用程序或特殊的代理。 最著名的暗网是建立在Tor网络上的隐藏服务。 可以使用以.onion结尾的特殊URL访问它们。

![image-1.png](https://creekorful.me/content/images/2019/09/image-1.png)

## Trandoshan 是怎样设计的？

![Trandoshan-1.png](https://creekorful.me/content/images/2019/09/Trandoshan-1.png)

在谈论每个过程的责任之前，重要的是要了解它们如何彼此交谈。

进程间通信（IPC）主要基于生产者/消费者模式使用称为NATS的消息传递协议（图中的黄线）完成。 NATS中的每个消息都有一个主题（例如电子邮件），该主题允许其他进程对其进行识别，因此只能读取他们想要阅读的消息。 允许扩展的NATS：例如，它们可以是10个搜寻器进程，它们从消息传递服务器读取URL。 这些过程中的每一个都会收到一个唯一的URL进行爬网。 这允许进程并发（许多实例可以同时运行而没有任何错误），因此可以提高性能。

Trandoshan 被拆分为4个主要进程：

* 爬网程序：负责爬网的过程：它读取URL以从NATS进行爬网（由主题“ todoUrls”标识的消息），对页面进行爬网，并提取页面中存在的所有URL。 将这些提取的URL发送给主题为“ crawledUrls”的NATS，并将页面正文（整个内容）发送给主题为“ content”的NATS。
* Scheduler：负责URL批准的过程：该过程读取“ crawledUrls”消息，检查是否要爬网URL（如果尚未爬网URL），如果是，则将URL发送给主题为“ todoUrls”的NATS。
* Persister：负责内容归档的过程：它读取页面内容（由主题“ content”标识的消息）并将其存储到NoSQL数据库（MongoDB）中。
* API：其他进程用于收集信息的进程。 例如，调度程序使用它来确定页面是否已被爬网。 调度程序不使用直接调用数据库来检查URL是否存在（这将增加与数据库技术的额外耦合），而是使用API：这允许在数据库/进程之间进行某种抽象。

不同的过程是使用Go编写的：因为它提供了很多性能（因为它是作为本机二进制文件编译的），并且具有很多库支持。 Go经过精心设计，可以构建高性能的分布式系统。

Trandoshan的源代码可在以下github上找到：https：//github.com/trandoshan-io。

## 怎样运行 Trandoshan？

如前所述，Trandoshan设计为可在分布式系统上运行，并且可作为docker映像使用，这使其成为云计算的理想选择。 实际上，存在一个存储库，其中包含在Kubernetes集群上部署Trandoshan的生产实例所需的所有配置文件。 这些文件位于此处：https://github.com/trandoshan-io/k8s，容器映像位于docker hub上。

如果您正确配置了kubectl，则可以通过简单的命令部署Trandoshan：

```bash
./bootstrap.sh
```

否则，您可以使用docker和docker-compose在本地运行Trandoshan。 在trandoshan-parent储存库中，有一个撰写文件和一个Shell脚本，可使用以下命令运行应用程序：

```bash
./deploy.sh
```

## 怎样使用 Trandoshan？

目前，有一个小的Angular应用程序可以搜索索引内容。 该页面使用API流程对数据库执行搜索。

![Screenshot-from-2019-09-22-17-09-49.png](https://creekorful.me/content/images/2019/09/Screenshot-from-2019-09-22-17-09-49.png)

## 总结

目前仅此而已。 Trandoshan已准备好进行生产，但是还有很多优化工作需要完成，功能也需要合并。 由于它是一个开源项目，因此每个人都可以通过对相应项目进行拉取请求来对此做出贡献。

尽情折腾吧！

---

via: https://creekorful.me/building-fast-modern-web-crawler/

作者：[Aloïs Micard](https://creekorful.me/author/creekorful/)
译者：[Anxk](https://github.com/Anxk)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出