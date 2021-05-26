package wx

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	. "github.com/bmbstack/gopay/common"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var clients map[string]*WxClient

func init() {
	clients = make(map[string]*WxClient)
}

//===================================================================
//					   WxClient
//	微信官方文档 https://pay.weixin.qq.com/wiki/doc/api/index.html
//===================================================================
type WxClient struct {
	AppID       string // 应用ID
	MchID       string // 商户号
	ApiKey      string // API密钥，商户平台设置的密钥key
	ApiCertData []byte // API证书，微信支付接口中，涉及资金回滚的接口会使用到API证书，包括退款、撤销接口
	IsSandbox   bool   // 是否为沙盒环境
	SignType    string // 签名类型，目前支持HMAC-SHA256和MD5，默认为MD5
}

func AddWxClient(key string, client *WxClient) {
	clients[key] = client
}

func GetWxClient(key string) *WxClient {
	value, ok := clients[key]
	if !ok {
		panic("This WxClient not found")
	}
	return value
}

// Order 统一下单
func (client *WxClient) Order(chargeParam *ChargeParam) (*ChargeObject, error) {
	var requestUrl string
	if client.IsSandbox {
		requestUrl = SandboxUnifiedOrderUrl
	} else {
		requestUrl = UnifiedOrderUrl
	}

	params := make(map[string]string)
	params["trade_type"] = chargeParam.PayChannel                     // 【必传】支付类型，有：支付码付款-MICROPAY；扫二维码支付-NATIVE； APP支付(Android)-APP；H5支付(iOS)-MWEB；公众号和小程序支付-JSAPI
	params["out_trade_no"] = chargeParam.OrderID                      // 【必传】商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|*且在同一个商户号下唯一
	params["total_fee"] = strconv.FormatInt(chargeParam.TotalFee, 10) // 【必传】订单总金额，单位为分
	params["body"] = chargeParam.Description                          // 【必传】商品描述 APP: 需传入应用市场上的APP名字-实际商品名称，天天爱消除-游戏充值
	params["spbill_create_ip"] = chargeParam.ClientIP                 // 【必传】用户端实际ip
	params["notify_url"] = chargeParam.CallbackURL                    // 【必传】接收微信支付异步通知回调地址

	if strings.EqualFold(chargeParam.PayChannel, PayChannelWxH5) {
		params["scene_info"] = chargeParam.SceneInfo // 【H5必传】场景信息, iOS移动应用, Android移动应用, WAP网站应用
	}
	if strings.EqualFold(chargeParam.PayChannel, PayChannelWxJsapi) {
		params["openid"] = chargeParam.OpenID // 【JSAPI必传】OpenID
	}
	params = client.appendBasicParams(params)

	xmlStr, err := client.postWithXml(false, requestUrl, params)
	if err != nil {
		return nil, err
	}

	var respObject WxUnifiedOrderResponse
	err = xml.Unmarshal([]byte(xmlStr), &respObject)
	if err != nil {
		return nil, err
	}
	if respObject.ReturnCode != "SUCCESS" { // 通信失败
		return nil, errors.New(respObject.ReturnMsg)
	}
	if respObject.ResultCode != "SUCCESS" { // 支付失败
		return nil, errors.New(respObject.ErrCodeDes)
	}

	// ChargeObject
	object := &ChargeObject{}
	object.Status = OrderCreated // 1: 下单成功
	object.PrepayID = respObject.PrepayID

	if strings.EqualFold(chargeParam.PayChannel, PayChannelWxNative) {
		object.CodeURL = respObject.CodeURL
	} else if strings.EqualFold(chargeParam.PayChannel, PayChannelWxH5) {
		object.MWebURL = respObject.MWebURL
	}
	object.ChargeParam = chargeParam

	wxPayParam := make(map[string]string)
	if strings.EqualFold(chargeParam.PayChannel, PayChannelWxJsapi) {
		wxPayParam = map[string]string{
			"appId":     client.AppID,
			"timeStamp": strconv.FormatInt(time.Now().Unix(), 10),
			"nonceStr":  NonceStr(),
			"package":   fmt.Sprintf("prepay_id=%s", respObject.PrepayID),
			"signType":  client.SignType,
		}
		wxPayParam["paySign"] = client.sign(wxPayParam)
	} else if strings.EqualFold(chargeParam.PayChannel, PayChannelWxApp) {
		wxPayParam = map[string]string{
			"appid":     client.AppID,
			"partnerid": client.MchID,
			"prepayid":  respObject.PrepayID,
			"package":   "Sign=WXPay",
			"noncestr":  NonceStr(),
			"timestamp": strconv.FormatInt(time.Now().Unix(), 10),
		}
		wxPayParam["sign"] = client.sign(wxPayParam)
	} else {
		// nothing
	}

	object.PayParam = Marshal(wxPayParam)
	return object, nil
}

