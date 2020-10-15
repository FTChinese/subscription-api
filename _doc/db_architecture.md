# 数据库结构

## 订阅

相关表定义的最新数据详见 https://github.com/FTChinese/sql-schema/blob/master/premium/membership.sql

### 会员信息

`premium.ftc_vip`应该视为会员当前状态的single source of truth。定义如下：

```sql
CREATE TABLE premium.ftc_vip (
    PRIMARY KEY (vip_id),
    id              VARCHAR(32),
                    UNIQUE INDEX (id),
    vip_id          VARCHAR(50)     NOT NULL,
    ftc_user_id     VARCHAR(36),
                    UNIQUE (ftc_user_id),
    wx_union_id     VARCHAR(256),
    vip_id_alias    VARCHAR(50) DEFAULT NULL,
                    UNIQUE KEY vip_id_alias (vip_id_alias) USING BTREE,
    vip_type        TINYINT(1)      NOT NULL DEFAULT '0',
    expire_time     INT(10)         NOT NULL DEFAULT '0',
                    UNIQUE (wx_union_id),
    member_tier             ENUM('standard','premium')  DEFAULT NULL,
    billing_cycle           ENUM('year','month') DEFAULT NULL,
    expire_date             DATE            DEFAULT NULL,
    payment_method          ENUM('alipay', 'wechat', 'stripe', 'apple', 'b2b'),
    ftc_plan_id             VARCHAR(32),
                            INDEX (ftc_plan_id),
    stripe_subscription_id  VARCHAR(64),
                            UNIQUE INDEX (stripe_subscription_id),
    stripe_plan_id          VARCHAR(64),
    auto_renewal            BOOLEAN DEFAULT FALSE,
    sub_status              ENUM('incomplete', 'incomplete_expired', 'trialing', 'active', 'past_due', 'canceled', 'unpaid'),
    apple_subscription_id   VARCHAR(64),
                            UNIQUE INDEX (apple_subscription_id),
    b2b_licence_id          VARCHAR(32),
                            UNIQUE INDEX (b2b_licence_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

最重要的几列是 `expire_time`, `vip_type`, `member_tier`, `expire_date`, `payment_method`。其中，`expire_time`等同于`expire_date`，`vip_type`等同于`member_tier`，只是采用的数据类型不同。支付方式`payment_method`也是必填项，只是因为保持兼容之前的数据而没有设置成 DEFAULT NOT NULL，它表明当前会员通过哪种渠道获得，这是一个ENUM类型，某一时刻只能选择其中之一。

支付方式可以分成四类，每一类又有随后的列与之关联：

* alipay/wechat，选择此支付方式则需填写 `ftc_plan_id`;
* stripe，选择此支付方式则需填写 `stripe_subscription_id`、`stripe_plan_id`、`auto_renewal`和`sub_status`;
* apple, 选择此支付方式则需填写`auto_renewal`和`apple_subscripiton_id`;
* b2b, 选择此支付方式泽穴填写`b2b_licnece_id`.

四类是互斥的，选择某种支付方式时，其关联列则为必填，其他支付方式的关联列则必须为NULL。如果需要手动更改数据库，务请注意设置支付方式极其关联列的值并清空其他支付方式的关联列。

### 会员信息快照

```sql
CREATE TABLE premium.member_snapshot (
    PRIMARY KEY (id),
    id           VARCHAR(32) NOT NULL,
    reason          ENUM('renew', 'upgrade', 'delete', 'link', 'unlink', 'apple_link', 'apple_unlink', 'b2b', 'manual', 'iap_update'),
    created_utc     DATETIME,
    created_by      VARCHAR(32),
    order_id        VARCHAR(32),
    compound_id     VARCHAR(64) NOT NULL,
                    INDEX (compound_id),
    ftc_user_id     VARCHAR(36),
                    INDEX (ftc_user_id),
    wx_union_id     VARCHAR(256),
                    INDEX (wx_union_id),
    tier            ENUM('standard', 'premium'),
    cycle           ENUM('month', 'year'),
    expire_date     DATE,
    payment_method  ENUM('alipay', 'wechat', 'stripe', 'apple', 'b2b'),
    ftc_plan_id             VARCHAR(32),
                            INDEX (ftc_plan_id),
    stripe_subscription_id  VARCHAR(64),
    stripe_plan_id          VARCHAR(64),
    auto_renewal            BOOLEAN DEFAULT FALSE,
    sub_status              ENUM('incomplete', 'incomplete_expired', 'trialing', 'active', 'past_due', 'canceled', 'unpaid'),
    apple_subscription_id   VARCHAR(64),
    b2b_licence_id          VARCHAR(32)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

会员信息在每次变动后，都会把变动前的状态备份到此表。因此，在此表中你可以查看一个人的会员变动历史。该表的数据每行一旦生成，不应更改。

## 付费墙

## 企业订阅
