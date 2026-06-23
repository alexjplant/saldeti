package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/saldeti/saldeti/internal/entra/store"
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
			if isFilterError(err) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", err.Error())
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
		if len(opts.ExpandOptions) > 0 {
			expandedUsers := make([]map[string]interface{}, 0, len(users))
			for _, u := range users {
				userMap := make(map[string]interface{})
				// Serialize user to map
				userJSON, err := json.Marshal(u)
				if err != nil {
					writeError(c, http.StatusInternalServerError, "InternalError", "Failed to process user expansion")
					return
				}
				if err := json.Unmarshal(userJSON, &userMap); err != nil {
					writeError(c, http.StatusInternalServerError, "InternalError", "Failed to process user expansion")
					return
				}

				// Add expanded properties
				for _, expand := range opts.ExpandOptions {
					switch expand.Property {
					case "manager":
						mgr, err := st.GetManager(c.Request.Context(), u.ID)
						if err == nil && mgr != nil {
							// Get full user object for the manager to support nested $select
							mgrUser, mgrErr := st.GetUser(c.Request.Context(), mgr.ID)
							if mgrErr == nil && mgrUser != nil {
								userMap["manager"] = applyNestedSelectToUser(mgrUser, expand.Select)
							} else if len(expand.Select) > 0 {
								// Fallback: use DirectoryObject with nested select
								mgrJSON, _ := json.Marshal(mgr)
								var mgrMap map[string]interface{}
								if err := json.Unmarshal(mgrJSON, &mgrMap); err != nil {
									mgrMap = map[string]interface{}{}
								}
								mgrMap = applySelect(mgrMap, expand.Select)
								mgrMap["@odata.type"] = "#microsoft.graph.user"
								userMap["manager"] = mgrMap
							} else {
								userMap["manager"] = mgr
							}
						}
						// If no manager, don't set the key at all (omit it)
					case "directReports":
						reports, _, err := st.ListDirectReports(c.Request.Context(), u.ID, model.ListOptions{Top: 999})
						if err == nil {
							userMap["directReports"] = applyNestedSelectToDirectoryObjects(nilToEmptyDirectoryObjects(reports), expand.Select)
						}
					case "memberOf":
						groups, _, err := st.ListUserMemberOf(c.Request.Context(), u.ID, model.ListOptions{Top: 999})
						if err == nil {
							userMap["memberOf"] = applyNestedSelectToDirectoryObjects(nilToEmptyDirectoryObjects(groups), expand.Select)
						}
					}
				}
				expandedUsers = append(expandedUsers, userMap)
			}
			responseValue = expandedUsers
		}

		// Apply $select if specified
		if len(opts.Select) > 0 {
			if len(opts.ExpandOptions) > 0 {
				// Items are already maps from expand handling.
				// Build a set of expanded property names so applySelect doesn't strip them.
				expandedProps := computeExpandedPropertyNames(opts.ExpandOptions)
				maps := responseValue.([]map[string]interface{})
				for i, m := range maps {
					maps[i] = applySelect(m, opts.Select, expandedProps)
				}
			} else {
				// Items are structs, serialize to maps first
				filteredItems := make([]map[string]interface{}, 0, len(users))
				for i := range users {
					itemJSON, err := json.Marshal(users[i])
					if err != nil {
						writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
						return
					}
					var itemMap map[string]interface{}
					if err := json.Unmarshal(itemJSON, &itemMap); err != nil {
						writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
						return
					}
					filteredItems = append(filteredItems, applySelect(itemMap, opts.Select))
				}
				responseValue = filteredItems
			}
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

			response.NextLink = getBaseURL(c) + nextURL.String()
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

		opts := parseListOptions(c.Request.URL.Query())

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

		response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#users/$entity", user)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		// Handle $expand
		for _, expand := range opts.ExpandOptions {
			switch expand.Property {
			case "manager":
				mgr, err := st.GetManager(c.Request.Context(), user.ID)
				if err == nil && mgr != nil {
					mgrUser, mgrErr := st.GetUser(c.Request.Context(), mgr.ID)
					if mgrErr == nil && mgrUser != nil {
						response["manager"] = applyNestedSelectToUser(mgrUser, expand.Select)
					} else if len(expand.Select) > 0 {
						mgrJSON, _ := json.Marshal(mgr)
						var mgrMap map[string]interface{}
						if err := json.Unmarshal(mgrJSON, &mgrMap); err != nil {
							mgrMap = map[string]interface{}{}
						}
						mgrMap = applySelect(mgrMap, expand.Select)
						mgrMap["@odata.type"] = "#microsoft.graph.user"
						response["manager"] = mgrMap
					} else {
						response["manager"] = mgr
					}
				}
				// If no manager, don't set the key at all (omit it)
			case "directReports":
				reports, _, err := st.ListDirectReports(c.Request.Context(), user.ID, model.ListOptions{Top: 999})
				if err == nil {
					response["directReports"] = applyNestedSelectToDirectoryObjects(nilToEmptyDirectoryObjects(reports), expand.Select)
				}
			case "memberOf":
				groups, _, err := st.ListUserMemberOf(c.Request.Context(), user.ID, model.ListOptions{Top: 999})
				if err == nil {
					response["memberOf"] = applyNestedSelectToDirectoryObjects(nilToEmptyDirectoryObjects(groups), expand.Select)
				}
			}
		}

		// Apply $select if specified
		if len(opts.Select) > 0 {
			// Preserve expanded properties so $select doesn't strip them
			expandedProps := computeExpandedPropertyNames(opts.ExpandOptions)
			response = applySelect(response, opts.Select, expandedProps)
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// createUserHandler handles POST /v1.0/users
func createUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&user); err != nil {
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

		response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#users/$entity", createdUser)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
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
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
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

		response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#users/$entity", updatedUser)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
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

// parseExpandOptions parses $expand query parameter value, handling
// nested $select inside parentheses.
// Examples:
//
//	"manager" → []ExpandOption{{Property:"manager"}}
//	"manager($select=userPrincipalName)" → []ExpandOption{{Property:"manager", Select:[]string{"userPrincipalName"}}}
//	"manager($select=userPrincipalName,displayName),directReports" → two ExpandOptions
func parseExpandOptions(expandStr string) []model.ExpandOption {
	if expandStr == "" {
		return nil
	}

	// Split on commas that are OUTSIDE parentheses
	var segments []string
	depth := 0
	current := strings.Builder{}
	for _, ch := range expandStr {
		if ch == '(' {
			depth++
			current.WriteRune(ch)
		} else if ch == ')' {
			depth--
			current.WriteRune(ch)
		} else if ch == ',' && depth == 0 {
			segments = append(segments, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		segments = append(segments, strings.TrimSpace(current.String()))
	}

	var options []model.ExpandOption
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		opt := model.ExpandOption{}
		parenIdx := strings.Index(seg, "($select=")
		if parenIdx >= 0 {
			opt.Property = seg[:parenIdx]
			inner := seg[parenIdx+9:] // len("($select=") is 9
			inner = strings.TrimSuffix(inner, ")")
			fields := strings.Split(inner, ",")
			for _, s := range fields {
				s = strings.TrimSpace(s)
				if s != "" {
					opt.Select = append(opt.Select, s)
				}
			}
		} else {
			opt.Property = seg
		}
		options = append(options, opt)
	}
	return options
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
		opts.ExpandOptions = parseExpandOptions(expandStr)
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

// getUserPhotoHandler handles GET /v1.0/users/{id}/photo
func getUserPhotoHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		writeJSON(c, http.StatusOK, gin.H{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users('" + id + "')/photo",
			"id":             "1X1",
			"height":         1,
			"width":          1,
		})
	}
}

// getUserPhotoValueHandler handles GET /v1.0/users/{id}/photo/$value
func getUserPhotoValueHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
}

// updateUserPhotoValueHandler handles PATCH /v1.0/users/{id}/photo/$value
func updateUserPhotoValueHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
}

// changePasswordHandler handles POST /v1.0/users/{id}/changePassword
func changePasswordHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	}
}

// reprocessLicenseHandler handles POST /v1.0/users/{id}/reprocessLicenseAssignment
func reprocessLicenseHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "User ID is required")
			return
		}

		var user *model.User
		var err error

		if strings.Contains(id, "@") {
			user, err = st.GetUserByUPN(c.Request.Context(), id)
		} else {
			user, err = st.GetUser(c.Request.Context(), id)
		}

		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get user")
			}
			return
		}

		response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#users/$entity", user)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// listLicenseDetailsHandler handles GET /v1.0/users/{id}/licenseDetails
func listLicenseDetailsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		writeJSON(c, http.StatusOK, gin.H{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users('" + c.Param("id") + "')/licenseDetails",
			"value":          []interface{}{},
		})
	}
}