// OrderQuery 订单查询
func (client *WxClient) OrderQuery(orderQueryParam *OrderQueryParam) (*OrderQueryObject, error) {
	var requestUrl string
	if client.IsSandbox {
		requestUrl = SandboxOrderQueryUrl
	} else {
		requestUrl = OrderQueryUrl
	}

	params := make(map[string]string)
	params["out_trade_no"] = orderQueryParam.OrderID // 【必传】商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|*且在同一个商户号下唯一
	params = client.appendBasicParams(params)

	xmlStr, err := client.postWithXml(false, requestUrl, params)
	if err != nil {
		return nil, err
	}

	var respObject WxOrderQueryResponse
	err = xml.Unmarshal([]byte(xmlStr), &respObject)
	if err != nil {
		return nil, err
	}
	if respObject.ReturnCode != "SUCCESS" { // 通信失败
		return nil, errors.New(respObject.ReturnMsg)
	}
	if respObject.ResultCode != "SUCCESS" { // 支付失败
		return nil, errors.New(respObject.ErrCodeDes)
	}

	// OrderQueryObject
	object := &OrderQueryObject{
		OrderID:         respObject.OutTradeNO,
		Status:          mapTradeStateToStatus[respObject.TradeState],
		PayTime:         GetWxPayTime(respObject.TimeEnd),
		ThirdOrderID:    respObject.TransactionID,
		ThirdOrderFee:   respObject.TotalFee,
		OrderQueryParam: orderQueryParam,
	}
	return object, nil
}

// Refund 退款
func (client *WxClient) Refund(refundParam *RefundParam) (*RefundObject, error) {
	var requestUrl string
	if client.IsSandbox {
		requestUrl = SandboxRefundUrl
	} else {
		requestUrl = RefundUrl
	}

	params := make(map[string]string)
	params["out_trade_no"] = refundParam.OrderID                        // 【必传】商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|*且在同一个商户号下唯一
	params["total_fee"] = strconv.FormatInt(refundParam.OrderFee, 10)   // 【必传】订单金额, 单位：分
	params["out_refund_no"] = refundParam.RefundID                      // 【必传】商户系统内部退款单号
	params["refund_fee"] = strconv.FormatInt(refundParam.RefundFee, 10) // 【必传】退款金额, 单位：分
	params["refund_desc"] = refundParam.RefundDesc                      // 【必传】退款描述
	params["notify_url"] = refundParam.CallbackURL                      // 【非必传】回调地址，如果不传使用微信商户后台的url
	params = client.appendBasicParams(params)

	// 微信支付接口中，涉及资金回滚的接口会使用到API证书，包括退款、撤销接口。
	xmlStr, err := client.postWithXml(true, requestUrl, params)
	if err != nil {
		return nil, err
	}

	var respObject WxRefundResponse
	err = xml.Unmarshal([]byte(xmlStr), &respObject)
	if err != nil {
		return nil, err
	}
	if respObject.ReturnCode != "SUCCESS" { // 通信失败
		return nil, errors.New(respObject.ReturnMsg)
	}
	if respObject.ResultCode != "SUCCESS" { // 支付失败
		return nil, errors.New(respObject.ErrCodeDes)
	}

	// RefundObject
	object := &RefundObject{
		OrderID:        respObject.OutTradeNO,
		RefundID:       respObject.OutRefundNo,
		Status:         OrderRefunding,
		ThirdOrderID:   respObject.TransactionID,
		ThirdOrderFee:  respObject.TotalFee,
		ThirdRefundID:  respObject.RefundID,
		ThirdRefundFee: respObject.RefundFee,
		RefundParam:    refundParam,
	}
	return object, nil
}

