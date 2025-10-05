package api

//Error Response
type ErrorResponse struct {
	ErrorNbr int    `json:"error_number"`
	ErrorMsg string `json:"error_message"`
}

//Store and Store Display
type StoreDisplay struct {
	MyStores  []Store `json:"my_stores"`
	AllStores []Store `json:"all_stores"`
}

type Store struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
