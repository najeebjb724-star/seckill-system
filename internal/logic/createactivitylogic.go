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

type CreateActivityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateActivityLogic {
	return &CreateActivityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateActivityLogic) CreateActivity(req *types.ActivityRequest) (resp *types.IdResponse, err error) {
	// 1. 校验库存
	if req.Stock <= 0 {
		return nil, fmt.Errorf("库存必须大于 0")
	}

	// 2. 解析时间
	startTime, err := time.Parse("2006-01-02 15:04:05", req.Start_time)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误")
	}
	endTime, err := time.Parse("2006-01-02 15:04:05", req.End_time)
	if err != nil {
		return nil, fmt.Errorf("结束时间格式错误")
	}
	if !endTime.After(startTime) {
		return nil, fmt.Errorf("结束时间必须晚于开始时间")
	}

	// 3. 写入 MySQL
	result, err := l.svcCtx.ActivityModel.Insert(l.ctx, &model.Activity{
		Name:      req.Name,
		Stock:     int64(req.Stock),
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		return nil, err
	}

	// 4. 获取活动 ID
	activityId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 5. Redis 初始化库存
	stockKey := fmt.Sprintf("seckill:stock:%d", activityId)
	err = l.svcCtx.Redis.Set(stockKey, strconv.Itoa(req.Stock))
	if err != nil {
		return nil, err
	}

	// 6. 返回 ID
	return &types.IdResponse{
		Id: int(activityId),
	}, nil
}
