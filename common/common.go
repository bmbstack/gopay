package common

// Pay channel

const (
	PayTypeWx     = "wx"     // 微信支付
	PayTypeAlipay = "alipay" // 支付宝支付
)
const (
	// 微信支付渠道
	PayChannelWxMicro  = "MICROPAY" // 付款码支付，付款码支付有单独的支付接口，所以接口不需要上传，该字段在对账单中会出现
	PayChannelWxNative = "NATIVE"   // Native支付（扫二维码支付）
	PayChannelWxApp    = "APP"      // App支付（Android端使用）
	PayChannelWxH5     = "MWEB"     // H5支付（iOS端使用），因为苹果审核禁止微信App支付
	PayChannelWxJsapi  = "JSAPI"    // JSAPI支付（公众号支付，小程序支付）

	// 支付宝支付渠道
	PayChannelAlipayApp = "APP"
	PayChannelAlipayH5  = "H5"
)

const (
	OrderWait                 = iota // 0 等待下单
	OrderCreated                     // 1 下单成功
	OrderWaitPay                     // 2 等待支付
	OrderUserPaying                  // 3 用户支付中
	OrderPaidSuccess                 // 4 支付成功
	OrderPaidFail                    // 5 支付失败
	OrderToRefund                    // 6 转入退款
	OrderRefunding                   // 7 退款中
	OrderRefundSuccess               // 8 退款成功
	OrderRefundFail                  // 9 退款失败
	OrderClosed                      // 10 订单已关闭
	OrderFinishedCanNotRefund        // 11 订单已完成，不能退款
)

//========================================
//              Charge Order
//========================================
// ChargeParam
type ChargeParam struct {
	PayType     string `json:"payType,omitempty" validate:"required"`     // 支付方式
	PayChannel  string `json:"payChannel,omitempty" validate:"required"`  // 支付渠道
	CallbackURL string `json:"callbackURL,omitempty" validate:"required"` // 支付回调地址

	OrderID     string `json:"orderID,omitempty" validate:"required"`     // 本地订单号
	TotalFee    int64  `json:"totalFee,omitempty" validate:"required"`    // 订单总金额 单位：分
	Description string `json:"description,omitempty" validate:"required"` // 订单描述
	ClientIP    string `json:"clientIP,omitempty" validate:"required"`    // 用户端实际ip

	OpenID    string `json:"openID,omitempty"`    // 微信openid
	SceneInfo string `json:"sceneInfo,omitempty"` // 微信对H5支付有以下三种场景, iOS移动应用, Android移动应用, WAP网站应用
	ReturnURL string `json:"returnURL,omitempty"` // 支付结果页, 阿里quit_url, 用户付款中途退出返回商户网站的地址
}

// ChargeObject
type ChargeObject struct {
	Status int64 `json:"status,omitempty"` // 支付状态， 0: 等待下单, 1: 下单成功, 2: 未支付, 3: 用户支付中, 4: 支付成功, 5: 支付失败, 6: 转入退款, 7: 退款中, 8: 退款成功, 9: 退款失败, 10: 订单已关闭, 默认为0

	PrepayID string `json:"prepayID,omitempty"` //【微信】预支付交易会话标识 微信生成的预支付回话标识，用于后续接口调用中使用，该值有效期为2小时,针对H5支付此参数无特殊用途
	CodeURL  string `json:"codeURL,omitempty"`  //【微信】二维码链接 trade_type=NATIVE时有返回，此url用于生成支付二维码，然后提供给用户进行扫码支付。
	MWebURL  string `json:"mwebURL,omitempty"`  //【微信】支付跳转链接 mweb_url为拉起微信支付收银台的中间页面，可通过访问该url来拉起微信客户端，完成支付,mweb_url的有效期为5分钟

	ChargeParam *ChargeParam `json:"chargeParam,omitempty"`

	PayParam string `json:"payParam,omitempty"` // Android/iOS客户端支付时需要的参数
}

//========================================
//              OrderQuery
//========================================
type OrderQueryParam struct {
	PayType    string `json:"payType,omitempty" validate:"required"`    // 支付方式
	PayChannel string `json:"payChannel,omitempty" validate:"required"` // 支付渠道

	OrderID string `json:"orderID,omitempty" validate:"required"` // 本地订单号
}

