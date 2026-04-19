package store

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/saldeti/saldeti/internal/model"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrClientNotFound   = errors.New("client not found")
	ErrDuplicateUPN     = errors.New("user with same UPN already exists")
	ErrDuplicateClient  = errors.New("client already registered")
	ErrGroupNotFound    = errors.New("group not found")
	ErrDuplicateGroup   = errors.New("group already exists")
	ErrAlreadyMember    = errors.New("object is already a member of the group")
	ErrNotMember        = errors.New("object is not a member of the group")
	ErrAlreadyOwner     = errors.New("object is already an owner of the group")
	ErrNotOwner         = errors.New("object is not an owner of the group")
	ErrObjectNotFound   = errors.New("object not found")
	ErrInvalidObjType   = errors.New("invalid object type")
	ErrDisplayNameRequired = errors.New("displayName is required")
	ErrManagerNotFound   = errors.New("manager not found")
)

type memoryStore struct {
	mu                 sync.RWMutex
	users              map[string]model.User
	clients            map[string]clientEntry
	groups             map[string]model.Group
	servicePrincipals  map[string]model.ServicePrincipal
	members            map[string]map[string]string // groupID → {objectID → objectType}
	owners             map[string]map[string]string // groupID → {objectID → objectType}
	managers           map[string]string            // userID → managerID
	deletedUsers       map[string]time.Time         // ID → deletedAt
	deletedGroups      map[string]time.Time         // ID → deletedAt
}

type clientEntry struct {
	clientID     string
	clientSecret string
	tenantID     string
}

func NewMemoryStore() Store {
	return &memoryStore{
		users:             make(map[string]model.User),
		clients:           make(map[string]clientEntry),
		groups:            make(map[string]model.Group),
		servicePrincipals: make(map[string]model.ServicePrincipal),
		members:           make(map[string]map[string]string),
		owners:            make(map[string]map[string]string),
		managers:          make(map[string]string),
		deletedUsers:      make(map[string]time.Time),
		deletedGroups:     make(map[string]time.Time),
	}
}

func (s *memoryStore) GetUser(ctx context.Context, id string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (s *memoryStore) GetUserByUPN(ctx context.Context, upn string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.UserPrincipalName == upn {
			return &user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *memoryStore) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user with same UPN already exists
	for _, existingUser := range s.users {
		if existingUser.UserPrincipalName == user.UserPrincipalName {
			return model.User{}, ErrDuplicateUPN
		}
	}

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Initialize BusinessPhones to empty slice if nil
	if user.BusinessPhones == nil {
		user.BusinessPhones = []string{}
	}

	// Set creation time if not provided
	if user.CreatedDateTime == nil {
		now := time.Now()
		user.CreatedDateTime = &now
	}

	// Set modification time
	user.ModifiedAt = time.Now()

	s.users[user.ID] = user
	return user, nil
}

func (s *memoryStore) GetClient(ctx context.Context, clientID string) (string, string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.clients[clientID]
	if !exists {
		return "", "", "", ErrClientNotFound
	}
	return entry.clientID, entry.clientSecret, entry.tenantID, nil
}

func (s *memoryStore) RegisterClient(ctx context.Context, clientID, clientSecret, tenantID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[clientID]; exists {
		return ErrDuplicateClient
	}

	s.clients[clientID] = clientEntry{
		clientID:     clientID,
		clientSecret: clientSecret,
		tenantID:     tenantID,
	}

	// Create a service principal for this client
	spID := "sp-" + clientID
	s.servicePrincipals[spID] = model.ServicePrincipal{
		ID:          spID,
		AppID:       clientID,
		DisplayName: "Service Principal for " + clientID,
		ODataType:   "#microsoft.graph.servicePrincipal",
	}

	return nil
}

func (s *memoryStore) ListClients(ctx context.Context) ([]Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]Client, 0, len(s.clients))
	for _, entry := range s.clients {
		clients = append(clients, Client{
			ClientID:     entry.clientID,
			ClientSecret: entry.clientSecret,
			TenantID:     entry.tenantID,
		})
	}
	return clients, nil
}

func (s *memoryStore) ListUsers(ctx context.Context, opts model.ListOptions) ([]model.User, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert map to slice
	users := make([]model.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	// Apply OData filtering, sorting, and pagination
	filteredUsers, totalCount, err := ApplyOData(users, opts)
	if err != nil {
		return nil, 0, err
	}

	return filteredUsers, totalCount, nil
}

func (s *memoryStore) UpdateUser(ctx context.Context, id string, patch map[string]interface{}) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Apply patch using reflection
	updatedUser, err := applyPatch(user, patch)
	if err != nil {
		return nil, err
	}

	// Set modification time
	updatedUser.ModifiedAt = time.Now()

	s.users[id] = updatedUser
	return &updatedUser, nil
}

