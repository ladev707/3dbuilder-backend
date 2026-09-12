package customer

import (
	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	CustomerService service.CustomerService
}

func NewCustomerHandler(customerService service.CustomerService) *CustomerHandler {
	return &CustomerHandler{CustomerService: customerService}
}

// func (h *ClientHandler) Register(c *gin.Context) {
// 	var user models.User
// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// }
