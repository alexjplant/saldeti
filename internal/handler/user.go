package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

// listUsersHandler handles GET /v1.0/users
func listUsersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		// Validate $top parameter
		if topStr := c.Request.URL.Query().Get("$top"); topStr != "" {
			if top, err := strconv.Atoi(topStr); err == nil && top > 0 {
				if top > 999 {
					writeError(c, http.StatusBadRequest, "Request_BadRequest",
						fmt.Sprintf("$top value %d exceeds maximum of 999.", top))
					return
				}
			}
		}

		// Call store to list users
		users, totalCount, err := st.ListUsers(c.Request.Context(), opts)
		if err != nil {
			// Check if error is a filter parsing error
			errStr := err.Error()
			if strings.Contains(errStr, "unable to parse filter expression") ||
			   strings.Contains(errStr, "cannot compare values") ||
			   strings.Contains(errStr, "operator not supported") ||
			   strings.Contains(errStr, "function value must be string") ||
			   strings.Contains(errStr, "unknown function") ||
			   strings.Contains(errStr, "invalid filter node") {
				writeError(c, http.StatusBadRequest, "InvalidRequest", errStr)
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list users")
			}
			return
		}

		// Ensure users is not nil
		if users == nil {
			users = []model.User{}
		}

		// Handle $expand - convert users to maps with expanded properties
		var responseValue interface{} = users
		if len(opts.Expand) > 0 {
			expandedUsers := make([]map[string]interface{}, 0, len(users))
			for _, u := range users {
				userMap := make(map[string]interface{})
				// Serialize user to map
				userJSON, err := json.Marshal(u)
				if err != nil {
					continue
				}
				json.Unmarshal(userJSON, &userMap)

				// Add expanded properties
				for _, prop := range opts.Expand {
					prop = strings.TrimSpace(prop)
					switch prop {
					case "manager":
						mgr, err := st.GetManager(c.Request.Context(), u.ID)
						if err == nil && mgr != nil {
							userMap["manager"] = mgr
						} else {
							userMap["manager"] = nil
						}
					case "directReports":
						reports, _, err := st.ListDirectReports(c.Request.Context(), u.ID, model.ListOptions{Top: 999})
						if err == nil {
							if reports == nil {
								reports = []model.DirectoryObject{}
							}
							userMap["directReports"] = reports
						}
					case "memberOf":
						groups, _, err := st.ListUserMemberOf(c.Request.Context(), u.ID, model.ListOptions{Top: 999})
						if err == nil {
							if groups == nil {
								groups = []model.DirectoryObject{}
							}
							userMap["memberOf"] = groups
						}
					}
				}
				expandedUsers = append(expandedUsers, userMap)
			}
			responseValue = expandedUsers
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#users",
			Value:   responseValue,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(users) == opts.Top && opts.Skip+opts.Top < totalCount {
			nextSkip := opts.Skip + opts.Top
			nextURL := url.URL{
				Path:     c.Request.URL.Path,
				RawQuery: buildNextLinkQuery(c.Request.URL.Query(), nextSkip),
			}
			
			// Use request host and scheme
			host := c.Request.Host
			if forwarded := c.GetHeader("X-Forwarded-Host"); forwarded != "" {
				host = forwarded
			}
			scheme := "http"
			if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
				scheme = "https"
			}
			
			response.NextLink = scheme + "://" + host + nextURL.String()
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// getUserHandler handles GET /v1.0/users/{id}
func getUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "User ID is required")
			return
		}

		var user *model.User
		var err error

		// Check if ID is a UPN (contains @)
		if strings.Contains(id, "@") {
			user, err = st.GetUserByUPN(c.Request.Context(), id)
		} else {
			user, err = st.GetUser(c.Request.Context(), id)
		}

		if err != nil {
			// Check if error is ErrUserNotFound using errors.Is
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get user")
			}
			return
		}

		// Build response with OData context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users/$entity",
		}

		// Merge user fields into response
		userJSON, err := json.Marshal(user)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var userMap map[string]interface{}
		if err := json.Unmarshal(userJSON, &userMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range userMap {
			response[k] = v
		}

		// Handle $expand
		if expandStr := c.Request.URL.Query().Get("$expand"); expandStr != "" {
			expandProps := strings.Split(expandStr, ",")
			for _, prop := range expandProps {
				prop = strings.TrimSpace(prop)
				switch prop {
				case "manager":
					mgr, err := st.GetManager(c.Request.Context(), id)
					if err == nil && mgr != nil {
						response["manager"] = mgr
					} else {
						response["manager"] = nil
					}
				case "directReports":
					reports, _, err := st.ListDirectReports(c.Request.Context(), id, model.ListOptions{Top: 999})
					if err == nil {
						if reports == nil {
							reports = []model.DirectoryObject{}
						}
						response["directReports"] = reports
					}
				case "memberOf":
					groups, _, err := st.ListUserMemberOf(c.Request.Context(), id, model.ListOptions{Top: 999})
					if err == nil {
						if groups == nil {
							groups = []model.DirectoryObject{}
						}
						response["memberOf"] = groups
					}
				}
			}
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// createUserHandler handles POST /v1.0/users
func createUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		if err := json.NewDecoder(c.Request.Body).Decode(&user); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Validate required fields
		if user.DisplayName == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "displayName is required")
			return
		}
		if user.UserPrincipalName == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "userPrincipalName is required")
			return
		}

		// Set OData type if not provided
		if user.ODataType == "" {
			user.ODataType = "#microsoft.graph.user"
		}

		// Generate ID if not provided
		if user.ID == "" {
			user.ID = uuid.New().String()
		}

		// Create user
		createdUser, err := st.CreateUser(c.Request.Context(), user)
		if err != nil {
			// Check if error is ErrDuplicateUPN using errors.Is
			if errors.Is(err, store.ErrDuplicateUPN) {
				writeError(c, http.StatusConflict, "Conflict", "User with this userPrincipalName already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to create user")
			}
			return
		}

		// Set Location header
		c.Header("Location", "/v1.0/users/"+createdUser.ID)

		// Build response with OData context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users/$entity",
		}

		// Merge user fields into response
		userJSON, err := json.Marshal(createdUser)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var userMap map[string]interface{}
		if err := json.Unmarshal(userJSON, &userMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range userMap {
			response[k] = v
		}

		writeJSON(c, http.StatusCreated, response)
	}
}

