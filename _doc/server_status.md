# Server Internal Status

## Cache Latest Promotion Schedule

    GET /__refresh

Tell this API to retrieve a promotion schedule and put it into cache. It will only retrieve one row from database whose creation time is the latest. You have no way to tell it which only to pick.

The response is the same as `/paywall/promo`

## Build Version

    GET /__version

The running program's build verison, commit tag and when it is built.

```json
{
    "build": "2018-12-02T11:13:33+0800",
    "version": "v0.1.0-24-g02ea926"
}
```