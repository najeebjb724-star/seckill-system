package types

type ActivityRequest struct {
	Name       string `json:"name"`
	Stock      int    `json:"stock"`
	Start_time string `json:"start_time"`
	End_time   string `json:"end_time"`
}

type IdResponse struct {
	Id int `json:"id"`
}

type MsgResponse struct {
	Msg string `json:"msg"`
}

type OrderStatusRequest struct {
	User_id     int `json:"user_id" form:"user_id"`
	Activity_id int `json:"activity_id" form:"activity_id"`
}

type SeckillRequest struct {
	User_id     int `json:"user_id" form:"user_id"`
	Activity_id int `json:"activity_id" form:"activity_id"`
}
