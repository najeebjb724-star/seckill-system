package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ActivityModel = (*customActivityModel)(nil)

type (
	// ActivityModel is an interface to be customized, add more methods here,
	// and implement the added methods in customActivityModel.
	ActivityModel interface {
		activityModel
		withSession(session sqlx.Session) ActivityModel
	}

	customActivityModel struct {
		*defaultActivityModel
	}
)

// NewActivityModel returns a model for the database table.
func NewActivityModel(conn sqlx.SqlConn) ActivityModel {
	return &customActivityModel{
		defaultActivityModel: newActivityModel(conn),
	}
}

func (m *customActivityModel) withSession(session sqlx.Session) ActivityModel {
	return NewActivityModel(sqlx.NewSqlConnFromSession(session))
}
