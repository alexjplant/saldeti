package store

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/saldeti/saldeti/internal/model"
)

// generateSecretText generates a random secret text for password credentials.
func generateSecretText() string {
	return strings.ReplaceAll(uuid.New().String()+uuid.New().String(), "-", "")[:32]
}

// resolvePrincipalInfo resolves the display name and type for a principal object.
// Caller must hold at least s.mu.RLock().
func (s *memoryStore) resolvePrincipalInfo(objectID string) (displayName string, objectType string, err error) {
	if user, exists := s.users[objectID]; exists {
		return user.DisplayName, "User", nil
	}
	if group, exists := s.groups[objectID]; exists {
		return group.DisplayName, "Group", nil
	}
	if sp, exists := s.servicePrincipals[objectID]; exists {
		return sp.DisplayName, "ServicePrincipal", nil
	}
	return "", "", ErrObjectNotFound
}

// ========== Application CRUD ==========

func (s *memoryStore) ListApplications(ctx context.Context, opts model.ListOptions) ([]model.Application, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apps := make([]model.Application, 0, len(s.applications))
	for _, app := range s.applications {
		apps = append(apps, app)
	}

	filtered, totalCount, err := ApplyOData(apps, opts)
	if err != nil {
		return nil, 0, err
	}
	return filtered, totalCount, nil
}

func (s *memoryStore) GetApplication(ctx context.Context, id string) (*model.Application, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	app, exists := s.applications[id]
	if !exists {
		return nil, ErrApplicationNotFound
	}
	return &app, nil
}

func (s *memoryStore) GetApplicationByAppID(ctx context.Context, appId string) (*model.Application, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, app := range s.applications {
		if app.AppID == appId {
			return &app, nil
		}
	}
	return nil, ErrApplicationNotFound
}

func (s *memoryStore) CreateApplication(ctx context.Context, app model.Application) (model.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if app.ID == "" {
		app.ID = uuid.New().String()
	}
	if app.AppID == "" {
		app.AppID = uuid.New().String()
	}

	// Check for duplicate appId
	for _, existing := range s.applications {
		if existing.AppID == app.AppID {
			return model.Application{}, ErrDuplicateAppID
		}
	}

	if app.ODataType == "" {
		app.ODataType = "#microsoft.graph.application"
	}
	if app.CreatedDateTime == nil {
		now := time.Now()
		app.CreatedDateTime = &now
	}
	app.ModifiedAt = time.Now()

	if app.SignInAudience == "" {
		app.SignInAudience = "AzureADMyOrg"
	}
	if app.PasswordCredentials == nil {
		app.PasswordCredentials = []model.PasswordCredential{}
	}
	if app.KeyCredentials == nil {
		app.KeyCredentials = []model.KeyCredential{}
	}
	if app.IdentifierUris == nil {
		app.IdentifierUris = []string{}
	}

	s.applications[app.ID] = app
	s.appOwners[app.ID] = make(map[string]string)
	s.appExtensions[app.ID] = make(map[string]model.ExtensionProperty)

	// Auto-create or update service principal
	accountEnabled := true
	spObj := model.ServicePrincipal{
		ODataType:      "#microsoft.graph.servicePrincipal",
		AppID:          app.AppID,
		DisplayName:    app.DisplayName,
		Description:    app.Description,
		AppRoles:       app.AppRoles,
		AccountEnabled: &accountEnabled,
		ModifiedAt:     time.Now(),
	}

	if app.API != nil {
		spObj.OAuth2PermissionScopes = app.API.OAuth2PermissionScopes
	}

	// Check if SP already exists (e.g., from RegisterClient)
	existingSPID := ""
	for spID, sp := range s.servicePrincipals {
		if sp.AppID == app.AppID {
			existingSPID = spID
			break
		}
	}

	if existingSPID != "" {
		// Update existing SP
		sp := s.servicePrincipals[existingSPID]
		sp.DisplayName = app.DisplayName
		sp.Description = app.Description
		sp.AppRoles = app.AppRoles
		if app.API != nil {
			sp.OAuth2PermissionScopes = app.API.OAuth2PermissionScopes
		}
		sp.ModifiedAt = time.Now()
		s.servicePrincipals[existingSPID] = sp
	} else {
		// Create new SP
		spObj.ID = uuid.New().String()
		spObj.ServicePrincipalNames = []string{spObj.AppID}
		s.servicePrincipals[spObj.ID] = spObj
		s.spOwners[spObj.ID] = make(map[string]string)
		s.spMemberOf[spObj.ID] = make(map[string]string)
	}

	return app, nil
}

