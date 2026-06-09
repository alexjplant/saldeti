package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
	"github.com/saldeti/saldeti/internal/google/store"
)

// listTokensHandler handles GET /admin/directory/v1/users/:userKey/tokens
func listTokensHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		tokens, err := st.ListTokens(c.Request.Context(), userKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list tokens")
			}
			return
		}
		if tokens == nil {
			tokens = []model.Token{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#tokens",
			"etag":  "\"placeholder\"",
			"items": tokens,
		})
	}
}

// getTokenHandler handles GET /admin/directory/v1/users/:userKey/tokens/:clientId
func getTokenHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		clientID := c.Param("clientId")
		token, err := st.GetToken(c.Request.Context(), userKey, clientID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Token not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get token")
			}
			return
		}
		writeJSON(c, http.StatusOK, token)
	}
}

// deleteTokenHandler handles DELETE /admin/directory/v1/users/:userKey/tokens/:clientId
func deleteTokenHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		clientID := c.Param("clientId")
		if err := st.DeleteToken(c.Request.Context(), userKey, clientID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Token not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete token")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// listASPsHandler handles GET /admin/directory/v1/users/:userKey/asps
func listASPsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		asps, err := st.ListASPs(c.Request.Context(), userKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list ASPs")
			}
			return
		}
		if asps == nil {
			asps = []model.ASP{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#asps",
			"etag":  "\"placeholder\"",
			"items": asps,
		})
	}
}

// getASPHandler handles GET /admin/directory/v1/users/:userKey/asps/:codeId
func getASPHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		codeID := c.Param("codeId")
		asp, err := st.GetASP(c.Request.Context(), userKey, codeID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "ASP not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get ASP")
			}
			return
		}
		writeJSON(c, http.StatusOK, asp)
	}
}

// deleteASPHandler handles DELETE /admin/directory/v1/users/:userKey/asps/:codeId
func deleteASPHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		codeID := c.Param("codeId")
		if err := st.DeleteASP(c.Request.Context(), userKey, codeID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "ASP not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete ASP")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// listVerificationCodesHandler handles GET /admin/directory/v1/users/:userKey/verificationCodes
func listVerificationCodesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		codes, err := st.ListVerificationCodes(c.Request.Context(), userKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list verification codes")
			}
			return
		}
		if codes == nil {
			codes = []model.VerificationCode{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#verificationCodes",
			"etag":  "\"placeholder\"",
			"items": codes,
		})
	}
}

// generateVerificationCodesHandler handles POST /admin/directory/v1/users/:userKey/verificationCodes/generate
func generateVerificationCodesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		if err := st.GenerateVerificationCodes(c.Request.Context(), userKey); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to generate verification codes")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// invalidateVerificationCodesHandler handles POST /admin/directory/v1/users/:userKey/verificationCodes/invalidate
func invalidateVerificationCodesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		if err := st.InvalidateVerificationCodes(c.Request.Context(), userKey); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to invalidate verification codes")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// turnOff2SVHandler handles POST /admin/directory/v1/users/:userKey/twoStepVerification/turnOff
func turnOff2SVHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		if err := st.TurnOff2SV(c.Request.Context(), userKey); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to turn off 2SV")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// listUserInvitationsHandler handles GET /v1/customers/:customer/userinvitations
func listUserInvitationsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customer := c.Param("customer")
		parent := "customers/" + customer
		invitations, err := st.ListUserInvitations(c.Request.Context(), parent)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Customer not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list user invitations")
			}
			return
		}
		if invitations == nil {
			invitations = []model.UserInvitation{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"invitations": invitations,
		})
	}
}

// getUserInvitationHandler handles GET /v1/customers/:customer/userinvitations/:invitationId
func getUserInvitationHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customer := c.Param("customer")
		invitationId := c.Param("invitationId")
		name := "customers/" + customer + "/userinvitations/" + invitationId
		invitation, err := st.GetUserInvitation(c.Request.Context(), name)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User invitation not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get user invitation")
			}
			return
		}
		writeJSON(c, http.StatusOK, invitation)
	}
}

// isInvitableUserHandler handles GET /v1/customers/:customer/userinvitations/:invitationId/isInvitable
func isInvitableUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customer := c.Param("customer")
		invitationId := c.Param("invitationId")
		name := "customers/" + customer + "/userinvitations/" + invitationId
		result, err := st.IsInvitableUser(c.Request.Context(), name)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User invitation not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to check if user is invitable")
			}
			return
		}
		writeJSON(c, http.StatusOK, gin.H{
			"isInvitableUser": result,
		})
	}
}

// sendUserInvitationHandler handles POST /v1/customers/:customer/userinvitations/:invitationId/send
func sendUserInvitationHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customer := c.Param("customer")
		invitationId := c.Param("invitationId")
		name := "customers/" + customer + "/userinvitations/" + invitationId
		if err := st.SendUserInvitation(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User invitation not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to send user invitation")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// cancelUserInvitationHandler handles POST /v1/customers/:customer/userinvitations/:invitationId/cancel
func cancelUserInvitationHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customer := c.Param("customer")
		invitationId := c.Param("invitationId")
		name := "customers/" + customer + "/userinvitations/" + invitationId
		if err := st.CancelUserInvitation(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User invitation not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to cancel user invitation")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}