package alipay

import (
	"crypto/rsa"
	. "github.com/bmbstack/gopay/common"
	"fmt"
	"sort"
	"strings"
	"crypto/sha1"
	"crypto/rand"
	"crypto"
	"net/url"
	"encoding/base64"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"time"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"encoding/pem"
)

var clients map[string]*AlipayClient

func init() {
	clients = make(map[string]*AlipayClient)
}

//================================================================================
//					   AlipayClient
//	支付宝官方文档 https://docs.open.alipay.com/200
//
//  请求：商户使用商户RSA私钥进行签名，发送请求，支付宝使用商户RSA公钥验证签名sign
//  响应：支付宝使用支付宝RSA私钥进行签名，发送响应结果，商家使用支付宝RSA公钥验证签名sign
//================================================================================
type AlipayClient struct {
	AppID           string // 应用ID
	PartnerID       string // 合作者ID
	SellerID        string // 卖家ID
	AppAuthToken    string // app auth token, 例如查询订单、退款、退款查询需要使用，下单不需要
	MchPrivateKey   []byte // 商户RSA私钥
	MchPublicKey    []byte // 商户RSA公钥(商户不会使用, 支付宝服务端使用)
	AlipayPublicKey []byte // 支付宝RSA公钥
	IsSandbox       bool   // 是否为沙盒环境
	SignType        string // 签名类型，RSA和RSA2 (RSA=>PKCS1, RSA2=>PKCS8)
}

func AddAlipayClient(key string, client *AlipayClient) {
	clients[key] = client
}

func GetAlipayClient(key string) *AlipayClient {
	value, ok := clients[key]
	if !ok {
		panic("This AlipayClient not found")
	}
	return value
}

// 下单(生成支付参数) https://docs.open.alipay.com/api_1/alipay.trade.app.pay
func (client *AlipayClient) Order(chargeParam *ChargeParam) (*ChargeObject, error) {
	var requestUrl string
	if client.IsSandbox {
		requestUrl = SandboxApiDomain
	} else {
		requestUrl = ApiDomain
	}

	// ChargeObject
	object := &ChargeObject{}
	object.Status = OrderCreated
	object.ChargeParam = chargeParam

	// 订单金额转换 (分=>元)
	totalAmountYuan := fmt.Sprintf("%.2f", float64(chargeParam.TotalFee)/float64(100))

	params := make(map[string]string)
	params["notify_url"] = chargeParam.CallbackURL
	if strings.EqualFold(chargeParam.PayChannel, PayChannelAlipayApp) {
		params["method"] = ApiNameTradeAppPay
		params["biz_content"] = Marshal(map[string]string{
			"out_trade_no": chargeParam.OrderID,     // 商户订单号，64个字符以内、可包含字母、数字、下划线；需保证在商户端不重复
			"total_amount": totalAmountYuan,         // 订单总金额，单位为元，精确到小数点后两位
			"subject":      chargeParam.Description, // 订单标题
			"product_code": DefaultProductCodeApp,   // 销售产品码，商家和支付宝签约的产品码
		})
		params = client.appendBasicParams(params)

		// 支付参数
		object.PayParam = MapToUrlValues(params).Encode()
	} else if strings.EqualFold(chargeParam.PayChannel, PayChannelAlipayH5) {
		params["method"] = ApiNameTradeWapPay
		params["biz_content"] = Marshal(map[string]string{
			"out_trade_no": chargeParam.OrderID,     // 商户订单号，64个字符以内、可包含字母、数字、下划线；需保证在商户端不重复
			"total_amount": totalAmountYuan,         // 订单总金额，单位为元，精确到小数点后两位
			"subject":      chargeParam.Description, // 订单标题
			"product_code": DefaultProductCodeWap,   // 销售产品码，商家和支付宝签约的产品码
			"quit_url":     chargeParam.ReturnURL,   // 用户付款中途退出返回商户网站的地址
		})
		params = client.appendBasicParams(params)

		hc := &http.Client{}
		paramStr := MapToUrlValues(params).Encode()
		req, err := http.NewRequest("POST", requestUrl, strings.NewReader(paramStr))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", MIMEApplicationForm)
		resp, err := hc.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// 支付参数
		object.PayParam = resp.Request.URL.String()
	} else {
		return nil, errors.New(fmt.Sprintf("alipay not support pay channel: %s", chargeParam.PayChannel))
	}

	return object, nil
}

