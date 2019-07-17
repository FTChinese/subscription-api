package paywall

// The default banner message used on web version of pay wall.
var defaultBanner = Banner{
	Heading:    "FT中文网会员订阅服务",
	CoverURL:   "http://www.ftacademy.cn/subscription.jpg",
	SubHeading: "欢迎您",
	Content: []string{
		"希望全球视野的FT中文网，能够带您站在高海拔的地方俯瞰世界，引发您的思考，从不同的角度看到不一样的事物，见他人之未见！",
	},
}

func GetDefaultBanner() Banner {
	return defaultBanner
}