func (s *memoryStore) DeleteUser(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return ErrUserNotFound
	}

	// Record deletion before deleting
	s.deletedUsers[id] = time.Now()

	delete(s.users, id)

	// Clean up membership references (remove user from all groups' membership lists)
	for _, memberMap := range s.members {
		delete(memberMap, id)
	}

	// Clean up ownership references (remove user from all groups' ownership lists)
	for _, ownerMap := range s.owners {
		delete(ownerMap, id)
	}

	// Clean up manager references (remove user's manager entry)
	delete(s.managers, id)

	// Remove user as someone else's manager
	for userID, managerID := range s.managers {
		if managerID == id {
			delete(s.managers, userID)
		}
	}

	return nil
}

// applyPatch applies a patch map to a struct using generics
func applyPatch[T any](obj T, patch map[string]interface{}) (T, error) {
	// Use reflection to update fields
	v := reflect.ValueOf(&obj).Elem()
	t := reflect.TypeOf(obj)

	for jsonName, patchValue := range patch {
		// Find field with matching JSON tag
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			tagName := strings.Split(jsonTag, ",")[0]
			if tagName == jsonName {
				fieldValue := v.Field(i)

				// Handle pointer fields
				if fieldValue.Kind() == reflect.Ptr {
					if patchValue == nil {
						// Set pointer to nil
						fieldValue.Set(reflect.Zero(fieldValue.Type()))
					} else {
						// Create new pointer and set value
						ptrType := fieldValue.Type()
						elemType := ptrType.Elem()
						newPtr := reflect.New(elemType)

						// Convert patch value to appropriate type
						convertedValue, err := convertValue(patchValue, elemType)
						if err != nil {
							return obj, err
						}

						newPtr.Elem().Set(convertedValue)
						fieldValue.Set(newPtr)
					}
				} else if fieldValue.Kind() == reflect.Slice {
					// Handle slice fields (GroupTypes, ProxyAddresses)
					if sliceValue, ok := patchValue.([]interface{}); ok {
						// Convert []interface{} to appropriate slice type
						sliceType := fieldValue.Type()
						elemType := sliceType.Elem()
						newSlice := reflect.MakeSlice(sliceType, len(sliceValue), len(sliceValue))

						for j, elem := range sliceValue {
							convertedElem, err := convertValue(elem, elemType)
							if err != nil {
								return obj, err
							}
							newSlice.Index(j).Set(convertedElem)
						}
						fieldValue.Set(newSlice)
					}
				} else {
					// Non-pointer, non-slice field
					convertedValue, err := convertValue(patchValue, fieldValue.Type())
					if err != nil {
						return obj, err
					}
					fieldValue.Set(convertedValue)
				}
				break
			}
		}
	}

	return obj, nil
}