func (s *memoryStore) UpdateApplication(ctx context.Context, id string, patch map[string]interface{}) (*model.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.applications[id]
	if !exists {
		return nil, ErrApplicationNotFound
	}

	updated, err := applyPatch(app, patch)
	if err != nil {
		return nil, err
	}

	updated.ModifiedAt = time.Now()
	s.applications[id] = updated
	return &updated, nil
}

func (s *memoryStore) DeleteApplication(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.applications[id]
	if !exists {
		return ErrApplicationNotFound
	}

	s.deletedApplications[id] = time.Now()
	delete(s.applications, id)
	delete(s.appOwners, id)
	delete(s.appExtensions, id)

	// Find and delete corresponding SP
	for spID, sp := range s.servicePrincipals {
		if sp.AppID == app.AppID {
			s.deletedSPs[spID] = time.Now()
			delete(s.servicePrincipals, spID)
			delete(s.spOwners, spID)
			delete(s.spMemberOf, spID)

			// Clean up app role assignments where SP is the resource
			for assignID, assignment := range s.appRoleAssignments {
				if assignment.ResourceID == spID {
					delete(s.appRoleAssignments, assignID)
				}
			}

			// Clean up app role assignments where SP is the principal
			for assignID, assignment := range s.appRoleAssignments {
				if assignment.PrincipalID == spID {
					delete(s.appRoleAssignments, assignID)
				}
			}

			// Clean up OAuth2 permission grants where SP is the client or resource
			for grantID, grant := range s.oauth2PermissionGrants {
				if grant.ClientID == spID || grant.ResourceID == spID {
					delete(s.oauth2PermissionGrants, grantID)
				}
			}
			break
		}
	}

	return nil
}

// ========== Credential Management ==========

func (s *memoryStore) AddApplicationPassword(ctx context.Context, appID string, cred model.PasswordCredential) (model.PasswordCredential, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.applications[appID]
	if !exists {
		return model.PasswordCredential{}, ErrApplicationNotFound
	}

	if cred.KeyID == "" {
		cred.KeyID = uuid.New().String()
	}

	cred.SecretText = generateSecretText()
	now := time.Now()
	cred.StartDateTime = &now
	end := now.Add(2 * 365 * 24 * time.Hour)
	cred.EndDateTime = &end
	if len(cred.SecretText) >= 3 {
		cred.Hint = cred.SecretText[:3] + "***"
	}

	app.PasswordCredentials = append(app.PasswordCredentials, cred)
	app.ModifiedAt = time.Now()
	s.applications[appID] = app

	return cred, nil
}

