package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
	"github.com/saldeti/saldeti/internal/google/store"
)

// listSubscriptionsHandler handles GET /v1/subscriptions
func listSubscriptionsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		subs, err := st.ListSubscriptions(c.Request.Context())
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list subscriptions")
			return
		}
		if subs == nil {
			subs = []model.Subscription{}
		}
		writeJSON(c, http.StatusOK, gin.H{"subscriptions": subs})
	}
}

// getSubscriptionHandler handles GET /v1/subscriptions/:subscriptionName
func getSubscriptionHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		subscriptionName := c.Param("subscriptionName")
		name := "subscriptions/" + subscriptionName
		sub, err := st.GetSubscription(c.Request.Context(), name)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Subscription not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get subscription")
			}
			return
		}
		writeJSON(c, http.StatusOK, sub)
	}
}

// createSubscriptionHandler handles POST /v1/subscriptions
func createSubscriptionHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sub model.Subscription
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&sub); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateSubscription(c.Request.Context(), sub)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to create subscription")
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// updateSubscriptionHandler handles PATCH /v1/subscriptions/:subscriptionName
func updateSubscriptionHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		subscriptionName := c.Param("subscriptionName")
		name := "subscriptions/" + subscriptionName
		var sub model.Subscription
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&sub); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateSubscription(c.Request.Context(), name, sub)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Subscription not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update subscription")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// deleteSubscriptionHandler handles DELETE /v1/subscriptions/:subscriptionName
func deleteSubscriptionHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		subscriptionName := c.Param("subscriptionName")
		name := "subscriptions/" + subscriptionName
		if err := st.DeleteSubscription(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Subscription not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete subscription")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// reactivateSubscriptionHandler handles POST /v1/subscriptions/:subscriptionName/reactivate
func reactivateSubscriptionHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		subscriptionName := c.Param("subscriptionName")
		name := "subscriptions/" + subscriptionName
		if err := st.ReactivateSubscription(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Subscription not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to reactivate subscription")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}