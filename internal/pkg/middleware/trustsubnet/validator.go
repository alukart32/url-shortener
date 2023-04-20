// Package trustsubnet provides gin middleware to validate a remote subnet request address.
package trustsubnet

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// validator is a representation of the handler that validate a remote subnet request address.
type validator struct {
	subnet *net.IPNet
}

// Validator return a new subnet validator.
func Validator(subnet string) (*validator, error) {
	if len(subnet) == 0 {
		subnet = os.Getenv("TRUSTED_SUBNET")
	}

	var ipnet *net.IPNet
	if len(subnet) != 0 {
		var err error
		if _, ipnet, err = net.ParseCIDR(subnet); err != nil {
			return nil, fmt.Errorf("failed to parse: %v", err)
		}
	}

	return &validator{
		subnet: ipnet,
	}, nil
}

// Handle processes the incoming request remote address.
func (v *validator) Handle(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if v.subnet == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		ipStr := c.Request.Header.Get("X-Real-IP")
		ip := net.ParseIP(ipStr)
		if ip == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if !v.subnet.Contains(ip) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		next(c)
	}
}