// 订单查询 https://docs.open.alipay.com/api_1/alipay.trade.query
func (client *AlipayClient) OrderQuery(orderQueryParam *OrderQueryParam) (*OrderQueryObject, error) {
	var requestUrl string
	if client.IsSandbox {
		requestUrl = SandboxApiDomain
	} else {
		requestUrl = ApiDomain
	}

	params := make(map[string]string)
	params["method"] = ApiNameTradeQuery
	params["app_auth_token"] = client.AppAuthToken //  查询订单、退款、退款查询需要使用，下单不需要
	params["biz_content"] = Marshal(map[string]string{
		"out_trade_no": orderQueryParam.OrderID, // 订单支付时传入的商户订单号
	})
	params = client.appendBasicParams(params)

	var respObject *AlipayTradeQueryResponse
	err := client.postWithForm(requestUrl, params, &respObject)
	if err != nil {
		return nil, err
	}
	if !respObject.IsSuccess() {
		return nil, errors.New(respObject.Msg())
	}

	// OrderQueryObject
	thirdOrderAmount, _ := strconv.ParseInt(respObject.AlipayTradeQuery.ReceiptAmount, 10, 64)
	object := &OrderQueryObject{
		OrderID:         respObject.AlipayTradeQuery.OutTradeNo,
		Status:          mapTradeStateToStatus[respObject.AlipayTradeQuery.TradeStatus],
		ThirdOrderID:    respObject.AlipayTradeQuery.TradeNo,
		ThirdOrderFee:   thirdOrderAmount * 100, // 元=>分
		OrderQueryParam: orderQueryParam,
	}
	return object, nil
}

// 退款 https://docs.open.alipay.com/api_1/alipay.trade.refund
func (client *AlipayClient) Refund(refundParam *RefundParam) (*RefundObject, error) {
	var requestUrl string
	if client.IsSandbox {
		requestUrl = SandboxApiDomain
	} else {
		requestUrl = ApiDomain
	}

	// 订单金额转换 (分=>元)
	refundAmountYuan := fmt.Sprintf("%.2f", float64(refundParam.RefundFee)/float64(100))

	params := make(map[string]string)
	params["method"] = ApiNameTradeRefund
	params["app_auth_token"] = client.AppAuthToken //  查询订单、退款、退款查询需要使用，下单不需要
	params["biz_content"] = Marshal(map[string]string{
		"out_trade_no":   refundParam.OrderID,    // 订单支付时传入的商户订单号
		"refund_amount":  refundAmountYuan,       // 需要退款的金额，该金额不能大于订单金额,单位为元，支持两位小数
		"refund_reason":  refundParam.RefundDesc, // 退款的原因说明
		"out_request_no": refundParam.RefundID,   // 标识一次退款请求，同一笔交易多次退款需要保证唯一，如需部分退款，则此参数必传。
	})
	params = client.appendBasicParams(params)

	var respObject *AlipayTradeRefundResponse
	err := client.postWithForm(requestUrl, params, &respObject)
	if err != nil {
		return nil, err
	}
	if !respObject.IsSuccess() {
		return nil, errors.New(respObject.Msg())
	}

	// RefundObject
	thirdRefundAmount, _ := strconv.ParseInt(respObject.AlipayTradeRefund.RefundFee, 10, 64)
	object := &RefundObject{
		OrderID:        respObject.AlipayTradeRefund.OutTradeNo,
		RefundID:       refundParam.RefundID,
		Status:         OrderRefunding,
		ThirdOrderID:   respObject.AlipayTradeRefund.TradeNo,
		ThirdOrderFee:  refundParam.OrderFee,
		ThirdRefundID:  "",
		ThirdRefundFee: thirdRefundAmount * 100, // 元=>分
		RefundParam:    refundParam,
	}

	return object, nil
}

