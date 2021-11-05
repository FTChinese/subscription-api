# Wechat OAuth Details

网页和移动app获取code的过程不同。网页遵循OAuth规定，通过跳转和回调获取，移动app则是通过调用SDK。

二者的文档分别散落在不同地方：

* [网页版](https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html)
* [移动app版](https://developers.weixin.qq.com/doc/oplatform/Mobile_App/WeChat_Login/Development_Guide.html)

注意事项：

* 网页的回调地址必须属于最初在微信注册的域名，子域名不能用。

## Step 1: Get the Code

### For web

第三方使用网站应用授权登录前请注意已获取相应网页授权作用域（scope=snsapi_login），可以通过在PC端打开以下链接

```
https://open.weixin.qq.com/connect/qrconnect?appid=<APPID>&redirect_uri=<REDIRECT_URI>&response_type=code&scope=<SCOPE>&state=<STATE>#wechat_redirect
```

* `appid: string`. Required. Your app's unique identifier.
* `redirect_uri: string`. Required. Url-encoded callback url on your site.
* `response_type: 'code'`
* `scope: 'snsapi_login'`
* `state: string` An optional random. Recommend. Returned as is.

跳转到上述网址后，微信会显示一个二维码，用户使用微信扫码，情况分两种：

* 用户允许授权后，将会重定向到上述`redirect_uri`的网址上，并且带上`code`和`state`参数：

```
<redirect_uri>?code=<CODE>&state=<STATE>
```

* 若用户禁止授权，则重定向后不会带上code参数，仅会带上state参数

```
<redirect_uri>?state=<STATE>
```

### For mobile app

## Step 2: Exchange code for access token

## Step 3: Use access token to get user information

