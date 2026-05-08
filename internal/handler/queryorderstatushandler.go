package handler

import (
	"net/http"
	"strconv"

	"seckill-system/internal/logic"
	"seckill-system/internal/svc"
	"seckill-system/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func queryOrderStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrderStatusRequest
		
		// 手动从 URL 获取参数
		if v := r.URL.Query().Get("user_id"); v != "" {
			req.User_id, _ = strconv.Atoi(v)
		}
		if v := r.URL.Query().Get("activity_id"); v != "" {
			req.Activity_id, _ = strconv.Atoi(v)
		}

		l := logic.NewQueryOrderStatusLogic(r.Context(), svcCtx)
		resp, err := l.QueryOrderStatus(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