// 退款查询 https://docs.open.alipay.com/api_1/alipay.trade.fastpay.refund.query
func (client *AlipayClient) RefundQuery(refundQueryParam *RefundQueryParam) (*RefundQueryObject, error) {
	var requestUrl string
	if client.IsSandbox {
		requestUrl = SandboxApiDomain
	} else {
		requestUrl = ApiDomain
	}

	params := make(map[string]string)
	params["method"] = ApiNameTradeRefund
	params["app_auth_token"] = client.AppAuthToken //  查询订单、退款、退款查询需要使用，下单不需要
	params["biz_content"] = Marshal(map[string]string{
		"out_trade_no":   refundQueryParam.OrderID,  // 订单支付时传入的商户订单号
		"out_request_no": refundQueryParam.RefundID, // 请求退款接口时，传入的退款请求号
	})
	params = client.appendBasicParams(params)

	var respObject *AlipayFastpayTradeRefundQueryResponse
	err := client.postWithForm(requestUrl, params, &respObject)
	if err != nil {
		return nil, err
	}
	if !respObject.IsSuccess() {
		return nil, errors.New(respObject.Msg())
	}

	// RefundQueryObject
	var object *RefundQueryObject
	if IsNotEmpty(respObject.AlipayTradeRefundQuery) {
		thirdOrderAmount, _ := strconv.ParseInt(respObject.AlipayTradeRefundQuery.TotalAmount, 10, 64)
		thirdRefundAmount, _ := strconv.ParseInt(respObject.AlipayTradeRefundQuery.RefundAmount, 10, 64)
		object = &RefundQueryObject{
			OrderID:          respObject.AlipayTradeRefundQuery.OutTradeNo,
			RefundID:         refundQueryParam.RefundID,
			Status:           OrderRefundSuccess,
			ThirdOrderID:     respObject.AlipayTradeRefundQuery.TradeNo,
			ThirdOrderFee:    thirdOrderAmount * 100, // 元=>分
			ThirdRefundID:    "",
			ThirdRefundFee:   thirdRefundAmount * 100, // 元=>分
			RefundQueryParam: refundQueryParam,
		}
	} else {
		object = &RefundQueryObject{
			OrderID:          refundQueryParam.OrderID,
			RefundID:         refundQueryParam.RefundID,
			Status:           OrderRefundFail,
			ThirdOrderID:     "",
			ThirdOrderFee:    0,
			ThirdRefundID:    "",
			ThirdRefundFee:   0,
			RefundQueryParam: refundQueryParam,
		}
	}
	return object, nil
}

//===================================================
//		 Request; Params; Sign; Response
//===================================================
// 交易状态和Status映射
var mapTradeStateToStatus = map[string]int64{
	"WAIT_BUYER_PAY": OrderWaitPay,              // 2: 等待支付
	"TRADE_SUCCESS":  OrderPaidSuccess,          // 4: 支付成功
	"TRADE_CLOSED":   OrderClosed,               // 6: 未付款交易超时关闭，或支付完成后全额退款
	"TRADE_FINISHED": OrderFinishedCanNotRefund, // 11: 订单已完成，不能退款
}

// method, biz_content 自定义
func (client *AlipayClient) appendBasicParams(params map[string]string) map[string]string {
	params["app_id"] = client.AppID                         // 支付宝分配给开发者的应用ID
	params["format"] = "JSON"                               // 仅支持JSON
	params["charset"] = "utf-8"                             // 请求使用的编码格式，如utf-8,gbk,gb2312等
	params["timestamp"] = time.Now().Format(DateFullLayout) // 发送请求的时间，格式"yyyy-MM-dd HH:mm:ss"
	params["version"] = "1.0"                               // 调用的接口版本，固定为：1.0
	params["sign_type"] = client.SignType                   // 商户生成签名字符串所使用的签名算法类型，目前支持RSA2和RSA，推荐使用RSA2
	params["sign"] = client.sign(params)                    // 签名
	return params
}