func (s *memoryStore) RemoveApplicationPassword(ctx context.Context, appID, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.applications[appID]
	if !exists {
		return ErrApplicationNotFound
	}

	found := false
	for i, pc := range app.PasswordCredentials {
		if pc.KeyID == keyID {
			app.PasswordCredentials = append(app.PasswordCredentials[:i], app.PasswordCredentials[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return ErrCredentialNotFound
	}

	app.ModifiedAt = time.Now()
	s.applications[appID] = app
	return nil
}

func (s *memoryStore) AddApplicationKey(ctx context.Context, appID string, cred model.KeyCredential) (model.KeyCredential, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.applications[appID]
	if !exists {
		return model.KeyCredential{}, ErrApplicationNotFound
	}

	if cred.KeyID == "" {
		cred.KeyID = uuid.New().String()
	}
	if cred.StartDateTime == nil {
		now := time.Now()
		cred.StartDateTime = &now
	}

	app.KeyCredentials = append(app.KeyCredentials, cred)
	app.ModifiedAt = time.Now()
	s.applications[appID] = app

	return cred, nil
}

func (s *memoryStore) RemoveApplicationKey(ctx context.Context, appID, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.applications[appID]
	if !exists {
		return ErrApplicationNotFound
	}

	found := false
	for i, kc := range app.KeyCredentials {
		if kc.KeyID == keyID {
			app.KeyCredentials = append(app.KeyCredentials[:i], app.KeyCredentials[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return ErrCredentialNotFound
	}

	app.ModifiedAt = time.Now()
	s.applications[appID] = app
	return nil
}

// ========== Application Owners ==========

func (s *memoryStore) ListApplicationOwners(ctx context.Context, appID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.applications[appID]; !exists {
		return nil, 0, ErrApplicationNotFound
	}

	ownerMap, exists := s.appOwners[appID]
	if !exists {
		return []model.DirectoryObject{}, 0, nil
	}

	owners := make([]model.DirectoryObject, 0, len(ownerMap))
	for objectID, objectType := range ownerMap {
		var obj model.DirectoryObject
		obj.ID = objectID

		switch objectType {
		case "user":
			obj.ODataType = "#microsoft.graph.user"
			if user, exists := s.users[objectID]; exists {
				obj.DisplayName = user.DisplayName
			}
		case "group":
			obj.ODataType = "#microsoft.graph.group"
			if group, exists := s.groups[objectID]; exists {
				obj.DisplayName = group.DisplayName
			}
		case "servicePrincipal":
			obj.ODataType = "#microsoft.graph.servicePrincipal"
			if sp, exists := s.servicePrincipals[objectID]; exists {
				obj.DisplayName = sp.DisplayName
			}
		}

		owners = append(owners, obj)
	}

	filtered, totalCount, err := ApplyOData(owners, opts)
	if err != nil {
		return nil, 0, err
	}
	return filtered, totalCount, nil
}

func (s *memoryStore) AddApplicationOwner(ctx context.Context, appID, objectID, objectType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.applications[appID]; !exists {
		return ErrApplicationNotFound
	}

	// Validate object exists
	if err := s.validateObjectExists(objectID, objectType); err != nil {
		return err
	}

	if ownerMap, exists := s.appOwners[appID]; exists {
		if _, alreadyOwner := ownerMap[objectID]; alreadyOwner {
			return ErrAlreadyAppOwner
		}
		ownerMap[objectID] = objectType
	} else {
		s.appOwners[appID] = map[string]string{objectID: objectType}
	}

	if app, ok := s.applications[appID]; ok {
		app.ModifiedAt = time.Now()
		s.applications[appID] = app
	}
	return nil
}

func (s *memoryStore) RemoveApplicationOwner(ctx context.Context, appID, objectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.applications[appID]; !exists {
		return ErrApplicationNotFound
	}

	if ownerMap, exists := s.appOwners[appID]; exists {
		if _, isOwner := ownerMap[objectID]; !isOwner {
			return ErrNotAppOwner
		}
		delete(ownerMap, objectID)
	} else {
		return ErrNotAppOwner
	}

	if app, ok := s.applications[appID]; ok {
		app.ModifiedAt = time.Now()
		s.applications[appID] = app
	}
	return nil
}

// ========== Extension Properties ==========

func (s *memoryStore) ListExtensionProperties(ctx context.Context, appID string) ([]model.ExtensionProperty, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.applications[appID]; !exists {
		return nil, ErrApplicationNotFound
	}

	extMap, exists := s.appExtensions[appID]
	if !exists {
		return []model.ExtensionProperty{}, nil
	}

	result := make([]model.ExtensionProperty, 0, len(extMap))
	for _, ep := range extMap {
		result = append(result, ep)
	}
	return result, nil
}

func (s *memoryStore) CreateExtensionProperty(ctx context.Context, appID string, ep model.ExtensionProperty) (model.ExtensionProperty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.applications[appID]
	if !exists {
		return model.ExtensionProperty{}, ErrApplicationNotFound
	}

	if ep.ID == "" {
		ep.ID = uuid.New().String()
	}
	if ep.Name == "" {
		ep.Name = "extension_" + strings.ReplaceAll(app.AppID, "-", "") + "_" + strings.ReplaceAll(uuid.New().String()[:8], "-", "")
	}
	ep.AppDisplayName = app.DisplayName

	if s.appExtensions[appID] == nil {
		s.appExtensions[appID] = make(map[string]model.ExtensionProperty)
	}
	s.appExtensions[appID][ep.ID] = ep

	return ep, nil
}

func (s *memoryStore) DeleteExtensionProperty(ctx context.Context, appID, extID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.applications[appID]; !exists {
		return ErrApplicationNotFound
	}

	extMap, exists := s.appExtensions[appID]
	if !exists {
		return ErrExtensionNotFound
	}
	if _, exists := extMap[extID]; !exists {
		return ErrExtensionNotFound
	}

	delete(extMap, extID)
	return nil
}

// ========== Application Delta ==========

func (s *memoryStore) GetApplicationsDelta(ctx context.Context, deltaToken string) ([]map[string]interface{}, string, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

	result := make([]map[string]interface{}, 0)

	for _, app := range s.applications {
		if deltaToken == "" || app.ModifiedAt.After(sinceTime) {
			appJSON, err := json.Marshal(app)
			if err != nil {
				continue
			}
			var appMap map[string]interface{}
			if err := json.Unmarshal(appJSON, &appMap); err != nil {
				continue
			}
			appMap["@odata.type"] = "#microsoft.graph.application"
			result = append(result, appMap)
		}
	}

	for appID, deletedAt := range s.deletedApplications {
		if deltaToken == "" || deletedAt.After(sinceTime) {
			result = append(result, map[string]interface{}{
				"id": appID,
				"@removed": map[string]interface{}{
					"reason": "deleted",
				},
			})
		}
	}

	newDeltaToken := base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	return result, newDeltaToken, len(result), nil
}

// ========== Service Principal CRUD ==========

func (s *memoryStore) ListServicePrincipals(ctx context.Context, opts model.ListOptions) ([]model.ServicePrincipal, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sps := make([]model.ServicePrincipal, 0, len(s.servicePrincipals))
	for _, sp := range s.servicePrincipals {
		sps = append(sps, sp)
	}

	filtered, totalCount, err := ApplyOData(sps, opts)
	if err != nil {
		return nil, 0, err
	}
	return filtered, totalCount, nil
}

func (s *memoryStore) GetServicePrincipal(ctx context.Context, id string) (*model.ServicePrincipal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sp, exists := s.servicePrincipals[id]
	if !exists {
		return nil, ErrServicePrincipalNotFound
	}
	return &sp, nil
}

func (s *memoryStore) GetServicePrincipalByAppID(ctx context.Context, appId string) (*model.ServicePrincipal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, sp := range s.servicePrincipals {
		if sp.AppID == appId {
			return &sp, nil
		}
	}
	return nil, ErrServicePrincipalNotFound
}

func (s *memoryStore) CreateServicePrincipal(ctx context.Context, sp model.ServicePrincipal) (model.ServicePrincipal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sp.AppID == "" {
		return model.ServicePrincipal{}, fmt.Errorf("appId is required")
	}

	// Find matching application (optional — if found, we enrich the SP from it)
	var app *model.Application
	for _, a := range s.applications {
		if a.AppID == sp.AppID {
			app = &a
			break
		}
	}
	// app may be nil — that's OK, we create the SP anyway

	// Check for duplicate
	for _, existing := range s.servicePrincipals {
		if existing.AppID == sp.AppID {
			return model.ServicePrincipal{}, ErrDuplicateSPAppID
		}
	}

	if sp.ID == "" {
		sp.ID = uuid.New().String()
	}
	if sp.ODataType == "" {
		sp.ODataType = "#microsoft.graph.servicePrincipal"
	}

	// If a matching application was found, enrich the SP with app fields
	if app != nil {
		if sp.DisplayName == "" {
			sp.DisplayName = app.DisplayName
		}
		if sp.Description == "" {
			sp.Description = app.Description
		}
		sp.AppRoles = app.AppRoles
		if app.API != nil {
			sp.OAuth2PermissionScopes = app.API.OAuth2PermissionScopes
		}
	}
	if sp.ServicePrincipalNames == nil {
		sp.ServicePrincipalNames = []string{sp.AppID}
	}
	if sp.AccountEnabled == nil {
		accountEnabled := true
		sp.AccountEnabled = &accountEnabled
	}
	sp.ModifiedAt = time.Now()

	s.servicePrincipals[sp.ID] = sp
	s.spOwners[sp.ID] = make(map[string]string)
	s.spMemberOf[sp.ID] = make(map[string]string)

	return sp, nil
}

func (s *memoryStore) UpdateServicePrincipal(ctx context.Context, id string, patch map[string]interface{}) (*model.ServicePrincipal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sp, exists := s.servicePrincipals[id]
	if !exists {
		return nil, ErrServicePrincipalNotFound
	}

	updated, err := applyPatch(sp, patch)
	if err != nil {
		return nil, err
	}

	updated.ModifiedAt = time.Now()
	s.servicePrincipals[id] = updated
	return &updated, nil
}

func (s *memoryStore) DeleteServicePrincipal(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.servicePrincipals[id]; !exists {
		return ErrServicePrincipalNotFound
	}

	s.deletedSPs[id] = time.Now()
	delete(s.servicePrincipals, id)
	delete(s.spOwners, id)
	delete(s.spMemberOf, id)

	// Clean up app role assignments where SP is the resource
	for assignID, assignment := range s.appRoleAssignments {
		if assignment.ResourceID == id {
			delete(s.appRoleAssignments, assignID)
		}
	}

	// Clean up app role assignments where SP is the principal
	for assignID, assignment := range s.appRoleAssignments {
		if assignment.PrincipalID == id {
			delete(s.appRoleAssignments, assignID)
		}
	}

	// Clean up OAuth2 permission grants where SP is the client or resource
	for grantID, grant := range s.oauth2PermissionGrants {
		if grant.ClientID == id || grant.ResourceID == id {
			delete(s.oauth2PermissionGrants, grantID)
		}
	}

	return nil
}

func (s *memoryStore) UpdateSPCredentials(ctx context.Context, spID string, update func(*model.ServicePrincipal) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sp, exists := s.servicePrincipals[spID]
	if !exists {
		return ErrServicePrincipalNotFound
	}

	if err := update(&sp); err != nil {
		return err
	}

	sp.ModifiedAt = time.Now()
	s.servicePrincipals[spID] = sp

	return nil
}

// ========== SP Owners ==========

func (s *memoryStore) ListSPOwners(ctx context.Context, spID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.servicePrincipals[spID]; !exists {
		return nil, 0, ErrServicePrincipalNotFound
	}

	ownerMap, exists := s.spOwners[spID]
	if !exists {
		return []model.DirectoryObject{}, 0, nil
	}

	owners := make([]model.DirectoryObject, 0, len(ownerMap))
	for objectID, objectType := range ownerMap {
		var obj model.DirectoryObject
		obj.ID = objectID

		switch objectType {
		case "user":
			obj.ODataType = "#microsoft.graph.user"
			if user, exists := s.users[objectID]; exists {
				obj.DisplayName = user.DisplayName
			}
		case "group":
			obj.ODataType = "#microsoft.graph.group"
			if group, exists := s.groups[objectID]; exists {
				obj.DisplayName = group.DisplayName
			}
		case "servicePrincipal":
			obj.ODataType = "#microsoft.graph.servicePrincipal"
			if sp, exists := s.servicePrincipals[objectID]; exists {
				obj.DisplayName = sp.DisplayName
			}
		}

		owners = append(owners, obj)
	}

	filtered, totalCount, err := ApplyOData(owners, opts)
	if err != nil {
		return nil, 0, err
	}
	return filtered, totalCount, nil
}

func (s *memoryStore) AddSPOwner(ctx context.Context, spID, objectID, objectType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.servicePrincipals[spID]; !exists {
		return ErrServicePrincipalNotFound
	}

	if err := s.validateObjectExists(objectID, objectType); err != nil {
		return err
	}

	if ownerMap, exists := s.spOwners[spID]; exists {
		if _, alreadyOwner := ownerMap[objectID]; alreadyOwner {
			return ErrAlreadySPOwner
		}
		ownerMap[objectID] = objectType
	} else {
		s.spOwners[spID] = map[string]string{objectID: objectType}
	}

	if sp, ok := s.servicePrincipals[spID]; ok {
		sp.ModifiedAt = time.Now()
		s.servicePrincipals[spID] = sp
	}
	return nil
}

func (s *memoryStore) RemoveSPOwner(ctx context.Context, spID, objectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.servicePrincipals[spID]; !exists {
		return ErrServicePrincipalNotFound
	}

	if ownerMap, exists := s.spOwners[spID]; exists {
		if _, isOwner := ownerMap[objectID]; !isOwner {
			return ErrNotSPOwner
		}
		delete(ownerMap, objectID)
	} else {
		return ErrNotSPOwner
	}

	if sp, ok := s.servicePrincipals[spID]; ok {
		sp.ModifiedAt = time.Now()
		s.servicePrincipals[spID] = sp
	}
	return nil
}

// ========== SP MemberOf ==========

func (s *memoryStore) ListSPMemberOf(ctx context.Context, spID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.servicePrincipals[spID]; !exists {
		return nil, 0, ErrServicePrincipalNotFound
	}

	memberOf := make([]model.DirectoryObject, 0)
	for groupID, memberMap := range s.members {
		if objectType, isMember := memberMap[spID]; isMember && objectType == "servicePrincipal" {
			if group, exists := s.groups[groupID]; exists {
				memberOf = append(memberOf, model.DirectoryObject{
					ODataType:   "#microsoft.graph.group",
					ID:          group.ID,
					DisplayName: group.DisplayName,
				})
			}
		}
	}

	filtered, totalCount, err := ApplyOData(memberOf, opts)
	if err != nil {
		return nil, 0, err
	}
	return filtered, totalCount, nil
}

func (s *memoryStore) ListSPTransitiveMemberOf(ctx context.Context, spID string, opts model.ListOptions) ([]model.DirectoryObject, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.servicePrincipals[spID]; !exists {
		return nil, 0, ErrServicePrincipalNotFound
	}

	visited := make(map[string]bool)
	queue := []string{spID}
	allMemberOf := make([]model.DirectoryObject, 0)

	for len(queue) > 0 {
		currentObjectID := queue[0]
		queue = queue[1:]

		if visited[currentObjectID] {
			continue
		}
		visited[currentObjectID] = true

		for groupID, memberMap := range s.members {
			if _, isMember := memberMap[currentObjectID]; isMember {
				if visited[groupID] {
					continue
				}
				if group, exists := s.groups[groupID]; exists {
					allMemberOf = append(allMemberOf, model.DirectoryObject{
						ODataType:   "#microsoft.graph.group",
						ID:          group.ID,
						DisplayName: group.DisplayName,
					})
					queue = append(queue, groupID)
				}
			}
		}
	}

	filtered, totalCount, err := ApplyOData(allMemberOf, opts)
	if err != nil {
		return nil, 0, err
	}
	return filtered, totalCount, nil
}

// ========== App Role Assignments ==========

func (s *memoryStore) CreateAppRoleAssignment(ctx context.Context, resourceID, principalID, appRoleID string) (model.AppRoleAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate resource SP
	resourceSP, exists := s.servicePrincipals[resourceID]
	if !exists {
		return model.AppRoleAssignment{}, ErrServicePrincipalNotFound
	}

	// Validate principal exists
	displayName, principalType, err := s.resolvePrincipalInfo(principalID)
	if err != nil {
		return model.AppRoleAssignment{}, err
	}

	// Validate appRoleID exists on resource SP
	roleFound := false
	for _, role := range resourceSP.AppRoles {
		if role.ID == appRoleID {
			roleFound = true
			break
		}
	}
	if !roleFound {
		return model.AppRoleAssignment{}, ErrAppRoleNotFound
	}

	now := time.Now()
	assignment := model.AppRoleAssignment{
		ODataType:            "#microsoft.graph.appRoleAssignment",
		ID:                   uuid.New().String(),
		AppRoleID:            appRoleID,
		CreatedDateTime:      &now,
		PrincipalDisplayName: displayName,
		PrincipalID:          principalID,
		PrincipalType:        principalType,
		ResourceDisplayName:  resourceSP.DisplayName,
		ResourceID:           resourceID,
	}

	s.appRoleAssignments[assignment.ID] = assignment
	return assignment, nil
}

func (s *memoryStore) ListAppRoleAssignments(ctx context.Context, principalID string, opts model.ListOptions) ([]model.AppRoleAssignment, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assignments := make([]model.AppRoleAssignment, 0)
	for _, a := range s.appRoleAssignments {
		if a.PrincipalID == principalID {
			assignments = append(assignments, a)
		}
	}

	filtered, totalCount, err := ApplyOData(assignments, opts)
	if err != nil {
		return nil, 0, err
	}
	return filtered, totalCount, nil
}

func (s *memoryStore) ListAppRoleAssignedTo(ctx context.Context, resourceID string, opts model.ListOptions) ([]model.AppRoleAssignment, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assignments := make([]model.AppRoleAssignment, 0)
	for _, a := range s.appRoleAssignments {
		if a.ResourceID == resourceID {
			assignments = append(assignments, a)
		}
	}

	filtered, totalCount, err := ApplyOData(assignments, opts)
	if err != nil {
		return nil, 0, err
	}
	return filtered, totalCount, nil
}

func (s *memoryStore) DeleteAppRoleAssignment(ctx context.Context, assignmentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.appRoleAssignments[assignmentID]; !exists {
		return ErrAssignmentNotFound
	}
	delete(s.appRoleAssignments, assignmentID)
	return nil
}

// ========== OAuth2 Permission Grants ==========

func (s *memoryStore) ListOAuth2PermissionGrants(ctx context.Context, opts model.ListOptions) ([]model.OAuth2PermissionGrant, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	grants := make([]model.OAuth2PermissionGrant, 0, len(s.oauth2PermissionGrants))
	for _, g := range s.oauth2PermissionGrants {
		grants = append(grants, g)
	}

	filtered, totalCount, err := ApplyOData(grants, opts)
	if err != nil {
		return nil, 0, err
	}
	return filtered, totalCount, nil
}

func (s *memoryStore) GetOAuth2PermissionGrant(ctx context.Context, id string) (*model.OAuth2PermissionGrant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	grant, exists := s.oauth2PermissionGrants[id]
	if !exists {
		return nil, ErrGrantNotFound
	}
	return &grant, nil
}

func (s *memoryStore) CreateOAuth2PermissionGrant(ctx context.Context, grant model.OAuth2PermissionGrant) (model.OAuth2PermissionGrant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if grant.ID == "" {
		grant.ID = uuid.New().String()
	}
	if grant.ODataType == "" {
		grant.ODataType = "#microsoft.graph.oAuth2PermissionGrant"
	}

	s.oauth2PermissionGrants[grant.ID] = grant
	return grant, nil
}

func (s *memoryStore) UpdateOAuth2PermissionGrant(ctx context.Context, id string, patch map[string]interface{}) (*model.OAuth2PermissionGrant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	grant, exists := s.oauth2PermissionGrants[id]
	if !exists {
		return nil, ErrGrantNotFound
	}

	updated, err := applyPatch(grant, patch)
	if err != nil {
		return nil, err
	}

	s.oauth2PermissionGrants[id] = updated
	return &updated, nil
}

func (s *memoryStore) DeleteOAuth2PermissionGrant(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.oauth2PermissionGrants[id]; !exists {
		return ErrGrantNotFound
	}
	delete(s.oauth2PermissionGrants, id)
	return nil
}

// ========== Helper: validate object exists ==========

// validateObjectExists checks if an object exists in the store.
// Caller must hold at least s.mu.Lock().
func (s *memoryStore) validateObjectExists(objectID, objectType string) error {
	switch objectType {
	case "user":
		if _, exists := s.users[objectID]; !exists {
			return ErrUserNotFound
		}
	case "group":
		if _, exists := s.groups[objectID]; !exists {
			return ErrGroupNotFound
		}
	case "servicePrincipal":
		if _, exists := s.servicePrincipals[objectID]; !exists {
			return ErrObjectNotFound
		}
	default:
		return ErrInvalidObjType
	}
	return nil
}
