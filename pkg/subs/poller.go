package subs

const StmtAliUnconfirmed = colOrder + `
FROM premium.log_ali_notification AS a
    LEFT JOIN premium.ftc_trade AS o
    ON a.ftc_order_id = o.trade_no
WHERE o.confirmed_utc IS NULL
    AND a.trade_status = 'TRADE_SUCCESS'`

const StmtWxUnconfirmed = colOrder + `
FROM premium.log_wx_notification AS w
    LEFT JOIN premium.ftc_trade AS o
    ON w.ftc_order_id = o.trade_no
WHERE o.confirmed_utc IS NULL
    AND w.result_code = 'SUCCESS'`
