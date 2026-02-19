package dtos

type CommentResponse struct {
	ID        string `json:"id"`
	PostID    string `json:"postId"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

type CommentListResponse struct {
	Data []CommentResponse `json:"data"`
}

type CreateCommentRequest struct {
	PostID  string `json:"postId"`
	Author  string `json:"author"`
	Content string `json:"content"`
}