// convertValue converts a patch value to the target type
func convertValue(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	sourceValue := reflect.ValueOf(value)
	sourceType := sourceValue.Type()

	// If types match, return directly
	if sourceType.AssignableTo(targetType) {
		return sourceValue, nil
	}

	// Handle string to time.Time conversion
	if targetType == reflect.TypeOf(time.Time{}) && sourceType.Kind() == reflect.String {
		strValue := value.(string)
		t, err := time.Parse(time.RFC3339, strValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid time format: %v", err)
		}
		return reflect.ValueOf(t), nil
	}

	// Handle string to *time.Time conversion
	if targetType == reflect.TypeOf(&time.Time{}) && sourceType.Kind() == reflect.String {
		strValue := value.(string)
		t, err := time.Parse(time.RFC3339, strValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid time format: %v", err)
		}
		return reflect.ValueOf(&t), nil
	}

	// Handle string to bool conversion
	if targetType.Kind() == reflect.Bool && sourceType.Kind() == reflect.String {
		strValue := value.(string)
		boolValue, err := strconv.ParseBool(strValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid boolean value: %v", err)
		}
		return reflect.ValueOf(boolValue), nil
	}

	// Handle string to *bool conversion
	if targetType == reflect.TypeOf((*bool)(nil)) && sourceType.Kind() == reflect.String {
		strValue := value.(string)
		boolValue, err := strconv.ParseBool(strValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid boolean value: %v", err)
		}
		return reflect.ValueOf(&boolValue), nil
	}

	// Handle numeric conversions
	if isNumericType(targetType) && isNumericType(sourceType) {
		return convertNumericValue(value, targetType)
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %v to %v", sourceType, targetType)
}

// isNumericType checks if a type is numeric
func isNumericType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// convertNumericValue converts between numeric types
func convertNumericValue(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	floatValue := 0.0
	
	switch v := value.(type) {
	case int:
		floatValue = float64(v)
	case int64:
		floatValue = float64(v)
	case float32:
		floatValue = float64(v)
	case float64:
		floatValue = v
	default:
		return reflect.Value{}, fmt.Errorf("unsupported numeric type: %T", value)
	}

	switch targetType.Kind() {
	case reflect.Int:
		return reflect.ValueOf(int(floatValue)), nil
	case reflect.Int64:
		return reflect.ValueOf(int64(floatValue)), nil
	case reflect.Float32:
		return reflect.ValueOf(float32(floatValue)), nil
	case reflect.Float64:
		return reflect.ValueOf(floatValue), nil
	default:
		return reflect.Value{}, fmt.Errorf("unsupported target numeric type: %v", targetType)
	}
}

// Group methods
func (s *memoryStore) ListGroups(ctx context.Context, opts model.ListOptions) ([]model.Group, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert map to slice
	groups := make([]model.Group, 0, len(s.groups))
	for _, group := range s.groups {
		groups = append(groups, group)
	}

	// Apply OData filtering, sorting, and pagination
	filteredGroups, totalCount, err := ApplyOData(groups, opts)
	if err != nil {
		return nil, 0, err
	}

	return filteredGroups, totalCount, nil
}

func (s *memoryStore) GetGroup(ctx context.Context, id string) (*model.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	group, exists := s.groups[id]
	if !exists {
		return nil, ErrGroupNotFound
	}
	return &group, nil
}

func (s *memoryStore) CreateGroup(ctx context.Context, group model.Group) (model.Group, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate displayName is required
	if group.DisplayName == "" {
		return model.Group{}, ErrDisplayNameRequired
	}

	// Generate ID if not provided
	if group.ID == "" {
		group.ID = uuid.New().String()
	}

	// Check if group with same ID already exists
	if _, exists := s.groups[group.ID]; exists {
		return model.Group{}, ErrDuplicateGroup
	}

	// Set creation time if not provided
	if group.CreatedDateTime == nil {
		now := time.Now()
		group.CreatedDateTime = &now
	}

	// Generate mailNickname if not provided
	if group.MailNickname == "" {
		group.MailNickname = strings.ToLower(strings.ReplaceAll(group.DisplayName, " ", ""))
	}

	// Set default values
	if group.SecurityEnabled == nil {
		securityEnabled := true
		group.SecurityEnabled = &securityEnabled
	}
	if group.MailEnabled == nil {
		mailEnabled := false
		group.MailEnabled = &mailEnabled
	}
	if group.Visibility == "" {
		group.Visibility = "Public"
	}

	// Ensure ODataType is set for proper SDK deserialization
	if group.ODataType == "" {
		group.ODataType = "#microsoft.graph.group"
	}

	// Set modification time
	group.ModifiedAt = time.Now()

	// Store the group
	s.groups[group.ID] = group

	// Initialize membership maps
	s.members[group.ID] = make(map[string]string)
	s.owners[group.ID] = make(map[string]string)

	// Add members if provided in the request
	for _, memberRef := range group.Members {
		// Extract object ID from @odata.id URL
		// Format: "https://graph.microsoft.com/v1.0/directoryObjects/{objectId}"
		parts := strings.Split(memberRef.ODataID, "/")
		if len(parts) > 0 {
			objectID := parts[len(parts)-1]
			// Determine object type by checking if it exists in users or groups
			objectType, err := s.resolveObjectTypeUnsafe(objectID)
			if err != nil {
				// Skip if object not found (this matches the behavior of AddMember/AddOwner)
				continue
			}
			s.members[group.ID][objectID] = objectType
		}
	}

	// Add owners if provided in the request
	for _, ownerRef := range group.Owners {
		parts := strings.Split(ownerRef.ODataID, "/")
		if len(parts) > 0 {
			objectID := parts[len(parts)-1]
			objectType, err := s.resolveObjectTypeUnsafe(objectID)
			if err != nil {
				// Skip if object not found (this matches the behavior of AddMember/AddOwner)
				continue
			}
			s.owners[group.ID][objectID] = objectType
		}
	}

	return group, nil
}

func (s *memoryStore) UpdateGroup(ctx context.Context, id string, patch map[string]interface{}) (*model.Group, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, exists := s.groups[id]
	if !exists {
		return nil, ErrGroupNotFound
	}

	// Apply patch using reflection
	updatedGroup, err := applyPatch(group, patch)
	if err != nil {
		return nil, err
	}

	// Set modification time
	updatedGroup.ModifiedAt = time.Now()

	s.groups[id] = updatedGroup
	return &updatedGroup, nil
}

func (s *memoryStore) DeleteGroup(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.groups[id]; !exists {
		return ErrGroupNotFound
	}

	// Record deletion before deleting
	s.deletedGroups[id] = time.Now()

	// Delete the group
	delete(s.groups, id)

	// Clean up membership references
	delete(s.members, id)
	delete(s.owners, id)

	// Remove this group from all other groups' membership
	for _, memberMap := range s.members {
		delete(memberMap, id)
	}

	// Remove this group from all other groups' ownership
	for _, ownerMap := range s.owners {
		delete(ownerMap, id)
	}

	return nil
}

// Membership methods
func (s *memoryStore) AddMember(ctx context.Context, groupID, objectID, objectType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if group exists
	if _, exists := s.groups[groupID]; !exists {
		return ErrGroupNotFound
	}

	// Check if object exists
	if objectType == "user" {
		if _, exists := s.users[objectID]; !exists {
			return ErrUserNotFound
		}
	} else if objectType == "group" {
		if _, exists := s.groups[objectID]; !exists {
			return ErrGroupNotFound
		}
	} else if objectType == "servicePrincipal" {
		if _, exists := s.servicePrincipals[objectID]; !exists {
			return ErrObjectNotFound
		}
	} else {
		return ErrInvalidObjType
	}

	// Check if already a member
	if memberMap, exists := s.members[groupID]; exists {
		if _, alreadyMember := memberMap[objectID]; alreadyMember {
			return ErrAlreadyMember
		}
		memberMap[objectID] = objectType
	} else {
		s.members[groupID] = map[string]string{objectID: objectType}
	}

	// Update the group's modification time
	if group, ok := s.groups[groupID]; ok {
		group.ModifiedAt = time.Now()
		s.groups[groupID] = group
	}

	return nil
}

func (s *memoryStore) RemoveMember(ctx context.Context, groupID, objectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if group exists
	if _, exists := s.groups[groupID]; !exists {
		return ErrGroupNotFound
	}

	// Check if object is a member
	if memberMap, exists := s.members[groupID]; exists {
		if _, isMember := memberMap[objectID]; !isMember {
			return ErrNotMember
		}
		delete(memberMap, objectID)
	} else {
		return ErrNotMember
	}

	// Update the group's modification time
	if group, ok := s.groups[groupID]; ok {
		group.ModifiedAt = time.Now()
		s.groups[groupID] = group
	}

	return nil
}

func (s *memoryStore) ListMembers(ctx context.Context, groupID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if group exists
	if _, exists := s.groups[groupID]; !exists {
		return nil, 0, ErrGroupNotFound
	}

	// Get members for this group
	memberMap, exists := s.members[groupID]
	if !exists {
		return []model.DirectoryObject{}, 0, nil
	}

	// Convert to DirectoryObject slice
	members := make([]model.DirectoryObject, 0, len(memberMap))
	for objectID, objectType := range memberMap {
		var obj model.DirectoryObject
		obj.ID = objectID

		// Set ODataType based on object type
		switch objectType {
		case "user":
			obj.ODataType = "#microsoft.graph.user"
			// Get display name from user
			if user, exists := s.users[objectID]; exists {
				obj.DisplayName = user.DisplayName
			}
		case "group":
			obj.ODataType = "#microsoft.graph.group"
			// Get display name from group
			if group, exists := s.groups[objectID]; exists {
				obj.DisplayName = group.DisplayName
			}
		case "servicePrincipal":
			obj.ODataType = "#microsoft.graph.servicePrincipal"
			// Get display name from service principal
			if sp, exists := s.servicePrincipals[objectID]; exists {
				obj.DisplayName = sp.DisplayName
			}
		}

		members = append(members, obj)
	}

	// Apply OData filtering, sorting, and pagination
	filteredMembers, totalCount, err := ApplyOData(members, opts)
	if err != nil {
		return nil, 0, err
	}

	return filteredMembers, totalCount, nil
}

func (s *memoryStore) ListTransitiveMembers(ctx context.Context, groupID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if group exists
	if _, exists := s.groups[groupID]; !exists {
		return nil, 0, ErrGroupNotFound
	}

	// Use BFS to find all transitive members
	visited := make(map[string]bool)
	queue := []string{groupID}
	allMembers := make([]model.DirectoryObject, 0)

	for len(queue) > 0 {
		currentGroupID := queue[0]
		queue = queue[1:]

		// Skip if already visited (cycle detection)
		if visited[currentGroupID] {
			continue
		}
		visited[currentGroupID] = true

		// Get direct members of current group
		memberMap, exists := s.members[currentGroupID]
		if !exists {
			continue
		}

		for objectID, objectType := range memberMap {
			// Skip if we've already processed this object
			if visited[objectID] {
				continue
			}

			var obj model.DirectoryObject
			obj.ID = objectID

			switch objectType {
			case "user":
				obj.ODataType = "#microsoft.graph.user"
				// Get display name from user
				if user, exists := s.users[objectID]; exists {
					obj.DisplayName = user.DisplayName
				}
				allMembers = append(allMembers, obj)
			case "group":
				obj.ODataType = "#microsoft.graph.group"
				// Get display name from group
				if group, exists := s.groups[objectID]; exists {
					obj.DisplayName = group.DisplayName
				}
				allMembers = append(allMembers, obj)
				// Add this group to queue to explore its members
				queue = append(queue, objectID)
			case "servicePrincipal":
				obj.ODataType = "#microsoft.graph.servicePrincipal"
				// Get display name from service principal
				if sp, exists := s.servicePrincipals[objectID]; exists {
					obj.DisplayName = sp.DisplayName
				}
				allMembers = append(allMembers, obj)
			}
		}
	}

	// Apply OData filtering, sorting, and pagination
	filteredMembers, totalCount, err := ApplyOData(allMembers, opts)
	if err != nil {
		return nil, 0, err
	}

	return filteredMembers, totalCount, nil
}

func (s *memoryStore) AddOwner(ctx context.Context, groupID, objectID, objectType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if group exists
	if _, exists := s.groups[groupID]; !exists {
		return ErrGroupNotFound
	}

	// Check if object exists
	if objectType == "user" {
		if _, exists := s.users[objectID]; !exists {
			return ErrUserNotFound
		}
	} else if objectType == "group" {
		if _, exists := s.groups[objectID]; !exists {
			return ErrGroupNotFound
		}
	} else if objectType == "servicePrincipal" {
		if _, exists := s.servicePrincipals[objectID]; !exists {
			return ErrObjectNotFound
		}
	} else {
		return ErrInvalidObjType
	}

	// Check if already an owner
	if ownerMap, exists := s.owners[groupID]; exists {
		if _, alreadyOwner := ownerMap[objectID]; alreadyOwner {
			return ErrAlreadyOwner
		}
		ownerMap[objectID] = objectType
	} else {
		s.owners[groupID] = map[string]string{objectID: objectType}
	}

	// Update the group's modification time
	if group, ok := s.groups[groupID]; ok {
		group.ModifiedAt = time.Now()
		s.groups[groupID] = group
	}

	return nil
}

func (s *memoryStore) RemoveOwner(ctx context.Context, groupID, objectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if group exists
	if _, exists := s.groups[groupID]; !exists {
		return ErrGroupNotFound
	}

	// Check if object is an owner
	if ownerMap, exists := s.owners[groupID]; exists {
		if _, isOwner := ownerMap[objectID]; !isOwner {
			return ErrNotOwner
		}
		delete(ownerMap, objectID)
	} else {
		return ErrNotOwner
	}

	// Update the group's modification time
	if group, ok := s.groups[groupID]; ok {
		group.ModifiedAt = time.Now()
		s.groups[groupID] = group
	}

	return nil
}

func (s *memoryStore) ListOwners(ctx context.Context, groupID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if group exists
	if _, exists := s.groups[groupID]; !exists {
		return nil, 0, ErrGroupNotFound
	}

	// Get owners for this group
	ownerMap, exists := s.owners[groupID]
	if !exists {
		return []model.DirectoryObject{}, 0, nil
	}

	// Convert to DirectoryObject slice
	owners := make([]model.DirectoryObject, 0, len(ownerMap))
	for objectID, objectType := range ownerMap {
		var obj model.DirectoryObject
		obj.ID = objectID

		// Set ODataType based on object type
		switch objectType {
		case "user":
			obj.ODataType = "#microsoft.graph.user"
			// Get display name from user
			if user, exists := s.users[objectID]; exists {
				obj.DisplayName = user.DisplayName
			}
		case "group":
			obj.ODataType = "#microsoft.graph.group"
			// Get display name from group
			if group, exists := s.groups[objectID]; exists {
				obj.DisplayName = group.DisplayName
			}
		case "servicePrincipal":
			obj.ODataType = "#microsoft.graph.servicePrincipal"
			// Get display name from service principal
			if sp, exists := s.servicePrincipals[objectID]; exists {
				obj.DisplayName = sp.DisplayName
			}
		}

		owners = append(owners, obj)
	}

	// Apply OData filtering, sorting, and pagination
	filteredOwners, totalCount, err := ApplyOData(owners, opts)
	if err != nil {
		return nil, 0, err
	}

	return filteredOwners, totalCount, nil
}

func (s *memoryStore) ListGroupMemberOf(ctx context.Context, groupID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if group exists
	if _, exists := s.groups[groupID]; !exists {
		return nil, 0, ErrGroupNotFound
	}

	// Find all groups that contain this group as a member
	memberOf := make([]model.DirectoryObject, 0)
	for gid, memberMap := range s.members {
		if objectType, isMember := memberMap[groupID]; isMember && objectType == "group" {
			// This group (gid) contains our group as a member
			if group, exists := s.groups[gid]; exists {
				obj := model.DirectoryObject{
					ODataType:   "#microsoft.graph.group",
					ID:          group.ID,
					DisplayName: group.DisplayName,
				}
				memberOf = append(memberOf, obj)
			}
		}
	}

	// Apply OData filtering, sorting, and pagination
	filteredMemberOf, totalCount, err := ApplyOData(memberOf, opts)
	if err != nil {
		return nil, 0, err
	}

	return filteredMemberOf, totalCount, nil
}

func (s *memoryStore) ListGroupTransitiveMemberOf(ctx context.Context, groupID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if group exists
	if _, exists := s.groups[groupID]; !exists {
		return nil, 0, ErrGroupNotFound
	}

	// Use BFS to find all transitive member-of relationships
	visited := make(map[string]bool)
	queue := []string{groupID}
	allMemberOf := make([]model.DirectoryObject, 0)

	for len(queue) > 0 {
		currentGroupID := queue[0]
		queue = queue[1:]

		// Skip if already visited (cycle detection)
		if visited[currentGroupID] {
			continue
		}
		visited[currentGroupID] = true

		// Find all groups that contain current group as a member
		for gid, memberMap := range s.members {
			if objectType, isMember := memberMap[currentGroupID]; isMember && objectType == "group" {
				// Skip if we've already processed this parent group
				if visited[gid] {
					continue
				}
				
				// This group (gid) contains our current group as a member
				if group, exists := s.groups[gid]; exists {
					obj := model.DirectoryObject{
						ODataType:   "#microsoft.graph.group",
						ID:          group.ID,
						DisplayName: group.DisplayName,
					}
					allMemberOf = append(allMemberOf, obj)
					// Add this parent group to queue to explore its parents
					queue = append(queue, gid)
				}
			}
		}
	}

	// Apply OData filtering, sorting, and pagination
	filteredMemberOf, totalCount, err := ApplyOData(allMemberOf, opts)
	if err != nil {
		return nil, 0, err
	}

	return filteredMemberOf, totalCount, nil
}

func (s *memoryStore) CheckMemberGroups(ctx context.Context, objectID string, groupIDs []string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if object exists (user or group)
	objectType := ""
	if _, exists := s.users[objectID]; exists {
		objectType = "user"
	} else if _, exists := s.groups[objectID]; exists {
		objectType = "group"
	} else {
		return nil, ErrObjectNotFound
	}

	result := make([]string, 0)
	
	// For each group ID in the input list, check if object is a transitive member
	for _, groupID := range groupIDs {
		if _, exists := s.groups[groupID]; !exists {
			continue // Skip non-existent groups
		}
		
		// Check if object is a transitive member of this group
		if s.isTransitiveMember(objectID, objectType, groupID, make(map[string]bool)) {
			result = append(result, groupID)
		}
	}

	return result, nil
}

func (s *memoryStore) GetMemberGroups(ctx context.Context, objectID string, securityEnabledOnly bool) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if object exists (user or group)
	objectType := ""
	if _, exists := s.users[objectID]; exists {
		objectType = "user"
	} else if _, exists := s.groups[objectID]; exists {
		objectType = "group"
	} else {
		return nil, ErrObjectNotFound
	}

	// Find all groups that contain this object as a transitive member
	result := make([]string, 0)
	
	// Use BFS starting from all groups
	for groupID, group := range s.groups {
		// Skip if securityEnabledOnly is true and group is not security enabled
		if securityEnabledOnly && (group.SecurityEnabled == nil || !*group.SecurityEnabled) {
			continue
		}
		
		// Check if object is a transitive member of this group
		if s.isTransitiveMember(objectID, objectType, groupID, make(map[string]bool)) {
			result = append(result, groupID)
		}
	}

	return result, nil
}

