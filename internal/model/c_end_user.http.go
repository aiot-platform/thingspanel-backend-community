package model

// C 端用户接口请求/响应结构体（独立于后台鉴权）

// CEndRegisterReq C 端注册
type CEndRegisterReq struct {
	Phone        string `json:"phone" validate:"required,max=20"`
	Password     string `json:"password" validate:"required,min=6,max=64"`
	Nickname     string `json:"nickname" validate:"omitempty,max=99"`
	SolutionCode string `json:"solution_code" validate:"omitempty,max=64"`
}

// CEndLoginReq C 端登录
type CEndLoginReq struct {
	Phone    string `json:"phone" validate:"required,max=20"`
	Password string `json:"password" validate:"required,max=64"`
}

// CEndUserInfo C 端用户公开信息
type CEndUserInfo struct {
	ID                  string `json:"id"`
	Phone               string `json:"phone"`
	Nickname            string `json:"nickname"`
	Avatar              string `json:"avatar"`
	DefaultSolutionCode string `json:"default_solution_code"`
}

// CEndLoginResp C 端登录响应
type CEndLoginResp struct {
	Token     string        `json:"token"`
	ExpiresIn int64         `json:"expires_in"`
	User      CEndUserInfo  `json:"user"`
}

// CEndBindDeviceReq C 端绑定设备（凭门店出码）
type CEndBindDeviceReq struct {
	Code string `json:"code" validate:"required,max=64"`
}

// CEndDeviceItem C 端已绑定设备
type CEndDeviceItem struct {
	DeviceID     string `json:"device_id"`
	DeviceNumber string `json:"device_number"`
	DeviceName   string `json:"device_name"`
	SolutionCode string `json:"solution_code"`
	TenantID     string `json:"tenant_id"`
	BoundAt      string `json:"bound_at"`
}
