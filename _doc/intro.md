# Introduction

本API目前分成三个部分：

* 订阅支付
* 付费墙信息展示
* 微信登录（尚未实现）

## 订阅支付

处理订阅的流程中最重要的数据类型是`Subscription`，包含了订单的ID、购买的会员类型、会员周期、支付价格、支付方式、订单创建时间、订单确认时间、订单是否用于续订、订单购买的会员的开始时间和到期时间，以及该用户的UUID。参见`model/subscripiton.go`文件。

一个会员购买成功分成两个步骤，第一是客户端请求创建订单，这一步的流程是

1. 按照用户选择的支付方式，向`POST /wxpay/unified-order/{standard|premium}/{year|month}`或`POST /alipay/app-order/{tier}/{cycle}`发起创建订单的请求，请求本身不需要有body。但是请求的头部需要包含一些元数据，如`X-User-Id`必须有，值是用户的ID，否则API会拒绝请求。还要包含一些其他数据，如客户端的类型、版本号等，见下文细节。

2. API根据请求的URL，找出来当前使用的定价方案是什么。有三套方案：`standard_year`, `standard_month`, `premium_year`三种，不在这三种方案中的请求会被拒绝。然后用这三个方案作为键，去寻找对应的价格方案：如果当前有促销方案，则使用促销价，否则使用默认价格。

3. 根据选定的方案，创建`Subscrption`实例。此时的Subscription包含这些字段：
    * OrderID
    * TierToBuy
    * BillingCycle
    * Price
    * TotalAmount
    * PaymentMethod
    * UserID

4. 从请求头部取出User ID，用User ID去查找这是用户是否已经是会员了。如果没有找到，则认为这是一个新订阅；如果找到了，则说明这个用户已经是或者曾经是订阅用户。对于目前的订阅用户，我们要限制他无限制交费，即指定一个期限，在该期限内允许交费，否则告诉用户已经是会员了，不需要重复购买。这个限制策略是：
   
**当前时间到会员结束时间的差小于当前请求订阅的周期的时长**

公式：
```
Expiration date - now < billing cycle
```

举例：

一个用户在2018-12-04日购买了一个月的会员，接下来马上就可以再买一个月，但是马上进行第三次再购买一个月会被拒绝，因为**今天**到会员的**截止日期**（两个月后）已经超过了**一个订阅周期**（一个月）。但是第三次可以再买一年的版本，因为**今天**到会员的**截止日期**（两个月后）小于所请求的**订阅周期**（一年）。年度订阅同理：一个年订阅用户要续订一个月的版本，则只能等到结束前一个月续订；距离截止日期不到一年了， 则可以买下一年的，但是买第三年则被拒绝。

5. API把订单保存到数据库。
6. 按照各支付供应商的要求签名订单或者请求生成与支付订单，把生成的订单数据返回给客户端。数据格式见后。注意：按照微信和支付宝建议，数据签名在服务器端完成，因此，价格也是在服务器端定下来的，所以从API获取到的订单中包含了价格信息，客户端是无权指定价格的。
7. 客户端拿到数据后调用支付SDK。

第二个步骤是API接受支付供应商的服务器端通知，流程如下：

1. 支付供应商向本API发起POST请求。
   
2. 从请求的body中取出我方生成的订单ID、订单完成时间，去数据库中查询该订单。
   
3. 找到订单后，首先检查订单是否已经确认过，如果已经确认过，则告诉支付方不要再继续发通知。
   
4. 如果订单尚未确认，则开始确认订单：首先更新`Subscription`中的确认时间、取出确认时间的年月日部分作为订单购买的会员开始时间、把开始时间加上订阅周期作为结束时间；
   
5. 如果该订单此前记录了是用于续订的，则找出该会员的会员到期日期，把到期日期作为该笔订单的开始时间，并根据订阅周期推测出订单购买的到期时间。
   
6. 订单信息计算完成，此时尚未向数据库更新任何信息。
7. 我们使用SQL的Transaction来同时更新订单的确认时间、开始日期、截止日期以及会员的类型、周期和截止日期，让记录订单的表和记录会员信息的表同时更新或者同时失败，保证数据的完整性。这里会员信息在保存时，如果不存在，则创建（新会员）；如果已经存在了，则更新其截止日期。

至此，全部购买流程结束。

## 付费墙信息展示

见Paywall部分。

# Endpoints

Here's an overview of all the endpoint provided by this API.

## Internal Status
* `/__version` See current program's build info
* `/__refresh` Notify server to retrieve a promotion schedule
* `/__current_plans` See what procing plans are being used.

## Subscription Order
* `POST /wxpay/unified-order/{standard|premium}/{year|month}` Create a wxpay prepay order
* `GET /wxpay/query/{orderId}` Query an order paid via wxpay

* `POST /alipay/app-order/{tier}/{cycle}` Create a new order for alipay

## Server to Server Notification
* `POST /callback/wxpay`
* `POST /callback/alipay`

## Paywall
* `GET /paywall/promo` Get the promotion schedule
* `GET /paywall/products` Get the products description
* `GET /paywall/plans` Get the default pricing plans.
* `GET /paywall/banner` Get the banner content used on subscription page.

## Wechat Login

* `POST /auth/wx` Login via wechat
* `POST /auth/email` 
* `PUT /wx/bind` Bind an FTC account to wechat.
* `DELETE /wx/bind` Unbind an FTC account from wechat
* `GET /wx/account` Get the the complete account: FTC + wechat + membership.