// Helper function to check if an object is a transitive member of a group
func (s *memoryStore) isTransitiveMember(objectID, objectType, groupID string, visited map[string]bool) bool {
	// Cycle detection
	if visited[groupID] {
		return false
	}
	visited[groupID] = true

	// Get direct members of this group
	memberMap, exists := s.members[groupID]
	if !exists {
		return false
	}

	// Check if object is a direct member
	if memberObjectType, isDirectMember := memberMap[objectID]; isDirectMember && memberObjectType == objectType {
		return true
	}

	// Check transitive membership through nested groups
	for memberID, memberType := range memberMap {
		if memberType == "group" {
			if s.isTransitiveMember(objectID, objectType, memberID, visited) {
				return true
			}
		}
	}

	return false
}

func (s *memoryStore) ResolveObjectType(ctx context.Context, objectID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.resolveObjectTypeUnsafe(objectID)
}

// resolveObjectTypeUnsafe checks if an object ID is a user, group, or servicePrincipal without acquiring locks
// Caller must hold at least s.mu.RLock()
func (s *memoryStore) resolveObjectTypeUnsafe(objectID string) (string, error) {
	// Check if it's a user
	if _, exists := s.users[objectID]; exists {
		return "user", nil
	}

	// Check if it's a group
	if _, exists := s.groups[objectID]; exists {
		return "group", nil
	}

	// Check if it's a service principal
	if _, exists := s.servicePrincipals[objectID]; exists {
		return "servicePrincipal", nil
	}

	return "", ErrObjectNotFound
}

