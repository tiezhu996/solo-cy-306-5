package dto

// CommentCreateRequest 评论请求。
type CommentCreateRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Content string `json:"content" binding:"max=2000"`
}
