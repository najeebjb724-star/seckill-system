// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"seckill-system/internal/svc"
	"seckill-system/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)


type QueryOrderStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOrderStatusLogic {
	return &QueryOrderStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryOrderStatusLogic) QueryOrderStatus(req *types.OrderStatusRequest) (resp *types.MsgResponse, err error) {
	order, err := l.svcCtx.SeckillOrderModel.FindOneByUserIdActivityId(
		l.ctx,
		int64(req.User_id),
		int64(req.Activity_id),
	)

	if err == sqlx.ErrNotFound {
		return &types.MsgResponse{
			Msg: "未查询到订单，可能未参与或秒杀失败",
		}, nil
	}

	if err != nil {
		return nil, err
	}

	return &types.MsgResponse{
		Msg: "订单状态：" + order.Status,
	}, nil
}
