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

// listDomainsHandler handles GET /admin/directory/v1/customer/:customer/domains
func listDomainsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		domains, err := st.ListDomains(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list domains")
			return
		}
		if domains == nil {
			domains = []model.Domain{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#domains",
			"etag":  "\"placeholder\"",
			"domains": domains,
		})
	}
}

// getDomainHandler handles GET /admin/directory/v1/customer/:customer/domains/:domainName
func getDomainHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		domainName := c.Param("domainName")
		domain, err := st.GetDomain(c.Request.Context(), customerID, domainName)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Domain not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get domain")
			}
			return
		}
		writeJSON(c, http.StatusOK, domain)
	}
}

// addDomainHandler handles POST /admin/directory/v1/customer/:customer/domains
func addDomainHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var domain model.Domain
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&domain); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.AddDomain(c.Request.Context(), customerID, domain)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(c, http.StatusConflict, "duplicate", "Domain already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to add domain")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// deleteDomainHandler handles DELETE /admin/directory/v1/customer/:customer/domains/:domainName
func deleteDomainHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		domainName := c.Param("domainName")
		if err := st.DeleteDomain(c.Request.Context(), customerID, domainName); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Domain not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete domain")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// listDomainAliasesHandler handles GET /admin/directory/v1/customer/:customer/domainaliases
func listDomainAliasesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		aliases, err := st.ListDomainAliases(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list domain aliases")
			return
		}
		if aliases == nil {
			aliases = []model.DomainAlias{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#domainAliases",
			"etag":  "\"placeholder\"",
			"domainAliases": aliases,
		})
	}
}

// getDomainAliasHandler handles GET /admin/directory/v1/customer/:customer/domainaliases/:aliasName
func getDomainAliasHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		aliasName := c.Param("aliasName")
		alias, err := st.GetDomainAlias(c.Request.Context(), customerID, aliasName)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Domain alias not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get domain alias")
			}
			return
		}
		writeJSON(c, http.StatusOK, alias)
	}
}

// createDomainAliasHandler handles POST /admin/directory/v1/customer/:customer/domainaliases
func createDomainAliasHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var da model.DomainAlias
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&da); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateDomainAlias(c.Request.Context(), customerID, da)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(c, http.StatusConflict, "duplicate", "Domain alias already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create domain alias")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// deleteDomainAliasHandler handles DELETE /admin/directory/v1/customer/:customer/domainaliases/:aliasName
func deleteDomainAliasHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		aliasName := c.Param("aliasName")
		if err := st.DeleteDomainAlias(c.Request.Context(), customerID, aliasName); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Domain alias not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete domain alias")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}