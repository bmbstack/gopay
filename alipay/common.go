package alipay

const (
	SignTypeRSA           = "RSA"                 // 1024位, RSA => pkcs1格式
	SignTypeRSA2          = "RSA2"                // 2048位, RSA2 => pkcs8格式
	DefaultProductCodeApp = "QUICK_MSECURITY_PAY" // product code
	DefaultProductCodeWap = "QUICK_WAP_WAY"       // product code

	Sign       = "sign"
	RespSuffix = "_response"
	RespError  = "error_response"

	RespSuccessCode = "10000"

	MIMEApplicationForm = "application/x-www-form-urlencoded;charset=utf-8" // Content-Type

	ApiDomain  = "https://openapi.alipay.com/gateway.do"
	MApiDomain = "https://mapi.alipay.com/gateway.do"

	SandboxApiDomain = "https://openapi.alipaydev.com/gateway.do"

	ApiNameTradeAppPay      = "alipay.trade.app.pay"              // APP下订单，生成支付参数
	ApiNameTradeWapPay      = "alipay.trade.wap.pay"              // 手机网站下订单，生成支付参数
	ApiNameTradeQuery       = "alipay.trade.query"                // 订单查询
	ApiNameTradeRefund      = "alipay.trade.refund"               // 退款
	ApiNameTradeRefundQuery = "alipay.trade.fastpay.refund.query" // 退款查询
)

//=================================================================
//							[Response]查询订单
//=================================================================
type AlipayTradeQueryResponse struct {
	AlipayTradeQuery struct {
		Code             string `json:"code"`
		Msg              string `json:"msg"`
		SubCode          string `json:"sub_code"`
		SubMsg           string `json:"sub_msg"`
		AuthTradePayMode string `json:"auth_trade_pay_mode"` // 预授权支付模式，该参数仅在信用预授权支付场景下返回。信用预授权支付：CREDIT_PREAUTH_PAY
		BuyerLogonId     string `json:"buyer_logon_id"`      // 买家支付宝账号
		BuyerPayAmount   string `json:"buyer_pay_amount"`    // 买家实付金额，单位为元，两位小数。
		BuyerUserId      string `json:"buyer_user_id"`       // 买家在支付宝的用户id
		BuyerUserType    string `json:"buyer_user_type"`     // 买家用户类型。CORPORATE:企业用户；PRIVATE:个人用户。
		InvoiceAmount    string `json:"invoice_amount"`      // 交易中用户支付的可开具发票的金额，单位为元，两位小数。
		OutTradeNo       string `json:"out_trade_no"`        // 商家订单号
		PointAmount      string `json:"point_amount"`        // 积分支付的金额，单位为元，两位小数。
		ReceiptAmount    string `json:"receipt_amount"`      // 实收金额，单位为元，两位小数
		SendPayDate      string `json:"send_pay_date"`       // 本次交易打款给卖家的时间
		TotalAmount      string `json:"total_amount"`        // 交易的订单金额
		TradeNo          string `json:"trade_no"`            // 支付宝交易号
		TradeStatus      string `json:"trade_status"`        // 交易状态：WAIT_BUYER_PAY（交易创建，等待买家付款）、TRADE_CLOSED（未付款交易超时关闭，或支付完成后全额退款）、TRADE_SUCCESS（交易支付成功）、TRADE_FINISHED（交易结束，不可退款）

		DiscountAmount      string           `json:"discount_amount"`               // 平台优惠金额
		FundBillList        []*FundBill      `json:"fund_bill_list,omitempty"`      // 交易支付使用的资金渠道
		MdiscountAmount     string           `json:"mdiscount_amount"`              // 商家优惠金额
		PayAmount           string           `json:"pay_amount"`                    // 支付币种订单金额
		PayCurrency         string           `json:"pay_currency"`                  // 订单支付币种
		SettleAmount        string           `json:"settle_amount"`                 // 结算币种订单金额
		SettleCurrency      string           `json:"settle_currency"`               // 订单结算币种
		SettleTransRate     string           `json:"settle_trans_rate"`             // 结算币种兑换标价币种汇率
		StoreId             string           `json:"store_id"`                      // 商户门店编号
		StoreName           string           `json:"store_name"`                    // 请求交易支付中的商户店铺的名称
		TerminalId          string           `json:"terminal_id"`                   // 商户机具终端编号
		TransCurrency       string           `json:"trans_currency"`                // 标价币种
		TransPayRate        string           `json:"trans_pay_rate"`                // 标价币种兑换支付币种汇率
		DiscountGoodsDetail string           `json:"discount_goods_detail"`         // 本次交易支付所使用的单品券优惠的商品优惠信息
		IndustrySepcDetail  string           `json:"industry_sepc_detail"`          // 行业特殊信息（例如在医保卡支付业务中，向用户返回医疗信息）。
		VoucherDetailList   []*VoucherDetail `json:"voucher_detail_list,omitempty"` // 本交易支付时使用的所有优惠券信息
	} `json:"alipay_trade_query_response"`
	Sign string `json:"sign"`
}