// OrderQueryObject
type OrderQueryObject struct {
	OrderID string `json:"orderID,omitempty"` // 本地订单号
	Status  int64  `json:"status,omitempty"`  // 支付状态， 0: 等待下单, 1: 下单成功, 2: 未支付, 3: 用户支付中, 4: 支付成功, 5: 支付失败, 6: 转入退款, 7: 退款中, 8: 退款成功, 9: 退款失败, 10: 订单已关闭, 默认为0

	ThirdOrderID  string `json:"thirdOrderID,omitempty"`  // 第三方订单单号(微信，支付宝)
	ThirdOrderFee int64  `json:"thirdOrderFee,omitempty"` // 第三方订单金额，单位：分(微信，支付宝)

	OrderQueryParam *OrderQueryParam `json:"orderQueryParam,omitempty"`
}

//========================================
//              Refund
//========================================
// RefundParam
type RefundParam struct {
	PayType     string `json:"payType,omitempty" validate:"required"`     // 支付方式
	PayChannel  string `json:"payChannel,omitempty" validate:"required"`  // 支付渠道
	CallbackURL string `json:"callbackURL,omitempty" validate:"required"` // 支付回调地址

	OrderID    string `json:"orderID,omitempty" validate:"required"`    // 本地订单号
	OrderFee   int64  `json:"orderFee,omitempty" validate:"required"`   // 订单总金额 单位：分
	RefundID   string `json:"refundID,omitempty" validate:"required"`   // 本地退款号
	RefundFee  int64  `json:"refundFee,omitempty" validate:"required"`  // 退款总金额 单位：分
	RefundDesc string `json:"refundDesc,omitempty" validate:"required"` // 退款描述
}

// RefundObject
type RefundObject struct {
	OrderID  string `json:"orderID,omitempty"`  // 本地订单号
	RefundID string `json:"refundID,omitempty"` // 本地退款号
	Status   int64  `json:"status,omitempty"`   // 支付状态， 0: 等待下单, 1: 下单成功, 2: 未支付, 3: 用户支付中, 4: 支付成功, 5: 支付失败, 6: 转入退款, 7: 退款中, 8: 退款成功, 9: 退款失败, 10: 订单已关闭, 默认为0

	ThirdOrderID  string `json:"thirdOrderID,omitempty"`  // 第三方订单单号(微信，支付宝)
	ThirdOrderFee int64  `json:"thirdOrderFee,omitempty"` // 第三方订单金额，单位：分(微信，支付宝)

	ThirdRefundID  string `json:"thirdRefundID,omitempty"`  // 第三方退款单号(微信，支付宝)
	ThirdRefundFee int64  `json:"thirdRefundFee,omitempty"` // 第三方退款金额，单位：分(微信，支付宝)

	RefundParam *RefundParam `json:"refundParam,omitempty"`
}

//========================================
//              RefundQuery
//========================================
// RefundQueryParam
type RefundQueryParam struct {
	PayType    string `json:"payType,omitempty" validate:"required"`    // 支付方式
	PayChannel string `json:"payChannel,omitempty" validate:"required"` // 支付渠道

	OrderID  string `json:"orderID,omitempty" validate:"required"`  // 本地订单号
	RefundID string `json:"refundID,omitempty" validate:"required"` // 本地退款号
}

// RefundQueryObject
type RefundQueryObject struct {
	OrderID  string `json:"orderID,omitempty"`  // 本地订单号
	RefundID string `json:"refundID,omitempty"` // 本地退款号
	Status   int64  `json:"status,omitempty"`   // 支付状态， 0: 等待下单, 1: 下单成功, 2: 未支付, 3: 用户支付中, 4: 支付成功, 5: 支付失败, 6: 转入退款, 7: 退款中, 8: 退款成功, 9: 退款失败, 10: 订单已关闭, 默认为0

	ThirdOrderID  string `json:"thirdOrderID,omitempty"`  // 第三方订单单号(微信，支付宝)
	ThirdOrderFee int64  `json:"thirdOrderFee,omitempty"` // 第三方订单金额，单位：分(微信，支付宝)

	ThirdRefundID  string `json:"thirdRefundID,omitempty"`  // 第三方退款单号(微信，支付宝)
	ThirdRefundFee int64  `json:"thirdRefundFee,omitempty"` // 第三方退款金额，单位：分(微信，支付宝)

	RefundQueryParam *RefundQueryParam `json:"refundQueryParam,omitempty"`
}