// GetManager returns the manager of a user
func (s *memoryStore) GetManager(ctx context.Context, userID string) (*model.DirectoryObject, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if user exists
	if _, exists := s.users[userID]; !exists {
		return nil, ErrUserNotFound
	}

	// Get manager ID
	managerID, exists := s.managers[userID]
	if !exists {
		return nil, ErrManagerNotFound
	}

	// Get manager user
	manager, exists := s.users[managerID]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Return as DirectoryObject
	return &model.DirectoryObject{
		ODataType:   "#microsoft.graph.user",
		ID:          manager.ID,
		DisplayName: manager.DisplayName,
	}, nil
}

// SetManager sets the manager for a user
func (s *memoryStore) SetManager(ctx context.Context, userID, managerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	if _, exists := s.users[userID]; !exists {
		return ErrUserNotFound
	}

	// Check if manager exists
	if _, exists := s.users[managerID]; !exists {
		return ErrUserNotFound
	}

	// Set manager
	s.managers[userID] = managerID
	return nil
}

// RemoveManager removes the manager for a user
func (s *memoryStore) RemoveManager(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	if _, exists := s.users[userID]; !exists {
		return ErrUserNotFound
	}

	// Remove manager
	delete(s.managers, userID)
	return nil
}

// ListDirectReports returns all users who have the given user as their manager
func (s *memoryStore) ListDirectReports(ctx context.Context, userID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if user exists
	if _, exists := s.users[userID]; !exists {
		return nil, 0, ErrUserNotFound
	}

	// Find all users who have this user as manager
	directReports := make([]model.DirectoryObject, 0)
	for reportID, managerID := range s.managers {
		if managerID == userID {
			if user, exists := s.users[reportID]; exists {
				directReports = append(directReports, model.DirectoryObject{
					ODataType:   "#microsoft.graph.user",
					ID:          user.ID,
					DisplayName: user.DisplayName,
				})
			}
		}
	}

	// Apply OData filtering, sorting, and pagination
	filteredReports, totalCount, err := ApplyOData(directReports, opts)
	if err != nil {
		return nil, 0, err
	}

	return filteredReports, totalCount, nil
}