type FundBill struct {
	FundChannel string  `json:"fund_channel"`       // 交易使用的资金渠道，详见 支付渠道列表
	Amount      string  `json:"amount"`             // 该支付工具类型所使用的金额
	RealAmount  float64 `json:"real_amount,string"` // 渠道实际付款金额
}

type VoucherDetail struct {
	Id                 string `json:"id"`                  // 券id
	Name               string `json:"name"`                // 券名称
	Type               string `json:"type"`                // 当前有三种类型： ALIPAY_FIX_VOUCHER - 全场代金券, ALIPAY_DISCOUNT_VOUCHER - 折扣券, ALIPAY_ITEM_VOUCHER - 单品优惠
	Amount             string `json:"amount"`              // 优惠券面额，它应该会等于商家出资加上其他出资方出资
	MerchantContribute string `json:"merchant_contribute"` // 商家出资（特指发起交易的商家出资金额）
	OtherContribute    string `json:"other_contribute"`    // 其他出资方出资金额，可能是支付宝，可能是品牌商，或者其他方，也可能是他们的一起出资
	Memo               string `json:"memo"`                // 优惠券备注信息
}

func (this *AlipayTradeQueryResponse) IsSuccess() bool {
	if this.AlipayTradeQuery.Code == RespSuccessCode {
		return true
	}
	return false
}

func (this *AlipayTradeQueryResponse) Msg() string {
	return this.AlipayTradeQuery.Msg + ", " + this.AlipayTradeQuery.SubMsg
}

//=================================================================
//							[Response]退款
//=================================================================
type AlipayTradeRefundResponse struct {
	AlipayTradeRefund struct {
		Code                 string              `json:"code"`
		Msg                  string              `json:"msg"`
		SubCode              string              `json:"sub_code"`
		SubMsg               string              `json:"sub_msg"`
		TradeNo              string              `json:"trade_no"`                          // 支付宝交易号
		OutTradeNo           string              `json:"out_trade_no"`                      // 商户订单号
		BuyerLogonId         string              `json:"buyer_logon_id"`                    // 用户的登录id
		BuyerUserId          string              `json:"buyer_user_id"`                     // 买家在支付宝的用户id
		FundChange           string              `json:"fund_change"`                       // 本次退款是否发生了资金变化
		RefundFee            string              `json:"refund_fee"`                        // 退款总金额
		GmtRefundPay         string              `json:"gmt_refund_pay"`                    // 退款支付时间
		StoreName            string              `json:"store_name"`                        // 交易在支付时候的门店名称
		RefundDetailItemList []*RefundDetailItem `json:"refund_detail_item_list,omitempty"` // 退款使用的资金渠道
	} `json:"alipay_trade_refund_response"`
	Sign string `json:"sign"`
}

type RefundDetailItem struct {
	FundChannel string `json:"fund_channel"` // 交易使用的资金渠道，详见 支付渠道列表
	Amount      string `json:"amount"`       // 该支付工具类型所使用的金额
	RealAmount  string `json:"real_amount"`  // 渠道实际付款金额
}

func (this *AlipayTradeRefundResponse) IsSuccess() bool {
	if this.AlipayTradeRefund.Code == RespSuccessCode {
		return true
	}
	return false
}

func (this *AlipayTradeRefundResponse) Msg() string {
	return this.AlipayTradeRefund.Msg + ", " + this.AlipayTradeRefund.SubMsg
}

//=================================================================
//							[Response]退款查询
//=================================================================
type AlipayFastpayTradeRefundQueryResponse struct {
	AlipayTradeRefundQuery struct {
		Code         string `json:"code"`
		Msg          string `json:"msg"`
		SubCode      string `json:"sub_code"`
		SubMsg       string `json:"sub_msg"`
		OutRequestNo string `json:"out_request_no"` // 本笔退款对应的退款请求号
		OutTradeNo   string `json:"out_trade_no"`   // 创建交易传入的商户订单号
		RefundReason string `json:"refund_reason"`  // 发起退款时，传入的退款原因
		TotalAmount  string `json:"total_amount"`   // 发该笔退款所对应的交易的订单金额
		RefundAmount string `json:"refund_amount"`  // 本次退款请求，对应的退款金额
		TradeNo      string `json:"trade_no"`       // 支付宝交易号
	} `json:"alipay_trade_fastpay_refund_query_response"`
	Sign string `json:"sign"`
}

func (this *AlipayFastpayTradeRefundQueryResponse) IsSuccess() bool {
	if this.AlipayTradeRefundQuery.Code == RespSuccessCode {
		return true
	}
	return false
}

func (this *AlipayFastpayTradeRefundQueryResponse) Msg() string {
	return this.AlipayTradeRefundQuery.Msg + ", " + this.AlipayTradeRefundQuery.SubMsg
}
