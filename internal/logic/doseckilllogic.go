// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"seckill-system/internal/svc"
	"seckill-system/internal/types"
	"seckill-system/model"

	"github.com/zeromicro/go-zero/core/logx"
)


type DoSeckillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDoSeckillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DoSeckillLogic {
	return &DoSeckillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DoSeckillLogic) DoSeckill(req *types.SeckillRequest) (resp *types.MsgResponse, err error) {
	// 1. 查询活动是否存在
	activity, err := l.svcCtx.ActivityModel.FindOne(l.ctx, int64(req.Activity_id))
	if err != nil {
		return &types.MsgResponse{
			Msg: "活动不存在",
		}, nil
	}

	// 2. 判断活动时间
	now := time.Now()
	if now.Before(activity.StartTime) {
		return &types.MsgResponse{
			Msg: "活动未开始",
		}, nil
	}
	if now.After(activity.EndTime) {
		return &types.MsgResponse{
			Msg: "活动已结束",
		}, nil
	}

	// 3. 判断用户是否已经参与过
	userKey := fmt.Sprintf("seckill:user:%d:%d", req.Activity_id, req.User_id)
	exist, _ := l.svcCtx.Redis.Get(userKey)
	if exist != "" {
		return &types.MsgResponse{
			Msg: "不能重复下单",
		}, nil
	}

	// 4. Redis 预减库存
	stockKey := fmt.Sprintf("seckill:stock:%d", req.Activity_id)
	stockStr, err := l.svcCtx.Redis.Get(stockKey)
	if err != nil || stockStr == "" {
		return &types.MsgResponse{
			Msg: "活动库存不存在",
		}, nil
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		return &types.MsgResponse{
			Msg: "库存数据异常",
		}, nil
	}
	if stock <= 0 {
		return &types.MsgResponse{
			Msg: "库存不足",
		}, nil
	}

	// 5. 扣减 Redis 库存
	err = l.svcCtx.Redis.Set(stockKey, strconv.Itoa(stock-1))
	if err != nil {
		return nil, err
	}

	// 6. 记录用户已经参与过
	err = l.svcCtx.Redis.Set(userKey, "1")
	if err != nil {
		return nil, err
	}

	// 7. 写入 MySQL 订单表
	_, err = l.svcCtx.SeckillOrderModel.Insert(l.ctx, &model.SeckillOrder{
		UserId:     int64(req.User_id),
		ActivityId: int64(req.Activity_id),
		Status:     "success",
	})
	if err != nil {
		return nil, err
	}

	// 8. 返回秒杀成功
	return &types.MsgResponse{
		Msg: "秒杀成功",
	}, nil
}
