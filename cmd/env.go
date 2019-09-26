package cmd

var env struct {
	RtcApiUrl string
}

func SetEnv(envStr string) {
	switch envStr {
	case "", "staging":
		env.RtcApiUrl = "http://127.0.0.1:8080"
	case "qa":
		env.RtcApiUrl = "https://qa.p2shop.com.cn/pangpang-common/rtc-api"
	case "production":
		env.RtcApiUrl = "https://gateway.p2shop.com.cn/pangpang-common/rtc-api"
		fallthrough
	default:

	}
}