// ListUserMemberOf returns all groups the user is a direct member of
func (s *memoryStore) ListUserMemberOf(ctx context.Context, userID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if user exists
	if _, exists := s.users[userID]; !exists {
		return nil, 0, ErrUserNotFound
	}

	// Find all groups that contain this user as a direct member
	memberOf := make([]model.DirectoryObject, 0)
	for groupID, memberMap := range s.members {
		if objectType, isMember := memberMap[userID]; isMember && objectType == "user" {
			if group, exists := s.groups[groupID]; exists {
				memberOf = append(memberOf, model.DirectoryObject{
					ODataType:   "#microsoft.graph.group",
					ID:          group.ID,
					DisplayName: group.DisplayName,
				})
			}
		}
	}

	// Apply OData filtering, sorting, and pagination
	filteredMemberOf, totalCount, err := ApplyOData(memberOf, opts)
	if err != nil {
		return nil, 0, err
	}

	return filteredMemberOf, totalCount, nil
}

// ListUserTransitiveMemberOf returns all groups the user is transitively a member of
func (s *memoryStore) ListUserTransitiveMemberOf(ctx context.Context, userID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if user exists
	if _, exists := s.users[userID]; !exists {
		return nil, 0, ErrUserNotFound
	}

	// Use BFS to find all transitive member-of relationships
	visited := make(map[string]bool)
	queue := []string{userID}
	allMemberOf := make([]model.DirectoryObject, 0)

	for len(queue) > 0 {
		currentObjectID := queue[0]
		queue = queue[1:]

		// Skip if already visited (cycle detection)
		if visited[currentObjectID] {
			continue
		}
		visited[currentObjectID] = true

		// Find all groups that contain current object as a member
		for groupID, memberMap := range s.members {
			if _, isMember := memberMap[currentObjectID]; isMember {
				// Skip if we've already processed this group
				if visited[groupID] {
					continue
				}

				// This group contains our current object as a member
				if group, exists := s.groups[groupID]; exists {
					obj := model.DirectoryObject{
						ODataType:   "#microsoft.graph.group",
						ID:          group.ID,
						DisplayName: group.DisplayName,
					}
					allMemberOf = append(allMemberOf, obj)
					// Add this group to queue to explore its parents
					queue = append(queue, groupID)
				}
			}
		}
	}

	// Apply OData filtering, sorting, and pagination
	filteredMemberOf, totalCount, err := ApplyOData(allMemberOf, opts)
	if err != nil {
		return nil, 0, err
	}

	return filteredMemberOf, totalCount, nil
}

