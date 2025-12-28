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
	ID                    uint      `json:"id"`
	Name                  string    `json:"name"`
	Description           string    `json:"description"`
	Template              string    `json:"template"`
	ColorScheme           Color     `json:"color_scheme"`
	Logo                  string    `json:"logo"`
	Banner                string    `json:"banner"`
	OwnerID               uint      `json:"owner_id"`
	OwnerEmail            string    `json:"owner_email"`
	Products              []Product `json:"products"`
	IsOwner               bool      `json:"is_owner"`
	Phone                 string    `json:"phone"`
	Address               string    `json:"address"`
	PaymentMethods        string    `json:"payment_methods"`
	IBANNumber            string    `json:"iban_number"`
	ShippingInfo          string    `json:"shipping_info"`
	ShippingCost          float64   `json:"shipping_cost"`
	EstimatedShipping     int       `json:"estimated_shipping"`
	FreeShippingThreshold float64   `json:"free_shipping_threshold"`
	IsLoggedIn            bool      `json:"is_logged_in"`
}

type Product struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	Image         string   `json:"image"`
	Quantity      int      `json:"quantity"`
	Reviews       []Review `json:"reviews,omitempty"`
	AverageRating float64  `json:"average_rating,omitempty"`
	ReviewCount   int      `json:"review_count,omitempty"`
}

type Review struct {
	ID        uint   `json:"id"`
	ProductID uint   `json:"product_id"`
	UserID    uint   `json:"user_id"`
	OrderID   uint   `json:"order_id"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
	UserName  string `json:"user_name,omitempty"`
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
