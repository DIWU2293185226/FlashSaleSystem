// ═════════════════════════════════════════════════════════════════════
// 用户信息 DTO
// ═════════════════════════════════════════════════════════════════════
package dto

// UserDTO 用户简要信息（不含手机号、密码等敏感字段）
// 常用于博客作者信息、关注列表等场景
type UserDTO struct {
	ID       int64  `json:"id"`
	NickName string `json:"nickName"`
	Icon     string `json:"icon"`
}
