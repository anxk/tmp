# 基于 Jenkins 多分支流水线 CI/CD 的实现

## 创建具有多个分支的 CI/CD 流水线的指南。

![Making multiple branches with Jenkins.](https://dzone.com/storage/temp/12819412-branch.jpg)

## 简介

Jenkins 是一个持续集成服务器，可以从版本控制系统（VCS）中获取最新代码，对其进行构建、测试并通知开发人员。除了成为持续集成（CI）服务器之外，Jenkins 还可以做很多事情。 Jenkins 最初称为 Hudson，是川口浩辅（Kohsuke Kawaguchi）编写的一个开源项目。由于 Jenkins 是基于 Java 的项目，因此，在计算机上安装和运行 Jenkins 之前，首先需要安装Java 8。

多分支流水线允许您在 Jenkinsfile 的帮助下为源代码管理（SCM）存储库中的每个分支自动创建流水线。

## 什么是 Jenkinsfile

可以使用名为 Jenkinsfile 的文本文件来定义 Jenkins 流水线。您可以使用 Jenkinsfile 将管道实现为代码，并且可以使用域特定语言（DSL）来定义。使用 Jenkinsfile，您可以编写运行 Jenkins 流水线所需的步骤。

## 来自 Jenkins 的定义

多分支流水线项目类型使您可以为同一项目的不同分支实现不同的 Jenkinsfile。在 Multibranch Pipeline 项目中，Jenkins 自动发现，管理和执行包含源控件的分支的流水线。

![Architecture Diagram](https://dzone.com/storage/temp/12713975-multibranch-pipeline.png)

## 创建一个简单多分支流水线任务的步骤

1. 点击 Jenkins 主面板左上角的 **New Item** 选项：
![New Item](https://dzone.com/storage/temp/12713939-new-item.png)
2. 在 **Enter an item name** 中填入任务名，向下滚动，然后选择 **Multibranch Pipeline**，最后点击 **OK** 按钮：
![Multibranch pipeline](https://dzone.com/storage/temp/12713992-select-multibranch.png)
3. 填写**任务描述**（可选）
4. 添加一个 **Branch Source**（例如：GitHub）并且填写代码仓库的位置
5. 选择 **Add** 按钮添加凭证并点击 **Jenkins**
6. 键入 GitHub **用户名**、**密码**、**ID** 和描述
7. 从下拉菜单中选择凭证
![Branch sources](https://dzone.com/storage/temp/12714013-select-repo.png)
8. 点击 **Save** 保存该多分支流水线任务
9. Jenkins 自动扫描指定的仓库并为组织文件夹做一些索引。 Organization Folders 意味着使 Jenkins 能够监视整个 GitHub Organization 或 Bitbucket Team/Project，并自动为包含分支的仓库创建新的多分支流水线和包含 Jenkinsfile 的拉取请求。
![Scan repository log](https://dzone.com/storage/temp/12714002-scanning.png)
10. 当前，此功能仅适用于 GitHub 和 Bitbucket，其功能由 [GitHub Organization Folder](https://plugins.jenkins.io/github-organization-folder) 和 [Bitbucket Branch Source](https://plugins.jenkins.io/cloudbees-bitbucket-branch-source) 插件提供。
11. 一旦任务被创建，构建将会被自动触发
![Builds triggered](https://dzone.com/storage/temp/12714003-jobs.png)

## 配置 Webhooks

12. 我们必须配置我们的 Jenkins 机器来与我们的 GitHub 存储库通信。为此，我们需要获取 Jenkins 机器的 Hook URL。
13. 转到管理 Jenkins，然后选择配置系统视图。
14. 找到 GitHub Plugin Configuration 部分，然后单击 Advanced 按钮。
15. 选择“为GitHub配置指定其他挂钩URL”
![Webhooks](https://dzone.com/storage/temp/12713987-n-ci-sepcify-hook.png)
16. 在文本框字段中复制URL，然后取消选择它。
17. 单击Save，它将重定向到Jenkins仪表板。
18. 导航到浏览器上的GitHub选项卡，然后选择您的GitHub存储库。
19. 单击设置。它将导航到存储库设置。
![Settings](https://dzone.com/storage/temp/12713983-settings.png)
20. 单击Webhooks部分。
21. 单击添加Webhook按钮。将挂钩URL粘贴在有效载荷URL字段上。
22. 确保触发器Webhook字段已选中“仅推送事件”选项。
![Add webhook](https://dzone.com/storage/temp/12713985-add-webhook.png)
23. 单击Add webhook，它将把webhook添加到您的存储库。
24. 正确添加Webhook后，您会看到带有绿色勾号的Webhook。
![Added webhook](https://dzone.com/storage/temp/12713984-green-tick.png)
25. 返回到存储库，然后更改为“分支”并更新任何文件。在这种情况下，我们将更新README.md文件。
26. 现在看到Jenkins作业将自动触发。
![CI triggers](https://dzone.com/storage/temp/12714004-cicd.png)
27. 管道执行完成后，我们可以通过单击内部版本号来验证“内部版本历史”下已执行内部版本的历史。
28. 单击内部版本号，然后选择Console Output。在这里，您可以看到每个步骤的输出。
![Console Output](https://dzone.com/storage/temp/12714005-console-output.png)

## 进一步阅读

[Learn How to Set Up a CI/CD Pipeline From Scratch](https://dzone.com/articles/learn-how-to-setup-a-cicd-pipeline-from-scratch)
[API Builder: A Simple CI/CD Implementation – Part 1](https://dzone.com/articles/api-builder-a-simple-cicd-implementation-part-1)
