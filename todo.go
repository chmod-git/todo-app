package todo

type TodoList struct {
	Id          int    `json:"id"`
	UserId      int    `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TodoItem struct {
	Id          int    `json:"id"`
	ListId      int    `json:"list_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}
