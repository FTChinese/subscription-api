# Introduction

## One-time-purchase

ftc_vip如下字段组合互斥：
payment_method == (alipay || wechat) 和 ftc_plan_id 同时存在 并且 auto_renewal为0
payment_method == apple  和 apple_subscription、auto_renewal 同时存在
payment_method == stripe 和 stripe_subscription_id、stripe_plan_id、sub_status、auto_renewal 同时存在
