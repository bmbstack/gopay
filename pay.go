package gopay

import (
	"errors"
	"github.com/bmbstack/gopay/wx"
	. "github.com/bmbstack/gopay/common"
	"strings"
	"gopkg.in/go-playground/validator.v9"
	"github.com/bmbstack/gopay/alipay"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// PayClient
type PayClient interface {
	Order(chargeParam *ChargeParam) (*ChargeObject, error)
	OrderQuery(orderQueryParam *OrderQueryParam) (*OrderQueryObject, error)
	Refund(refundParam *RefundParam) (*RefundObject, error)
	RefundQuery(refundQueryParam *RefundQueryParam) (*RefundQueryObject, error)
}

// Order
func Order(clientKey string, param *ChargeParam) (*ChargeObject, error) {
	err := validate.Struct(param)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(param.PayChannel, PayChannelWxJsapi) && IsEmpty(param.OpenID) {
		return nil, errors.New("JSAPI, openID is NULL")
	}

	pc := getPayClient(clientKey, param.PayType)
	object, err := pc.Order(param)
	if err != nil {
		return nil, err
	}
	return object, err
}

// OrderQuery
func OrderQuery(clientKey string, param *OrderQueryParam) (*OrderQueryObject, error) {
	err := validate.Struct(param)
	if err != nil {
		return nil, err
	}

	pc := getPayClient(clientKey, param.PayType)
	object, err := pc.OrderQuery(param)
	if err != nil {
		return nil, err
	}
	return object, err
}

// Refund
func Refund(clientKey string, param *RefundParam) (*RefundObject, error) {
	err := validate.Struct(param)
	if err != nil {
		return nil, err
	}

	pc := getPayClient(clientKey, param.PayType)
	object, err := pc.Refund(param)
	if err != nil {
		return nil, err
	}
	return object, err
}

// RefundQuery
func RefundQuery(clientKey string, param *RefundQueryParam) (*RefundQueryObject, error) {
	err := validate.Struct(param)
	if err != nil {
		return nil, err
	}

	pc := getPayClient(clientKey, param.PayType)
	object, err := pc.RefundQuery(param)
	if err != nil {
		return nil, err
	}
	return object, err
}

func getPayClient(clientKey string, payType string) PayClient {
	var pc PayClient
	switch payType {
	case PayTypeWx:
		pc = wx.GetWxClient(clientKey)
	case PayTypeAlipay:
		pc = alipay.GetAlipayClient(clientKey)
	default:
	}
	return pc
}
