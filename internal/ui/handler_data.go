package ui

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
)

// fetchLicensePartialData fetches all data needed by the licenses partial.
func (h *UIHandler) fetchLicensePartialData(ctx context.Context, userID string) (gin.H, error) {
	sdkUser, err := h.client.Users().ByUserId(userID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	user := sdkUserToModel(sdkUser)

	subscribedSkus, err := h.fetchSubscribedSkus(ctx)
	if err != nil {
		subscribedSkus = []model.SubscribedSku{}
	}

	assignedSkuIDs := make(map[string]bool)
	for _, lic := range user.AssignedLicenses {
		assignedSkuIDs[lic.SkuID] = true
	}
	availableSkus := make([]model.SubscribedSku, 0)
	for _, sku := range subscribedSkus {
		if !assignedSkuIDs[sku.SkuID] {
			availableSkus = append(availableSkus, sku)
		}
	}

	return gin.H{
		"User":             user,
		"AssignedLicenses": user.AssignedLicenses,
		"AvailableSkus":    availableSkus,
	}, nil
}


// fetchAllUsers retrieves all users from Graph and converts them to model.User.
func (h *UIHandler) fetchAllUsers(ctx context.Context) []model.User {
	var allUsers []model.User
	sdkUsers, _ := h.client.Users().Get(ctx, nil)
	if sdkUsers != nil {
		for _, u := range sdkUsers.GetValue() {
			allUsers = append(allUsers, sdkUserToModel(u))
		}
	}
	return allUsers
}

// fetchMembersPartialData fetches all data needed by the members partial.
func (h *UIHandler) fetchMembersPartialData(ctx context.Context, groupID string) (gin.H, error) {
	sdkGroup, err := h.client.Groups().ByGroupId(groupID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	group := sdkGroupToModel(sdkGroup)

	members, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/groups/"+groupID+"/members")
	if err != nil {
		members = []model.DirectoryObject{}
	}

	allUsers := h.fetchAllUsers(ctx)

	return gin.H{
		"Group":    group,
		"Members":  members,
		"AllUsers": allUsers,
	}, nil
}

// fetchOwnersPartialData fetches all data needed by the owners partial.
func (h *UIHandler) fetchOwnersPartialData(ctx context.Context, groupID string) (gin.H, error) {
	sdkGroup, err := h.client.Groups().ByGroupId(groupID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	group := sdkGroupToModel(sdkGroup)

	owners, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/groups/"+groupID+"/owners")
	if err != nil {
		owners = []model.DirectoryObject{}
	}

	allUsers := h.fetchAllUsers(ctx)

	return gin.H{
		"Group":    group,
		"Owners":   owners,
		"AllUsers": allUsers,
	}, nil
}

// handleLicenseResponse sends either a partial HTML response (htmx) or a redirect (non-htmx).
func (h *UIHandler) handleLicenseResponse(c *gin.Context, userID string, level FlashLevel, message string) {
	if isHtmx(c) {
		data, err := h.fetchLicensePartialData(c.Request.Context(), userID)
		if err != nil {
			// Fall back to redirect if data fetch fails
			SetFlash(c, level, message)
			c.Redirect(http.StatusFound, "/ui/users/"+userID)
			return
		}
		data["Flash"] = &Flash{Level: level, Message: message}
		h.renderPartial(c, "licenses-partial", data)
		return
	}
	SetFlash(c, level, message)
	c.Redirect(http.StatusFound, "/ui/users/"+userID)
}

// handleMembersResponse sends either a partial HTML response (htmx) or a redirect (non-htmx).
func (h *UIHandler) handleMembersResponse(c *gin.Context, groupID string, level FlashLevel, message string) {
	if isHtmx(c) {
		data, err := h.fetchMembersPartialData(c.Request.Context(), groupID)
		if err != nil {
			SetFlash(c, level, message)
			c.Redirect(http.StatusFound, "/ui/groups/"+groupID)
			return
		}
		data["Flash"] = &Flash{Level: level, Message: message}
		h.renderPartial(c, "members-partial", data)
		return
	}
	SetFlash(c, level, message)
	c.Redirect(http.StatusFound, "/ui/groups/"+groupID)
}

// handleOwnersResponse sends either a partial HTML response (htmx) or a redirect (non-htmx).
func (h *UIHandler) handleOwnersResponse(c *gin.Context, groupID string, level FlashLevel, message string) {
	if isHtmx(c) {
		data, err := h.fetchOwnersPartialData(c.Request.Context(), groupID)
		if err != nil {
			SetFlash(c, level, message)
			c.Redirect(http.StatusFound, "/ui/groups/"+groupID)
			return
		}
		data["Flash"] = &Flash{Level: level, Message: message}
		h.renderPartial(c, "owners-partial", data)
		return
	}
	SetFlash(c, level, message)
	c.Redirect(http.StatusFound, "/ui/groups/"+groupID)
}