// 请求(POST)
func (client *AlipayClient) postWithForm(url string, params map[string]string, respObject interface{}) error {
	hc := &http.Client{}
	paramStr := MapToUrlValues(params).Encode()
	req, err := http.NewRequest("POST", url, strings.NewReader(paramStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", MIMEApplicationForm)
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if IsEmpty(client.AlipayPublicKey) {
		return errors.New("alipay public key is empty")
	} else {
		// 使用支付宝RSA公钥验证签名sign的有效性
		var bodyStr = string(respBody)
		var rootNodeName = strings.Replace(params["method"], ".", "_", -1) + RespSuffix
		var rootIndex = strings.LastIndex(bodyStr, rootNodeName)
		var errorIndex = strings.LastIndex(bodyStr, RespError)

		var data string
		var sign string
		if rootIndex > 0 {
			data, sign = parseJSONSource(bodyStr, rootNodeName, rootIndex)
		} else if errorIndex > 0 {
			data, sign = parseJSONSource(bodyStr, RespError, errorIndex)
		} else {
			return nil
		}

		if IsNotEmpty(sign) {
			err = client.checkSign(data, sign)
			if err != nil {
				return err
			}
		}
	}

	return json.Unmarshal(respBody, respObject)
}

func parseJSONSource(bodyStr string, nodeName string, nodeIndex int) (content string, sign string) {
	var dataStartIndex = nodeIndex + len(nodeName) + 2
	var signIndex = strings.LastIndex(bodyStr, "\""+Sign+"\"")
	var dataEndIndex = signIndex - 1

	var indexLen = dataEndIndex - dataStartIndex
	if indexLen < 0 {
		return "", ""
	}
	content = bodyStr[dataStartIndex:dataEndIndex]

	var signStartIndex = signIndex + len(Sign) + 4
	sign = bodyStr[signStartIndex:]
	var signEndIndex = strings.LastIndex(sign, "\"}")
	sign = sign[:signEndIndex]

	return content, sign
}

// 签名
func (client *AlipayClient) sign(params map[string]string) string {
	delete(params, "sign_type")
	delete(params, "sign")
	var paramArray []string
	for k, v := range params {
		if v != "" {
			paramArray = append(paramArray, fmt.Sprintf("%s=%s", k, v))
		}
	}

	sort.Strings(paramArray)
	paramStr := strings.Join(paramArray, "&")

	block, _ := pem.Decode(client.MchPrivateKey)
	if block == nil {
		panic("Private Key block Error")
	}

	var signBytes []byte
	switch client.SignType {
	case SignTypeRSA: // 1024位, RSA => pkcs1格式
		s := sha1.New()
		s.Write([]byte(paramStr))
		contentBytes := s.Sum(nil)

		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			panic(err)
		}

		signBytes, err = rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, contentBytes)
		if err != nil {
			panic(err)
		}
	case SignTypeRSA2: // 2048位, RSA2 => pkcs8格式
		s := sha256.New()
		s.Write([]byte(paramStr))
		contentBytes := s.Sum(nil)

		privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			panic(err)
		}
		privateKey := privateKeyInterface.(*rsa.PrivateKey)

		signBytes, err = rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, contentBytes)
		if err != nil {
			panic(err)
		}
	}

	result := url.QueryEscape(base64.StdEncoding.EncodeToString(signBytes))
	return result
}

// 验证签名
func (client *AlipayClient) checkSign(data string, sign string) error {
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return errors.New("decode sign fail")
	}

	block, _ := pem.Decode(client.AlipayPublicKey)
	if block == nil {
		panic("Private Key block Error ")
	}

	var publicKey *rsa.PublicKey
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return errors.New("alipay public key is incorrect")
	}
	publicKey = publicKeyInterface.(*rsa.PublicKey)

	switch client.SignType {
	case SignTypeRSA: // 1024位, RSA => pkcs1格式
		s := sha1.New()
		s.Write([]byte(data))
		contentBytes := s.Sum(nil)
		err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, contentBytes, signBytes)
		if err != nil {
			return errors.New("verify fail")
		}
	case SignTypeRSA2: // 2048位, RSA2 => pkcs8格式
		s := sha256.New()
		s.Write([]byte(data))
		contentBytes := s.Sum(nil)
		err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, contentBytes, signBytes)
		if err != nil {
			return errors.New("verify fail")
		}
	}
	return nil
}
