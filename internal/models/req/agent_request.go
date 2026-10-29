package req

type CreateAgentReq struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
}

type UpdateAgentReq struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	Status      string `json:"status" form:"status"`
}