// GetDirectoryObjects returns directory objects for the given IDs, optionally filtered by type
func (s *memoryStore) GetDirectoryObjects(ctx context.Context, ids []string, types []string) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(ids))

	for _, id := range ids {
		// Check if it's a user
		if user, exists := s.users[id]; exists {
			if len(types) == 0 || contains(types, "user") {
				userJSON, err := json.Marshal(user)
				if err == nil {
					var userMap map[string]interface{}
					if err := json.Unmarshal(userJSON, &userMap); err == nil {
						userMap["@odata.type"] = "#microsoft.graph.user"
						result = append(result, userMap)
					}
				}
			}
			continue
		}

		// Check if it's a group
		if group, exists := s.groups[id]; exists {
			if len(types) == 0 || contains(types, "group") {
				groupJSON, err := json.Marshal(group)
				if err == nil {
					var groupMap map[string]interface{}
					if err := json.Unmarshal(groupJSON, &groupMap); err == nil {
						groupMap["@odata.type"] = "#microsoft.graph.group"
						result = append(result, groupMap)
					}
				}
			}
			continue
		}

		// Check if it's a service principal
		if sp, exists := s.servicePrincipals[id]; exists {
			if len(types) == 0 || contains(types, "servicePrincipal") {
				spJSON, err := json.Marshal(sp)
				if err == nil {
					var spMap map[string]interface{}
					if err := json.Unmarshal(spJSON, &spMap); err == nil {
						spMap["@odata.type"] = "#microsoft.graph.servicePrincipal"
						result = append(result, spMap)
					}
				}
			}
			continue
		}

		// Object not found - skip (Microsoft Graph returns only found objects)
	}

	return result, nil
}

