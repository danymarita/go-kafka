package app

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Product struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type OrderReq struct {
	User     User    `json:"user"`
	Product  Product `json:"product"`
	Quantity uint    `json:"quantity"`
}

type SendEmail struct {
	User    User   `json:"user"`
	Message string `json:"message"`
}
