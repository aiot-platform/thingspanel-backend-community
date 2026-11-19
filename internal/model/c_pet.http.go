package model

// C 端宠物信息（按绑定关系聚合 宠物档案 + 设备 + 门店）
type CEndPetItem struct {
	DeviceID     string `json:"device_id"`
	DeviceNumber string `json:"device_number"`
	DeviceName   string `json:"device_name"`
	PetName      string `json:"pet_name"`
	Breed        string `json:"breed"`
	Age          string `json:"age"`
	Avatar       string `json:"avatar"`
	BoardingDays int32  `json:"boarding_days"`
	SolutionCode string `json:"solution_code"`
	TenantID     string `json:"tenant_id"`
	ShopName     string `json:"shop_name"`
	OpenTime     string `json:"open_time"`
	CloseTime    string `json:"close_time"`
	BoundAt      string `json:"bound_at"`
}
