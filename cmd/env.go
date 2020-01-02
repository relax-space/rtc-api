package cmd

var env struct {
	RtcApiUrl string
	RtcDbUrl  string
}

func SetEnv(envStr string) {
	switch envStr {
	case "", "local":
		env.RtcApiUrl = "http://127.0.0.1:8080"
		env.RtcDbUrl = "http://127.0.0.1:8081"
	case "qa":
		env.RtcApiUrl = "https://qa.p2shop.com.cn/pangpang-common/rtc-api"
		env.RtcDbUrl = "https://dmz-qa.p2shop.com.cn/rtc-dmz-api/v1"
	case "production":
		env.RtcApiUrl = "https://gateway.p2shop.com.cn/pangpang-common/rtc-api"
		env.RtcDbUrl = "https://dmz-staging.p2shop.com.cn/rtc-dmz-api/v1" // for this api: staging is prd by xiao.xinmiao
		fallthrough
	default:

	}
}
