# Development

## Install Go

推荐使用[gvm](https://github.com/moovweb/gvm)安装不同go版本.

## Setting Up

1. Clone本代码库

2. 进入项目跟目录，运行`go get`安装`go.mod`中列出的依赖

3. 添加外部文件。这些事纯文本文件，在Go编译时嵌入二进制文件中。这些文件处于安全考虑未添加到本库，需要手动放在P`build`目录下（请确保git忽略这些文件）:

    * `api.toml`: toml格式的配置文件，类似于其他框架使用的dot env文件。包含了各种用户名、密码、密钥等，通过viper在运行时加载。可以运行`make dev-env`来自自动化这一个过程，前提是在你使用的机器上的`~/config`目录下有这个文件。

    * `build_time` 文本文件，包含一个字符串，运行go build时的时间，用Linux的`date`命令生成: `date +%FT%T%z > build/build_time`

    * `commit` 文本文件，包含一个字符串，是最后一次commit的时间和hash，用这个命令生成：`git log --max-count=1 --pretty=format:%aI_%h > build/commit`

    * `version` 文本文件，包含一个字符串，是当前的最新版本，用这个命令生成：`git describe --tags > build/version`

4. 可用用Makefile中的命令自动化上述生成的文件，执行`make devenv`从本机的`~/config/api.toml`（你的机器上需要有这个文件）文件复制到`build/api.toml`。 运行 `make version`生成`build_time`, `commit`和`version`文件。见`main.go`文件中`go:embed`指令的用法。

5. 运行`make build`编译，生成的二进制在`out`目录。

6. 运行`make run`运行生成的二进制文件。

### 项目结构

`main.go`一般是Golang项目的入口，通常放在项目根目录下，一个`main.go`可以生成一个二进制文件，一个项目可以生成多个不同的二进制程序，如本项目中的`cmd`下还有三个文件夹，可以分别生成三个二进制文件。一个目录下只能有一个`main.go`。

`go.mod`和`go.sum`是Golang的外部依赖列表，类似于 npm 的`package.json`。

`Jenkinsfile`文件是持续集成使用的文件。

`Makefile`包括了编译、运行、持续集成的所有命令。Golang项目的管理是使用Makefile，类似于node的gulp、Java的gradle，使用方法参加官方手册。

`cmd`下面包含三个文件夹，分成生成三个二进制文件：

* aliwx-poller: 轮询支付宝和微信的支付状态
* iap-poller：轮询苹果订阅状态
* subs_sandbox：API的sandbox版

## 命令行参数

有两个命令行参数控制连接的数据库目标和订阅使用的价格模式。

* `-production=<true|false>` 决定使用本地开发用的数据库还是生产环境数据库，默认`true`

* `-livemode=<true|false>` 决定使用价格系统中哪种数据，对应数据库的`live_mode`字段，默认`true`

参见Makefile中 `run` 命令的使用方式，

## Upgrade Dependencies

* `go get -u` to use the latest minor and patch releases

* `go get -u=patch` to use the latest patch release.

## Avoid Cyclic Import

Golang禁止循环引用，很多编程语言中的包管理机制允许包`a`引用`b`中的变量、数据类型或函数，同时`b`引用`a`中的变量、数据类型或函数，但是Go编译器禁止这种做法，因此设计包结构式需要认真考虑，例如可以把在多个包中需要引用的定义放在一个单独的包内。

## 版本

本项目有很多branch，每个branch对应一个版本，一般有重要breaking change而又难以保持兼容时，使用一个新的版本，这要求服务器上运行一个新的binary，并且产生新的url共客户端使用。目前我在master branch上开发，完成后merge到当前版本的branch中，Jenkins持续集成时至使用最新版本对应的branch，不使用master。

## Continuous Integration

`Jenkinsfile`是Jenkis持续集成时默认使用的文件，这个文件并没有什么特殊功能，只是在调用Makefile中的命令，Jenkins会自动执行这些命令，其他Go项目相仿。

本项目的Jenkinsfile包含的命令解释如下：

1. 在`Build`步骤中，首先执行`make config`同步`api.toml`文件
2. 接下来执行`make version`生成需要embed到go二进制中的文本文件
3. 执行`make build`编译
4. 在`Deploy`步骤中，首先用`make publish`同步编译出来的二进制到运行程序的服务器
5. 执行`make restart`登录到目标服务器并重启程序

## Env File

环境变量文件`api.toml`使用toml格式， 用golang库[viper](https://github.com/spf13/viper)读取。

## Run on Server

Supervisor负责程序的运行。

## Test with Postman

When you want to test an endpoint with Postman, follow these steps:

1. Get a personal access token from Superyard. On your development machine, you can directly insert a new entry into `oauth.access` table and use this token.
2. Open Postman. Create a new collection.
3. In this collection's **Authorization** tab, select `Bearer Token` under `Type`.
4. Enter the access token you abtained in step 1 into the `Token` field.
5. Whenever you create a new HTTP request, select `Inherit from parent` in the `Type` field under the `Authorization` tab of this request.