// updateUserHandler handles PATCH /v1.0/users/{id}
func updateUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "User ID is required")
			return
		}

		// Decode patch as map
		var patch map[string]interface{}
		if err := json.NewDecoder(c.Request.Body).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Update user
		updatedUser, err := st.UpdateUser(c.Request.Context(), id, patch)
		if err != nil {
			// Check if error is ErrUserNotFound using errors.Is
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to update user")
			}
			return
		}

		// Build response with OData context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users/$entity",
		}

		// Merge user fields into response
		userJSON, err := json.Marshal(updatedUser)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var userMap map[string]interface{}
		if err := json.Unmarshal(userJSON, &userMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range userMap {
			response[k] = v
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// deleteUserHandler handles DELETE /v1.0/users/{id}
func deleteUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "User ID is required")
			return
		}

		// Delete user
		err := st.DeleteUser(c.Request.Context(), id)
		if err != nil {
			// Check if error is ErrUserNotFound using errors.Is
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to delete user")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// parseListOptions parses OData query parameters
func parseListOptions(query url.Values) model.ListOptions {
	opts := model.ListOptions{
		Top:  100, // Default page size
		Skip: 0,
	}

	// Parse $filter
	if filter := query.Get("$filter"); filter != "" {
		opts.Filter = filter
	}

	// Parse $select
	if selectStr := query.Get("$select"); selectStr != "" {
		opts.Select = strings.Split(selectStr, ",")
	}

	// Parse $top
	if topStr := query.Get("$top"); topStr != "" {
		if top, err := strconv.Atoi(topStr); err == nil && top > 0 {
			opts.Top = top
		}
	}

	// Parse $orderby
	if orderBy := query.Get("$orderby"); orderBy != "" {
		opts.OrderBy = orderBy
	}

	// Parse $count
	if countStr := query.Get("$count"); countStr != "" {
		opts.Count = strings.ToLower(countStr) == "true"
	}

	// Parse $search
	if search := query.Get("$search"); search != "" {
		opts.Search = search
	}

	// Parse $skip
	if skipStr := query.Get("$skip"); skipStr != "" {
		if skip, err := strconv.Atoi(skipStr); err == nil && skip >= 0 {
			opts.Skip = skip
		}
	}

	// Parse $expand
	if expandStr := query.Get("$expand"); expandStr != "" {
		opts.Expand = strings.Split(expandStr, ",")
	}

	return opts
}

// buildNextLinkQuery builds query string for nextLink
func buildNextLinkQuery(originalQuery url.Values, nextSkip int) string {
	q := url.Values{}
	
	// Copy all original parameters
	for key, values := range originalQuery {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	
	// Update $skip parameter
	q.Set("$skip", strconv.Itoa(nextSkip))
	
	return q.Encode()
}