// GetUsersDelta returns users changed since the delta token
func (s *memoryStore) GetUsersDelta(ctx context.Context, deltaToken string) ([]map[string]interface{}, string, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse delta token to get timestamp (if provided)
	var sinceTime time.Time
	if deltaToken != "" {
		tsBytes, err := base64.StdEncoding.DecodeString(deltaToken)
		if err == nil {
			ts, err := strconv.ParseInt(string(tsBytes), 10, 64)
			if err == nil {
				sinceTime = time.Unix(0, ts)
			}
		}
	}

	// Collect changed users (including deleted ones)
	result := make([]map[string]interface{}, 0)

	// Get current users modified after the delta token
	for _, user := range s.users {
		if deltaToken == "" || user.ModifiedAt.After(sinceTime) || user.ModifiedAt.Equal(sinceTime) {
			// Convert user to map
			userJSON, err := json.Marshal(user)
			if err != nil {
				continue
			}
			var userMap map[string]interface{}
			if err := json.Unmarshal(userJSON, &userMap); err != nil {
				continue
			}
			userMap["@odata.type"] = "#microsoft.graph.user"
			result = append(result, userMap)
		}
	}

	// Get deleted users modified after the delta token
	for userID, deletedAt := range s.deletedUsers {
		if deltaToken == "" || deletedAt.After(sinceTime) || deletedAt.Equal(sinceTime) {
			// Create a minimal representation of the deleted user
			deletedUser := map[string]interface{}{
				"id": userID,
				"@removed": map[string]interface{}{
					"reason": "deleted",
				},
			}
			result = append(result, deletedUser)
		}
	}

	// Generate a new delta token using base64-encoded timestamp
	newDeltaToken := base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))

	return result, newDeltaToken, len(result), nil
}

// GetGroupsDelta returns groups changed since the delta token
func (s *memoryStore) GetGroupsDelta(ctx context.Context, deltaToken string) ([]map[string]interface{}, string, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse delta token to get timestamp (if provided)
	var sinceTime time.Time
	if deltaToken != "" {
		tsBytes, err := base64.StdEncoding.DecodeString(deltaToken)
		if err == nil {
			ts, err := strconv.ParseInt(string(tsBytes), 10, 64)
			if err == nil {
				sinceTime = time.Unix(0, ts)
			}
		}
	}

	// Collect changed groups (including deleted ones)
	result := make([]map[string]interface{}, 0)

	// Get current groups modified after the delta token
	for _, group := range s.groups {
		if deltaToken == "" || group.ModifiedAt.After(sinceTime) || group.ModifiedAt.Equal(sinceTime) {
			// Convert group to map
			groupJSON, err := json.Marshal(group)
			if err != nil {
				continue
			}
			var groupMap map[string]interface{}
			if err := json.Unmarshal(groupJSON, &groupMap); err != nil {
				continue
			}
			groupMap["@odata.type"] = "#microsoft.graph.group"
			result = append(result, groupMap)
		}
	}

	// Get deleted groups modified after the delta token
	for groupID, deletedAt := range s.deletedGroups {
		if deltaToken == "" || deletedAt.After(sinceTime) || deletedAt.Equal(sinceTime) {
			// Create a minimal representation of the deleted group
			deletedGroup := map[string]interface{}{
				"id": groupID,
				"@removed": map[string]interface{}{
					"reason": "deleted",
				},
			}
			result = append(result, deletedGroup)
		}
	}

	// Generate a new delta token using base64-encoded timestamp
	newDeltaToken := base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))

	return result, newDeltaToken, len(result), nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}