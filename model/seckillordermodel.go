package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SeckillOrderModel = (*customSeckillOrderModel)(nil)

type (
	// SeckillOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSeckillOrderModel.
	SeckillOrderModel interface {
		seckillOrderModel
		withSession(session sqlx.Session) SeckillOrderModel
	}

	customSeckillOrderModel struct {
		*defaultSeckillOrderModel
	}
)

// NewSeckillOrderModel returns a model for the database table.
func NewSeckillOrderModel(conn sqlx.SqlConn) SeckillOrderModel {
	return &customSeckillOrderModel{
		defaultSeckillOrderModel: newSeckillOrderModel(conn),
	}
}

func (m *customSeckillOrderModel) withSession(session sqlx.Session) SeckillOrderModel {
	return NewSeckillOrderModel(sqlx.NewSqlConnFromSession(session))
}
