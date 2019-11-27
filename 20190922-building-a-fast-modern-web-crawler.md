# Building a fast modern web crawler

很久以来，我一直对网络爬虫充满热情。 我已经使用许多语言（例如 C++、JavaScript（Node.JS）、Python等）写了一些，我喜欢它们背后的理论。

但是首先，什么是网络爬虫？

## 什么是网络爬虫？

## What is a web crawler ?

A web crawler is a computer program that browse the internet to index existing pages, images, PDF, ... and allow user to search them using a search engine. It's basically the technology behind the famous google search engine.

Typically a efficient web crawler is designed to be distributed: instead of a single program that runs on a dedicated server, it's multiples instances of several programs that run on several servers (eg: on the cloud) that allows better task repartition, increased performances and increased bandwidth.

But distributed softwares does not come without drawbacks: there is factors that may add extra latency to your program and may decrease performances such as network latency, synchronization problems, poorly designed communication protocol, etc...

To be efficient, a distributed web crawler has to be well designed: it is important to eliminate as many bottlenecks as possible: as french admiral Olivier Lajous has said:

    The weakest link determines the strength of the whole chain.

## Trandoshan: a dark web crawler

You may know that there is several successful web crawler running on the web such as google bot. So I didn't wanted to make a new one again. What I wanted to do this time was to build a web crawler for the dark web.

## What's the dark web ?

I won't be too technical to describe what the dark web is, since it may need is own article.

The web is composed of 3 layers and we can think of it like an iceberg:

    The Surface Web, or Clear Web is the part that we browse everyday. It's indexed by popular web crawler such as Google, Qwant, Duckduckgo, etc...
    The Deep Web is a part of the web non indexed, It means that you cannot find these websites using a search engine but you'll need to access them by knowing the associated URL / IP address.
    The Dark Web is a part of the web that you cannot access using a regular browser. You'll need to use a particular application or a special proxy. The most famous dark web are the hidden services built on the tor network. They can be accessed using special URL who ends with .onion

![image-1.png](https://creekorful.me/content/images/2019/09/image-1.png)