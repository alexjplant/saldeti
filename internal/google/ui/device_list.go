package ui

import (
	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
)

func DeviceListHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		chromeOSDevices, err := h.client.GetChromeOSDevices(ctx, "my_customer")
		if err != nil {
			chromeOSDevices = []model.ChromeOSDevice{}
		}

		mobileDevices, err := h.client.GetMobileDevices(ctx, "my_customer")
		if err != nil {
			mobileDevices = []model.MobileDevice{}
		}

		h.render(c, "templates/devices/list.html", gin.H{
			"ActiveNav":        "devices",
			"ChromeOSDevices":  chromeOSDevices,
			"MobileDevices":    mobileDevices,
		})
	}
}