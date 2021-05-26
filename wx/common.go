package wx

const (
	Fail    = "FAIL"
	Success = "SUCCESS"

	Sign               = "sign"
	SignTypeHmacSha256 = "HMAC-SHA256"
	SignTypeMd5        = "MD5"

	MIMEApplicationXML = "application/xml; charset=utf-8"

	MicroPayUrl         = "https://api.mch.weixin.qq.com/pay/micropay"
	UnifiedOrderUrl     = "https://api.mch.weixin.qq.com/pay/unifiedorder"
	OrderQueryUrl       = "https://api.mch.weixin.qq.com/pay/orderquery"
	ReverseUrl          = "https://api.mch.weixin.qq.com/secapi/pay/reverse"
	CloseOrderUrl       = "https://api.mch.weixin.qq.com/pay/closeorder"
	RefundUrl           = "https://api.mch.weixin.qq.com/secapi/pay/refund"
	RefundQueryUrl      = "https://api.mch.weixin.qq.com/pay/refundquery"
	DownloadBillUrl     = "https://api.mch.weixin.qq.com/pay/downloadbill"
	ReportUrl           = "https://api.mch.weixin.qq.com/payitil/report"
	ShortUrl            = "https://api.mch.weixin.qq.com/tools/shorturl"
	AuthCodeToOpenidUrl = "https://api.mch.weixin.qq.com/tools/authcodetoopenid"

	SandboxMicroPayUrl         = "https://api.mch.weixin.qq.com/sandboxnew/pay/micropay"
	SandboxUnifiedOrderUrl     = "https://api.mch.weixin.qq.com/sandboxnew/pay/unifiedorder"
	SandboxOrderQueryUrl       = "https://api.mch.weixin.qq.com/sandboxnew/pay/orderquery"
	SandboxReverseUrl          = "https://api.mch.weixin.qq.com/sandboxnew/secapi/pay/reverse"
	SandboxCloseOrderUrl       = "https://api.mch.weixin.qq.com/sandboxnew/pay/closeorder"
	SandboxRefundUrl           = "https://api.mch.weixin.qq.com/sandboxnew/secapi/pay/refund"
	SandboxRefundQueryUrl      = "https://api.mch.weixin.qq.com/sandboxnew/pay/refundquery"
	SandboxDownloadBillUrl     = "https://api.mch.weixin.qq.com/sandboxnew/pay/downloadbill"
	SandboxReportUrl           = "https://api.mch.weixin.qq.com/sandboxnew/payitil/report"
	SandboxShortUrl            = "https://api.mch.weixin.qq.com/sandboxnew/tools/shorturl"
	SandboxAuthCodeToOpenidUrl = "https://api.mch.weixin.qq.com/sandboxnew/tools/authcodetoopenid"
)

//=================================================================
//							[Response]通用参数
//=================================================================
// ResponseReturnCode 基本信息
type ResponseReturnCode struct {
	ReturnCode string `xml:"return_code"` // 返回状态码 SUCCESS/FAIL
	ReturnMsg  string `xml:"return_msg"`  // 返回信息
}

// ResponseResultCode 返回通用数据
type ResponseResultCode struct {
	AppID      string `xml:"appid,emitempty"`
	MchID      string `xml:"mch_id,emitempty"`
	DeviceInfo string `xml:"device_info,emitempty"`
	NonceStr   string `xml:"nonce_str,emitempty"`
	Sign       string `xml:"sign,emitempty"`
	ResultCode string `xml:"result_code,emitempty"`
	ErrCode    string `xml:"err_code,emitempty"`
	ErrCodeDes string `xml:"err_code_des,emitempty"`
}

