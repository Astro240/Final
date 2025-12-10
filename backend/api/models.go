package api

//Error Response
type ErrorResponse struct {
	ErrorNbr int    `json:"error_number"`
	ErrorMsg string `json:"error_message"`
}

//Store and Store Display
type StoreDisplay struct {
	MyStores  []Store     `json:"my_stores"`
	AllStores []Store     `json:"all_stores"`
	User      UserProfile `json:"user"`
	ValidUser bool        `json:"valid_user"`
}

type UserProfile struct {
	Fullname       string `json:"full_name"`
	ProfilePicture string `json:"profile_picture"`
	Email          string `json:"email"`
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
	IsOwner     bool      `json:"is_owner"`
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

// Cart and Checkout
type CartItem struct {
	ID        int     `json:"id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Product   Product `json:"product"`
	ItemTotal float64 `json:"item_total"`
}

type CartResponse struct {
	Items      []CartItem `json:"items"`
	TotalItems int        `json:"total_items"`
	TotalPrice float64    `json:"total_price"`
	Store      Store      `json:"store"`
}

type PaymentPageData struct {
	Store Store `json:"store"`
}

type DashboardData struct {
	Store Store `json:"store"`
}
