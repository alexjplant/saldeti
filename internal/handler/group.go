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
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

// listGroupsHandler handles GET /v1.0/groups
func listGroupsHandler(st store.Store) gin.HandlerFunc {
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

		// Call store to list groups
		groups, totalCount, err := st.ListGroups(c.Request.Context(), opts)
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
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list groups")
			}
			return
		}

		// Ensure groups is not nil
		if groups == nil {
			groups = []model.Group{}
		}

		// Handle $expand - convert groups to maps with expanded properties
		var responseValue interface{} = groups
		if len(opts.Expand) > 0 {
			expandedGroups := make([]map[string]interface{}, 0, len(groups))
			for _, g := range groups {
				groupMap := make(map[string]interface{})
				groupJSON, err := json.Marshal(g)
				if err != nil {
					continue
				}
				json.Unmarshal(groupJSON, &groupMap)

				for _, prop := range opts.Expand {
					prop = strings.TrimSpace(prop)
					switch prop {
					case "members":
						members, _, err := st.ListMembers(c.Request.Context(), g.ID, model.ListOptions{Top: 999})
						if err == nil {
							if members == nil {
								members = []model.DirectoryObject{}
							}
							groupMap["members"] = members
						}
					case "owners":
						owners, _, err := st.ListOwners(c.Request.Context(), g.ID, model.ListOptions{Top: 999})
						if err == nil {
							if owners == nil {
								owners = []model.DirectoryObject{}
							}
							groupMap["owners"] = owners
						}
					case "memberOf":
						memberOf, _, err := st.ListGroupMemberOf(c.Request.Context(), g.ID, model.ListOptions{Top: 999})
						if err == nil {
							if memberOf == nil {
								memberOf = []model.DirectoryObject{}
							}
							groupMap["memberOf"] = memberOf
						}
					}
				}
				expandedGroups = append(expandedGroups, groupMap)
			}
			responseValue = expandedGroups
		}

		// Apply $select if specified
		if len(opts.Select) > 0 {
			if len(opts.Expand) > 0 {
				// Items are already maps from expand handling
				maps := responseValue.([]map[string]interface{})
				for i, m := range maps {
					maps[i] = applySelect(m, opts.Select)
				}
			} else {
				// Items are structs, serialize to maps first
				filteredItems := make([]map[string]interface{}, 0, len(groups))
				for i := range groups {
					itemJSON, err := json.Marshal(groups[i])
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
			Context: "https://graph.microsoft.com/v1.0/$metadata#groups",
			Value:   responseValue,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(groups) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// getGroupHandler handles GET /v1.0/groups/{id}
func getGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		group, err := st.GetGroup(c.Request.Context(), id)
		if err != nil {
			// Check if error is ErrGroupNotFound using errors.Is
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get group")
			}
			return
		}

		// Build response with proper context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#groups/$entity",
			"@odata.type":    "#microsoft.graph.group",
		}

		// Merge group fields into response
		groupJSON, err := json.Marshal(group)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var groupMap map[string]interface{}
		if err := json.Unmarshal(groupJSON, &groupMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range groupMap {
			response[k] = v
		}

		// Handle $expand
		if expandStr := c.Request.URL.Query().Get("$expand"); expandStr != "" {
			expandProps := strings.Split(expandStr, ",")
			for _, prop := range expandProps {
				prop = strings.TrimSpace(prop)
				switch prop {
				case "members":
					members, _, err := st.ListMembers(c.Request.Context(), id, model.ListOptions{Top: 999})
					if err == nil {
						if members == nil {
							members = []model.DirectoryObject{}
						}
						response["members"] = members
					}
				case "owners":
					owners, _, err := st.ListOwners(c.Request.Context(), id, model.ListOptions{Top: 999})
					if err == nil {
						if owners == nil {
							owners = []model.DirectoryObject{}
						}
						response["owners"] = owners
					}
				case "memberOf":
					memberOf, _, err := st.ListGroupMemberOf(c.Request.Context(), id, model.ListOptions{Top: 999})
					if err == nil {
						if memberOf == nil {
							memberOf = []model.DirectoryObject{}
						}
						response["memberOf"] = memberOf
					}
				}
			}
		}

		// Apply $select if specified
		opts := parseListOptions(c.Request.URL.Query())
		if len(opts.Select) > 0 {
			response = applySelect(response, opts.Select)
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// createGroupHandler handles POST /v1.0/groups
func createGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, read the raw request body to handle members@odata.bind and owners@odata.bind
		var requestBody map[string]interface{}
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Failed to read request body")
			return
		}
		if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		var group model.Group
		if err := json.Unmarshal(bodyBytes, &group); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Handle members@odata.bind
		if membersBind, ok := requestBody["members@odata.bind"]; ok {
			if membersArray, ok := membersBind.([]interface{}); ok {
				for _, memberRef := range membersArray {
					if memberStr, ok := memberRef.(string); ok {
						group.Members = append(group.Members, model.DirectoryObjectRef{
							ODataID: memberStr,
						})
					}
				}
			} else if memberStr, ok := membersBind.(string); ok {
				group.Members = append(group.Members, model.DirectoryObjectRef{
					ODataID: memberStr,
				})
			}
		}

		// Handle owners@odata.bind
		if ownersBind, ok := requestBody["owners@odata.bind"]; ok {
			if ownersArray, ok := ownersBind.([]interface{}); ok {
				for _, ownerRef := range ownersArray {
					if ownerStr, ok := ownerRef.(string); ok {
						group.Owners = append(group.Owners, model.DirectoryObjectRef{
							ODataID: ownerStr,
						})
					}
				}
			} else if ownerStr, ok := ownersBind.(string); ok {
				group.Owners = append(group.Owners, model.DirectoryObjectRef{
					ODataID: ownerStr,
				})
			}
		}

		// Validate required fields
		if group.DisplayName == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "displayName is required")
			return
		}

		// Set OData type if not provided
		if group.ODataType == "" {
			group.ODataType = "#microsoft.graph.group"
		}

		createdGroup, err := st.CreateGroup(c.Request.Context(), group)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateGroup) {
				writeError(c, http.StatusConflict, "Conflict", "Group already exists")
			} else if errors.Is(err, store.ErrDisplayNameRequired) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "displayName is required")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to create group")
			}
			return
		}

		// Build response with OData context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#groups/$entity",
		}

		// Merge group fields into response
		groupJSON, err := json.Marshal(createdGroup)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var groupMap map[string]interface{}
		if err := json.Unmarshal(groupJSON, &groupMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range groupMap {
			response[k] = v
		}

		c.Header("Location", "/v1.0/groups/"+createdGroup.ID)
		writeJSON(c, http.StatusCreated, response)
	}
}

