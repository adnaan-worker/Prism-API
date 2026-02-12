package user

import "time"

// GetUsersRequest 获取用户列表请求
type GetUsersRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive banned"`
	Search   string `form:"search" binding:"omitempty"`
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive banned"`
}

// UpdateUserQuotaRequest 更新用户配额请求
type UpdateUserQuotaRequest struct {
	Quota int64 `json:"quota" binding:"required,min=0"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID         uint       `json:"id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Quota      int64      `json:"quota"`
	UsedQuota  int64      `json:"used_quota"`
	IsAdmin    bool       `json:"is_admin"`
	Status     string     `json:"status"`
	LastSignIn *time.Time `json:"last_sign_in,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// GetUsersResponse 获取用户列表响应
type GetUsersResponse struct {
	Users    []*UserResponse `json:"users"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// ToResponse 转换为响应对象
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:         u.ID,
		Username:   u.Username,
		Email:      u.Email,
		Quota:      u.Quota,
		UsedQuota:  u.UsedQuota,
		IsAdmin:    u.IsAdmin,
		Status:     u.Status,
		LastSignIn: u.LastSignIn,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

// ToResponseList 批量转换为响应对象
func ToResponseList(users []*User) []*UserResponse {
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}
	return responses
}
