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
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Template    string    `json:"template"`
	ColorScheme Color     `json:"color_scheme"`
	Logo        string    `json:"logo"`
	Banner      string    `json:"banner"`
	OwnerID     uint      `json:"owner_id"`
	Products    []Product `json:"products"`
}

type Product struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Quantity    int     `json:"quantity"`
}

type Color struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
	Accent     string `json:"accent"`
	Supporting string `json:"supporting"`
	Tertiary   string `json:"tertiary"`
	Highlight  string `json:"highlight"`
}