// updateGroupHandler handles PATCH /v1.0/groups/{id}
func updateGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		// Decode patch as map
		var patch map[string]interface{}
		if err := json.NewDecoder(c.Request.Body).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		group, err := st.UpdateGroup(c.Request.Context(), id, patch)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to update group")
			}
			return
		}

		// Build response with proper context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#groups/$entity",
			"@odata.type":    "#microsoft.graph.group",
		}

		// Merge group fields into response
		groupJSON, err := json.Marshal(group)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var groupMap map[string]interface{}
		if err := json.Unmarshal(groupJSON, &groupMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range groupMap {
			response[k] = v
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// deleteGroupHandler handles DELETE /v1.0/groups/{id}
func deleteGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		err := st.DeleteGroup(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to delete group")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listMembersHandler handles GET /v1.0/groups/{id}/members
func listMembersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		members, totalCount, err := st.ListMembers(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list members")
			}
			return
		}

		// Ensure members is not nil
		if members == nil {
			members = []model.DirectoryObject{}
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   members,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(members) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// addMemberHandler handles POST /v1.0/groups/{id}/members/$ref
func addMemberHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		var refBody map[string]string
		if err := json.NewDecoder(c.Request.Body).Decode(&refBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		odataID, ok := refBody["@odata.id"]
		if !ok {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "@odata.id is required")
			return
		}

		// Extract object ID from @odata.id URL
		// Format: "https://graph.microsoft.com/v1.0/directoryObjects/{objectId}"
		parts := strings.Split(odataID, "/")
		if len(parts) == 0 {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid @odata.id format")
			return
		}
		objectID := parts[len(parts)-1]

		// Determine object type by looking it up in the store
		objectType, err := st.ResolveObjectType(c.Request.Context(), objectID)
		if err != nil {
			writeError(c, http.StatusNotFound, "ResourceNotFound", "Object not found")
			return
		}

		err = st.AddMember(c.Request.Context(), id, objectID, objectType)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "User not found")
			} else if errors.Is(err, store.ErrAlreadyMember) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "Object is already a member of the group")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to add member")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// removeMemberHandler handles DELETE /v1.0/groups/{id}/members/{memberId}/$ref
func removeMemberHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		memberID := c.Param("memberId")
		if memberID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Member ID is required")
			return
		}

		err := st.RemoveMember(c.Request.Context(), id, memberID)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else if errors.Is(err, store.ErrNotMember) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Object is not a member of the group")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to remove member")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listTransitiveMembersHandler handles GET /v1.0/groups/{id}/transitiveMembers
func listTransitiveMembersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		members, totalCount, err := st.ListTransitiveMembers(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list transitive members")
			}
			return
		}

		// Ensure members is not nil
		if members == nil {
			members = []model.DirectoryObject{}
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   members,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(members) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// listOwnersHandler handles GET /v1.0/groups/{id}/owners
func listOwnersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		owners, totalCount, err := st.ListOwners(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list owners")
			}
			return
		}

		// Ensure owners is not nil
		if owners == nil {
			owners = []model.DirectoryObject{}
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   owners,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(owners) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// addOwnerHandler handles POST /v1.0/groups/{id}/owners/$ref
func addOwnerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		var refBody map[string]string
		if err := json.NewDecoder(c.Request.Body).Decode(&refBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		odataID, ok := refBody["@odata.id"]
		if !ok {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "@odata.id is required")
			return
		}

		// Extract object ID from @odata.id URL
		parts := strings.Split(odataID, "/")
		if len(parts) == 0 {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid @odata.id format")
			return
		}
		objectID := parts[len(parts)-1]

		// Determine object type by looking it up in the store
		objectType, err := st.ResolveObjectType(c.Request.Context(), objectID)
		if err != nil {
			writeError(c, http.StatusNotFound, "ResourceNotFound", "Object not found")
			return
		}

		err = st.AddOwner(c.Request.Context(), id, objectID, objectType)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "User not found")
			} else if errors.Is(err, store.ErrAlreadyOwner) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "Object is already an owner of the group")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to add owner")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// removeOwnerHandler handles DELETE /v1.0/groups/{id}/owners/{ownerId}/$ref
func removeOwnerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		ownerID := c.Param("ownerId")
		if ownerID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Owner ID is required")
			return
		}

		err := st.RemoveOwner(c.Request.Context(), id, ownerID)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else if errors.Is(err, store.ErrNotOwner) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Object is not an owner of the group")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to remove owner")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listGroupMemberOfHandler handles GET /v1.0/groups/{id}/memberOf
func listGroupMemberOfHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		memberOf, totalCount, err := st.ListGroupMemberOf(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list memberOf")
			}
			return
		}

		// Ensure memberOf is not nil
		if memberOf == nil {
			memberOf = []model.DirectoryObject{}
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   memberOf,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(memberOf) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// listGroupTransitiveMemberOfHandler handles GET /v1.0/groups/{id}/transitiveMemberOf
func listGroupTransitiveMemberOfHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		memberOf, totalCount, err := st.ListGroupTransitiveMemberOf(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list transitive memberOf")
			}
			return
		}

		// Ensure memberOf is not nil
		if memberOf == nil {
			memberOf = []model.DirectoryObject{}
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   memberOf,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(memberOf) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// checkMemberGroupsHandler handles POST /v1.0/groups/{id}/checkMemberGroups
func checkMemberGroupsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		var requestBody struct {
			GroupIDs []string `json:"groupIds"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		if len(requestBody.GroupIDs) == 0 {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "groupIds is required")
			return
		}

		matchingGroups, err := st.CheckMemberGroups(c.Request.Context(), id, requestBody.GroupIDs)
		if err != nil {
			if errors.Is(err, store.ErrObjectNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Object not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to check member groups")
			}
			return
		}

		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#Collection(Edm.String)",
			"value":          matchingGroups,
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// getMemberGroupsHandler handles POST /v1.0/groups/{id}/getMemberGroups
func getMemberGroupsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		var requestBody struct {
			SecurityEnabledOnly bool `json:"securityEnabledOnly"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		memberGroups, err := st.GetMemberGroups(c.Request.Context(), id, requestBody.SecurityEnabledOnly)
		if err != nil {
			if errors.Is(err, store.ErrObjectNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Object not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get member groups")
			}
			return
		}

		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#Collection(Edm.String)",
			"value":          memberGroups,
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// listMembersByTypeHandler handles GET /v1.0/groups/{id}/members/microsoft.graph.{type}
func listMembersByTypeHandler(st store.Store, objectType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		members, totalCount, err := st.ListMembers(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list members")
			}
			return
		}

		// Filter by object type
		filteredMembers := make([]model.DirectoryObject, 0)
		for _, member := range members {
			expectedODataType := "#microsoft.graph." + objectType
			if member.ODataType == expectedODataType {
				filteredMembers = append(filteredMembers, member)
			}
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   filteredMembers,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// listOwnersByTypeHandler handles GET /v1.0/groups/{id}/owners/microsoft.graph.{type}
func listOwnersByTypeHandler(st store.Store, objectType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		owners, totalCount, err := st.ListOwners(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list owners")
			}
			return
		}

		// Filter by object type
		filteredOwners := make([]model.DirectoryObject, 0)
		for _, owner := range owners {
			expectedODataType := "#microsoft.graph." + objectType
			if owner.ODataType == expectedODataType {
				filteredOwners = append(filteredOwners, owner)
			}
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   filteredOwners,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// getMemberObjectsHandler handles POST /v1.0/groups/{id}/getMemberObjects
func getMemberObjectsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		var requestBody struct {
			SecurityEnabledOnly bool `json:"securityEnabledOnly"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		memberObjects, err := st.GetMemberGroups(c.Request.Context(), id, requestBody.SecurityEnabledOnly)
		if err != nil {
			if errors.Is(err, store.ErrObjectNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Object not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get member objects")
			}
			return
		}

		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#Collection(Edm.String)",
			"value":          memberObjects,
		}

		writeJSON(c, http.StatusOK, response)
	}
}

