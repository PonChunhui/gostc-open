package service

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"server/model"
	"server/pkg/bean"
	"server/pkg/jwt"
	"server/repository"
	paycommon "server/service/common/pay"
)

type CreateReq struct {
	Amount   string `json:"amount" binding:"required"`
	Kind     string `json:"kind"`
	TargetCode string `json:"targetCode"`
	Provider string `json:"provider" binding:"required"`
	Subject  string `json:"subject"`
}

type CreateResp struct {
	Code   string `json:"code"`
	PayURL string `json:"payUrl"`
	Amount string `json:"amount"`
}

func (s *service) Create(claims jwt.Claims, req CreateReq) (CreateResp, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || !amount.IsPositive() {
		return CreateResp{}, errors.New("金额无效")
	}
	provider, ok := paycommon.Get(req.Provider)
	if !ok {
		return CreateResp{}, errors.New("支付渠道不存在")
	}
	kind := req.Kind
	if kind == "" {
		kind = model.ORDER_KIND_RECHARGE
	}
	if kind == model.ORDER_KIND_GROUP && req.TargetCode == "" {
		return CreateResp{}, errors.New("购买套餐需指定用户组")
	}
	subject := req.Subject
	if subject == "" {
		subject = "账户充值"
	}

	db, _, _ := repository.Get("")
	order := &model.SystemOrder{
		UserCode:   claims.Code,
		Subject:    subject,
		Kind:       kind,
		TargetCode: req.TargetCode,
		Amount:     amount,
		Provider:   req.Provider,
		Status:     model.ORDER_STATUS_PENDING,
		ExpAt:      time.Now().Add(30 * time.Minute).Unix(),
	}
	if err := db.SystemOrder.Create(order); err != nil {
		return CreateResp{}, errors.New("创建订单失败")
	}
	payURL, err := provider.CreateOrder(order)
	if err != nil {
		_, _ = db.SystemOrder.Where(db.SystemOrder.Code.Eq(order.Code)).UpdateSimple(db.SystemOrder.Status.Value(model.ORDER_STATUS_FAILED))
		return CreateResp{}, err
	}
	return CreateResp{Code: order.Code, PayURL: payURL, Amount: amount.String()}, nil
}

type QueryReq struct {
	Code string `json:"code" binding:"required"`
}

func (s *service) Query(claims jwt.Claims, req QueryReq) (*model.SystemOrder, error) {
	db, _, _ := repository.Get("")
	order, err := db.SystemOrder.Where(db.SystemOrder.Code.Eq(req.Code), db.SystemOrder.UserCode.Eq(claims.Code)).First()
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	// 回调丢失补偿：待支付订单主动查单
	if order.Status == model.ORDER_STATUS_PENDING {
		if provider, ok := paycommon.Get(order.Provider); ok {
			if status, tradeNo, qerr := provider.QueryOrder(order.Code); qerr == nil && status == model.ORDER_STATUS_PAID {
				_ = SettleOrder(order.Code, tradeNo, order.Amount)
				order, _ = db.SystemOrder.Where(db.SystemOrder.Code.Eq(req.Code)).First()
			}
		}
	}
	return order, nil
}

type PageReq struct {
	bean.PageParam
	Status int `json:"status"`
}

func (s *service) Page(claims jwt.Claims, req PageReq) ([]*model.SystemOrder, int64) {
	db, _, _ := repository.Get("")
	q := db.SystemOrder.Where(db.SystemOrder.UserCode.Eq(claims.Code)).Order(db.SystemOrder.Id.Desc())
	if req.Status > 0 {
		q = q.Where(db.SystemOrder.Status.Eq(req.Status))
	}
	list, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	return list, total
}