// RefundQuery 退款查询
func (client *WxClient) RefundQuery(refundQueryParam *RefundQueryParam) (*RefundQueryObject, error) {
	var requestUrl string
	if client.IsSandbox {
		requestUrl = SandboxRefundQueryUrl
	} else {
		requestUrl = RefundQueryUrl
	}

	params := make(map[string]string)
	params["out_refund_no"] = refundQueryParam.RefundID // 【必传】商户系统内部的退款单号，商户系统内部唯一，只能是数字、大小写字母_-|*@ ，同一退款单号多次请求只退一笔。
	params = client.appendBasicParams(params)

	xmlStr, err := client.postWithXml(false, requestUrl, params)
	if err != nil {
		return nil, err
	}

	var respObject WxRefundQueryResponse
	err = xml.Unmarshal([]byte(xmlStr), &respObject)
	if err != nil {
		return nil, err
	}
	if respObject.ReturnCode != "SUCCESS" { // 通信失败
		return nil, errors.New(respObject.ReturnMsg)
	}
	if respObject.ResultCode != "SUCCESS" { // 支付失败
		return nil, errors.New(respObject.ErrCodeDes)
	}

	var orderRefundStatus = OrderRefunding
	if respObject.RefundStatus0 == "SUCCESS" {
		orderRefundStatus = OrderRefundSuccess
	} else if respObject.RefundStatus0 == "PROCESSING" {
		orderRefundStatus = OrderRefunding
	} else if respObject.RefundStatus0 == "CHANGE" {
		orderRefundStatus = OrderRefundFail
	} else if respObject.RefundStatus0 == "REFUNDCLOSE" {
		orderRefundStatus = OrderRefundFail
	} else {
		// nothing
	}

	// RefundQueryObject
	object := &RefundQueryObject{
		OrderID:          respObject.OutTradeNO,
		RefundID:         respObject.OutRefundNo0,
		Status:           int64(orderRefundStatus),
		ThirdOrderID:     respObject.TransactionID,
		ThirdOrderFee:    respObject.TotalFee,
		ThirdRefundID:    respObject.RefundID0,
		ThirdRefundFee:   respObject.RefundFee0,
		RefundQueryParam: refundQueryParam,
	}
	return object, nil
}

//===================================================
//		 Request; Params; Sign; Response
//===================================================
// 交易状态和Status映射
var mapTradeStateToStatus = map[string]int64{
	"NOTPAY":     OrderWaitPay,     // 2: 等待支付
	"USERPAYING": OrderUserPaying,  // 3: 用户支付中
	"SUCCESS":    OrderPaidSuccess, // 4: 支付成功
	"PAYERROR":   OrderPaidFail,    // 5: 支付失败
	"CLOSED":     OrderClosed,      // 6: 已关闭
}

func (client *WxClient) appendBasicParams(params map[string]string) map[string]string {
	params["appid"] = client.AppID        // 【必传】微信开放平台审核通过的应用APPID
	params["mch_id"] = client.MchID       // 【必传】微信支付分配的商户号
	params["nonce_str"] = NonceStr()      // 【必传】随机字符串，不长于32位
	params["sign_type"] = client.SignType // 【非必传】签名类型，目前支持HMAC-SHA256和MD5，默认为MD5
	params["sign"] = client.sign(params)  // 【必传】签名
	return params
}

// 请求
func (client *WxClient) postWithXml(useAppCert bool, url string, params map[string]string) (string, error) {
	hc, err := client.getHttpClient(useAppCert)
	if err != nil {
		return "", err
	}
	resp, err := hc.Post(url, MIMEApplicationXML, strings.NewReader(MapToXml(params)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}

func (client *WxClient) getHttpClient(useAppCert bool) (*http.Client, error) {
	var hc *http.Client
	if useAppCert { // 退款，退款查询需要app证书
		if client.ApiCertData == nil {
			return nil, errors.New("证书数据为空")
		}

		// 将pkcs12证书转成pem, API证书调用或安装需要使用到密码，该密码的值为微信商户号（mch_id）
		cert := Pkcs12ToPem(client.ApiCertData, client.MchID)
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		transport := &http.Transport{
			TLSClientConfig:    config,
			DisableCompression: true,
		}

		hc = &http.Client{Transport: transport}
	} else {
		hc = &http.Client{}
	}
	return hc, nil
}

// 签名
func (client *WxClient) sign(params map[string]string) string {
	delete(params, "sign")
	delete(params, "key")
	var paramArray []string
	for k, v := range params {
		if v != "" {
			paramArray = append(paramArray, fmt.Sprintf("%s=%s", k, v))
		}
	}

	sort.Strings(paramArray)
	paramStr := strings.Join(paramArray, "&")
	paramStr = paramStr + "&key=" + client.ApiKey

	var result string

	switch client.SignType {
	case SignTypeMd5:
		dataMd5 := md5.Sum([]byte(paramStr))
		result = hex.EncodeToString(dataMd5[:])
	case SignTypeHmacSha256:
		h := hmac.New(sha256.New, []byte(client.ApiKey))
		h.Write([]byte(paramStr))
		dataSha256 := h.Sum(nil)
		result = hex.EncodeToString(dataSha256[:])
	}

	return strings.ToUpper(result)
}

// 验证签名
func (client *WxClient) checkSign(params map[string]string) bool {
	value, ok := params[Sign]
	if !ok {
		return false
	}
	return value == client.sign(params)
}