// ResponseResultCodeSuccess 结果通用数据
type ResponseResultCodeSuccess struct {
	OpenID         string `xml:"openid,emitempty"`           // 用户在商户appid下的唯一标识
	IsSubscribe    string `xml:"is_subscribe,emitempty"`     // 是否关注公众号  Y/N
	TradeType      string `xml:"trade_type,emitempty"`       // 调用接口提交的交易类型 APP
	BankType       string `xml:"bank_type,emitempty"`        // 付款银行,采用字符串类型的银行标识
	FeeType        string `xml:"fee_type,emitempty"`         // 货币种类 CNY
	TotalFee       int64  `xml:"total_fee,emitempty"`        // 总金额 100
	CashFeeType    string `xml:"cash_fee_type,emitempty"`    // 现金支付货币类型 CNY
	CashFee        int64  `xml:"cash_fee,emitempty"`         // 现金支付金额 100
	TransactionID  string `xml:"transaction_id,emitempty"`   // 微信支付订单号
	OutTradeNO     string `xml:"out_trade_no,emitempty"`     // 商户订单号
	Attach         string `xml:"attach,emitempty"`           // 附加数据
	TimeEnd        string `xml:"time_end,emitempty"`         // 支付完成时间
	TradeStateDesc string `xml:"trade_state_desc,emitempty"` // 交易状态描述
}

type ResponseBaseCode struct {
	ResponseReturnCode
	ResponseResultCode
}

//=================================================================
//							[Response]统一下单
//=================================================================
type WxUnifiedOrderResponse struct {
	ResponseBaseCode

	// 以下字段在return_code 和result_code都为SUCCESS的时候有返回
	TradeType string `xml:"trade_type"` // 调用接口提交的交易类型 MICROPAY/NATIVE/APP/MWEB/JSAPI
	PrepayID  string `xml:"prepay_id"`  // 预支付交易会话标识 微信生成的预支付回话标识，用于后续接口调用中使用，该值有效期为2小时,针对H5支付此参数无特殊用途

	// 其他情况返回
	CodeURL string `xml:"code_url"` // 【NATIVE扫码支付返回】二维码链接 trade_type=NATIVE时有返回，此url用于生成支付二维码，然后提供给用户进行扫码支付。
	MWebURL string `xml:"mweb_url"` // 【H5支付返回】支付跳转链接 mweb_url为拉起微信支付收银台的中间页面，可通过访问该url来拉起微信客户端，完成支付,mweb_url的有效期为5分钟
}

//=================================================================
//							[Response]查询订单
//=================================================================
type WxOrderQueryResponse struct {
	ResponseBaseCode

	ResponseResultCodeSuccess

	TradeState string `xml:"trade_state"` // 交易状态 SUCCESS—支付成功, REFUND—转入退款, NOTPAY—未支付,CLOSED—已关闭,REVOKED—已撤销（刷卡支付）,USERPAYING--用户支付中,PAYERROR--支付失败
}

//=================================================================
//							[Response]退款
//=================================================================
type WxRefundResponse struct {
	ResponseBaseCode

	ResponseResultCodeSuccess

	OutRefundNo string `xml:"out_refund_no"` // 商户退款单号, 商户系统内部的退款单号
	RefundID    string `xml:"refund_id"`     // 微信退款单号
	RefundFee   int64  `xml:"refund_fee"`    // 微信退款金额
	TotalFee    int64  `xml:"total_fee"`     // 订单总金额
}

//=================================================================
//							[Response]退款查询
//=================================================================
type WxRefundQueryResponse struct {
	ResponseBaseCode

	ResponseResultCodeSuccess

	TotalRefundCount int64  `xml:"total_refund_count,emitempty"` // 订单总共已发生的部分退款次数，当请求参数传入offset后有返回
	TotalFee         int64  `xml:"total_fee"`                    // 订单总金额
	RefundCount      int64  `xml:"refund_count"`                 // 退款笔数
	OutRefundNo0     string `xml:"out_refund_no_0"`              // 商户退款单号0
	RefundID0        string `xml:"refund_id_0"`                  // 微信退款单号
	RefundFee0       int64  `xml:"refund_fee_0"`                 // 微信退款金额
	RefundStatus0    string `xml:"refund_status_0"`              // 微信退款状态,SUCCESS—退款成功,REFUNDCLOSE—退款关闭,PROCESSING—退款处理中,CHANGE—退款异常
}
