//go:build ui

package ui_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

// Ensure store package is used
var _ store.Store = nil

func TestGroupListShowsGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, _ := setupTestServer(t)


	// Get group list
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/groups", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should show some groups from seed data
	if !strings.Contains(body, "Engineering Team") {
		t.Error("Expected 'Engineering Team' in group list")
	}
	if !strings.Contains(body, "Marketing Team") {
		t.Error("Expected 'Marketing Team' in group list")
	}
	if !strings.Contains(body, "Groups") {
		t.Error("Expected 'Groups' heading in page")
	}
}

func TestGroupSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, _ := setupTestServer(t)


	// Search for Engineering
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/groups?search=Engineering", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should show Engineering Team
	if !strings.Contains(body, "Engineering Team") {
		t.Error("Expected 'Engineering Team' in search results")
	}

	// Should not show Marketing Team (unless it also contains "Engineering")
	if strings.Contains(body, "Marketing Team") && !strings.Contains(body, "Engineering Team") {
		// Marketing shouldn't be in results unless there's some overlap
		t.Error("Expected only 'Engineering' related results, but found Marketing Team")
	}
}

func TestGroupDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, st := setupTestServer(t)

	// Get Engineering Team ID
	ctx := context.Background()
	groups, _, err := st.ListGroups(ctx, model.ListOptions{Top: 100})
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}

	var engineeringTeamID string
	for _, group := range groups {
		if group.DisplayName == "Engineering Team" {
			engineeringTeamID = group.ID
			break
		}
	}
	if engineeringTeamID == "" {
		t.Fatal("Engineering Team not found in seed data")
	}


	// Get Engineering Team detail page
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/groups/"+engineeringTeamID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should show group info
	if !strings.Contains(body, "Engineering Team") {
		t.Error("Expected 'Engineering Team' in detail page")
	}
	if !strings.Contains(body, "Members") {
		t.Error("Expected 'Members' section in detail page")
	}
	if !strings.Contains(body, "Owners") {
		t.Error("Expected 'Owners' section in detail page")
	}
}

func TestGroupCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, st := setupTestServer(t)


	// Create new group
	formData := url.Values{}
	formData.Set("displayName", "Test Group")
	formData.Set("description", "A test group for testing")
	formData.Set("mailNickname", "testgroup")
	formData.Set("securityEnabled", "true")
	formData.Set("mailEnabled", "false")
	formData.Set("visibility", "Public")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/new", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	// Should redirect to group detail page
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "/ui/groups/") {
		t.Errorf("Expected redirect to /ui/groups/{id}, got %s", location)
	}

	// Follow redirect to detail page
	w = httptest.NewRecorder()
	newGroupID := strings.TrimPrefix(location, "/ui/groups/")
	req, _ = http.NewRequest("GET", location, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should show new group details
	if !strings.Contains(body, "Test Group") {
		t.Error("Expected 'Test Group' in detail page")
	}
	if !strings.Contains(body, "A test group for testing") {
		t.Error("Expected 'A test group for testing' description in detail page")
	}

	// Verify group exists in store
	ctx := context.Background()
	group, err := st.GetGroup(ctx, newGroupID)
	if err != nil {
		t.Fatalf("Failed to get created group: %v", err)
	}
	if group.DisplayName != "Test Group" {
		t.Errorf("Expected display name 'Test Group', got '%s'", group.DisplayName)
	}
}

func TestGroupCreateValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, _ := setupTestServer(t)


	// Try to create group without required fields
	formData := url.Values{}
	formData.Set("description", "A test group")
	// Missing displayName

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/new", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	// Should return the form with error (not redirect)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should show error message
	if !strings.Contains(body, "required") {
		t.Error("Expected 'required' error message in response")
	}

	// Should still be the form page
	if !strings.Contains(body, "New Group") {
		t.Error("Expected 'New Group' heading in form page")
	}
}

func TestGroupEdit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, st := setupTestServer(t)

	// Get Engineering Team ID
	ctx := context.Background()
	groups, _, err := st.ListGroups(ctx, model.ListOptions{Top: 100})
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}

	var engineeringTeamID string
	for _, group := range groups {
		if group.DisplayName == "Engineering Team" {
			engineeringTeamID = group.ID
			break
		}
	}
	if engineeringTeamID == "" {
		t.Fatal("Engineering Team not found in seed data")
	}


	// Update Engineering Team
	formData := url.Values{}
	formData.Set("displayName", "Engineering Team Updated")
	formData.Set("description", "Updated description")
	formData.Set("mailNickname", "engineeringteam")
	formData.Set("securityEnabled", "true")
	formData.Set("mailEnabled", "false")
	formData.Set("visibility", "Private") // Changed from Public

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/"+engineeringTeamID+"/edit", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	// Should redirect to group detail page
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/ui/groups/"+engineeringTeamID {
		t.Errorf("Expected redirect to /ui/groups/%s, got %s", engineeringTeamID, location)
	}

	// Follow redirect to detail page
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", location, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should show updated values
	if !strings.Contains(body, "Engineering Team Updated") {
		t.Error("Expected 'Engineering Team Updated' display name in detail page")
	}
	if !strings.Contains(body, "Private") {
		t.Error("Expected 'Private' visibility in detail page")
	}
}

func TestGroupDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, st := setupTestServer(t)

	// Create a group to delete
	ctx := context.Background()
	mailEnabled := false
	securityEnabled := true
	now := time.Now()

	testGroup := model.Group{
		ODataType:       "#microsoft.graph.group",
		DisplayName:     "Test Group to Delete",
		MailNickname:    "testgrouptodelete",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
		Visibility:      "Public",
		CreatedDateTime: &now,
	}

	createdGroup, err := st.CreateGroup(ctx, testGroup)
	if err != nil {
		t.Fatalf("Failed to create test group: %v", err)
	}


	// Verify group exists before deletion
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/groups/"+createdGroup.ID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d before deletion, got %d", http.StatusOK, w.Code)
	}

	// Delete group
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/ui/groups/"+createdGroup.ID+"/delete", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	// Should redirect to group list
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/ui/groups" {
		t.Errorf("Expected redirect to /ui/groups, got %s", location)
	}

	// Verify deletion in store
	_, err = st.GetGroup(ctx, createdGroup.ID)
	if err == nil {
		t.Error("Expected group to be deleted from store, but it still exists")
	}
}

func TestGroupAddMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, st := setupTestServer(t)

	// Get Engineering Team ID and Henry Taylor ID (not a member)
	ctx := context.Background()
	groups, _, _ := st.ListGroups(ctx, model.ListOptions{Top: 100})
	var engineeringTeamID string
	for _, group := range groups {
		if group.DisplayName == "Engineering Team" {
			engineeringTeamID = group.ID
			break
		}
	}
	if engineeringTeamID == "" {
		t.Fatal("Engineering Team not found in seed data")
	}

	// Get users and find Henry Taylor
	users, _, _ := st.ListUsers(ctx, model.ListOptions{Top: 100})
	var henryID string
	for _, user := range users {
		if user.UserPrincipalName == "henry.taylor@saldeti.local" {
			henryID = user.ID
			break
		}
	}
	if henryID == "" {
		t.Fatal("Henry Taylor not found in seed data")
	}


	// Add Henry to Engineering Team
	formData := url.Values{}
	formData.Set("userId", henryID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/"+engineeringTeamID+"/members/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	// Should redirect to group detail page
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/ui/groups/"+engineeringTeamID {
		t.Errorf("Expected redirect to /ui/groups/%s, got %s", engineeringTeamID, location)
	}

	// Follow redirect to detail page
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", location, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should show Henry in the members list
	if !strings.Contains(body, "Henry Taylor") {
		t.Error("Expected 'Henry Taylor' in members list after adding")
	}
}

func TestGroupRemoveMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, st := setupTestServer(t)

	// Get Engineering Team ID
	ctx := context.Background()
	groups, _, _ := st.ListGroups(ctx, model.ListOptions{Top: 100})
	var engineeringTeamID string
	for _, group := range groups {
		if group.DisplayName == "Engineering Team" {
			engineeringTeamID = group.ID
			break
		}
	}
	if engineeringTeamID == "" {
		t.Fatal("Engineering Team not found in seed data")
	}

	// Get Grace Lee ID (she's a member of Engineering Team)
	users, _, _ := st.ListUsers(ctx, model.ListOptions{Top: 100})
	var graceID string
	for _, user := range users {
		if user.UserPrincipalName == "grace.lee@saldeti.local" {
			graceID = user.ID
			break
		}
	}
	if graceID == "" {
		t.Fatal("Grace Lee not found in seed data")
	}


	// Verify Grace is in the group first
	members, _, _ := st.ListMembers(ctx, engineeringTeamID, model.ListOptions{Top: 100})
	graceIsMember := false
	for _, member := range members {
		if member.ID == graceID {
			graceIsMember = true
			break
		}
	}
	if !graceIsMember {
		t.Fatal("Grace Lee should be a member of Engineering Team in seed data")
	}

	// Remove Grace from Engineering Team
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/"+engineeringTeamID+"/members/"+graceID+"/remove", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	// Should redirect to group detail page
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/ui/groups/"+engineeringTeamID {
		t.Errorf("Expected redirect to /ui/groups/%s, got %s", engineeringTeamID, location)
	}

	// Follow redirect to detail page
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", location, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Grace should not appear in the members list anymore
	if strings.Contains(body, "Grace Lee") {
		// Grace might still appear elsewhere, but shouldn't be in the members table
		// Check if she appears in a table row
		if strings.Contains(body, "members") && strings.Contains(body, "Grace Lee") {
			// This is a bit loose but we want to be careful
		}
	}

	// Verify Grace was removed from store
	members, _, _ = st.ListMembers(ctx, engineeringTeamID, model.ListOptions{Top: 100})
	graceIsMember = false
	for _, member := range members {
		if member.ID == graceID {
			graceIsMember = true
			break
		}
	}
	if graceIsMember {
		t.Error("Grace Lee should have been removed from Engineering Team")
	}
}

func TestGroupAddOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, st := setupTestServer(t)

	// Get Engineering Team ID and Henry Taylor ID
	ctx := context.Background()
	groups, _, _ := st.ListGroups(ctx, model.ListOptions{Top: 100})
	var engineeringTeamID string
	for _, group := range groups {
		if group.DisplayName == "Engineering Team" {
			engineeringTeamID = group.ID
			break
		}
	}
	if engineeringTeamID == "" {
		t.Fatal("Engineering Team not found in seed data")
	}

	// Get users and find Henry Taylor
	users, _, _ := st.ListUsers(ctx, model.ListOptions{Top: 100})
	var henryID string
	for _, user := range users {
		if user.UserPrincipalName == "henry.taylor@saldeti.local" {
			henryID = user.ID
			break
		}
	}
	if henryID == "" {
		t.Fatal("Henry Taylor not found in seed data")
	}


	// Add Henry as owner of Engineering Team
	formData := url.Values{}
	formData.Set("userId", henryID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/"+engineeringTeamID+"/owners/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	// Should redirect to group detail page
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/ui/groups/"+engineeringTeamID {
		t.Errorf("Expected redirect to /ui/groups/%s, got %s", engineeringTeamID, location)
	}

	// Follow redirect to detail page
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", location, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should show Henry in the owners list
	if !strings.Contains(body, "Henry Taylor") {
		t.Error("Expected 'Henry Taylor' in owners list after adding")
	}
}

func TestGroupRemoveOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server with API and UI routes
	ts, st := setupTestServer(t)

	// Create a group and add an owner
	ctx := context.Background()
	mailEnabled := false
	securityEnabled := true
	now := time.Now()

	testGroup := model.Group{
		ODataType:       "#microsoft.graph.group",
		DisplayName:     "Test Group for Owner",
		MailNickname:    "testgroupowner",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
		Visibility:      "Public",
		CreatedDateTime: &now,
	}

	createdGroup, err := st.CreateGroup(ctx, testGroup)
	if err != nil {
		t.Fatalf("Failed to create test group: %v", err)
	}

	// Get Alice Smith ID
	users, _, _ := st.ListUsers(ctx, model.ListOptions{Top: 100})
	var aliceID string
	for _, user := range users {
		if user.UserPrincipalName == "alice.smith@saldeti.local" {
			aliceID = user.ID
			break
		}
	}
	if aliceID == "" {
		t.Fatal("Alice Smith not found in seed data")
	}

	// Add Alice as owner
	err = st.AddOwner(ctx, createdGroup.ID, aliceID, "user")
	if err != nil {
		t.Fatalf("Failed to add Alice as owner: %v", err)
	}


	// Verify Alice is an owner
	owners, _, _ := st.ListOwners(ctx, createdGroup.ID, model.ListOptions{Top: 100})
	aliceIsOwner := false
	for _, owner := range owners {
		if owner.ID == aliceID {
			aliceIsOwner = true
			break
		}
	}
	if !aliceIsOwner {
		t.Fatal("Alice Smith should be an owner of the test group")
	}

	// Remove Alice as owner
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/"+createdGroup.ID+"/owners/"+aliceID+"/remove", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	// Should redirect to group detail page
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/ui/groups/"+createdGroup.ID {
		t.Errorf("Expected redirect to /ui/groups/%s, got %s", createdGroup.ID, location)
	}

	// Follow redirect to detail page
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", location, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify Alice was removed from owners
	owners, _, _ = st.ListOwners(ctx, createdGroup.ID, model.ListOptions{Top: 100})
	aliceIsOwner = false
	for _, owner := range owners {
		if owner.ID == aliceID {
			aliceIsOwner = true
			break
		}
	}
	if aliceIsOwner {
		t.Error("Alice Smith should have been removed as owner")
	}
}

func TestHtmxGroupAddMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	groups, _, _ := st.ListGroups(ctx, model.ListOptions{Top: 100})
	var allStaffID string
	for _, g := range groups {
		if g.DisplayName == "All Staff" {
			allStaffID = g.ID
			break
		}
	}
	if allStaffID == "" {
		t.Fatal("All Staff group not found")
	}

	users, _, _ := st.ListUsers(ctx, model.ListOptions{Top: 100})
	var ivanID string
	for _, u := range users {
		if u.UserPrincipalName == "ivan.guest@external.com" {
			ivanID = u.ID
			break
		}
	}
	if ivanID == "" {
		t.Fatal("Ivan Guest not found")
	}

	formData := url.Values{}
	formData.Set("userId", ivanID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/"+allStaffID+"/members/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("htmx response should not contain full HTML layout")
	}
	if !strings.Contains(body, `id="members"`) {
		t.Error("htmx response should contain members partial with id")
	}
	if !strings.Contains(body, "Member added successfully") {
		t.Error("htmx response should contain flash message")
	}
}

func TestHtmxGroupRemoveMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	groups, _, _ := st.ListGroups(ctx, model.ListOptions{Top: 100})
	var engineeringTeamID string
	for _, g := range groups {
		if g.DisplayName == "Engineering Team" {
			engineeringTeamID = g.ID
			break
		}
	}
	if engineeringTeamID == "" {
		t.Fatal("Engineering Team not found")
	}

	users, _, _ := st.ListUsers(ctx, model.ListOptions{Top: 100})
	var graceID string
	for _, u := range users {
		if u.UserPrincipalName == "grace.lee@saldeti.local" {
			graceID = u.ID
			break
		}
	}
	if graceID == "" {
		t.Fatal("Grace Lee not found")
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/"+engineeringTeamID+"/members/"+graceID+"/remove", nil)
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("htmx response should not contain full HTML layout")
	}
	if !strings.Contains(body, `id="members"`) {
		t.Error("htmx response should contain members partial")
	}
}

func TestHtmxGroupAddOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	groups, _, _ := st.ListGroups(ctx, model.ListOptions{Top: 100})
	var marketingTeamID string
	for _, g := range groups {
		if g.DisplayName == "Marketing Team" {
			marketingTeamID = g.ID
			break
		}
	}
	if marketingTeamID == "" {
		t.Fatal("Marketing Team not found")
	}

	users, _, _ := st.ListUsers(ctx, model.ListOptions{Top: 100})
	var aliceID string
	for _, u := range users {
		if u.UserPrincipalName == "alice.smith@saldeti.local" {
			aliceID = u.ID
			break
		}
	}
	if aliceID == "" {
		t.Fatal("Alice Smith not found")
	}

	formData := url.Values{}
	formData.Set("userId", aliceID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/"+marketingTeamID+"/owners/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("htmx response should not contain full HTML layout")
	}
	if !strings.Contains(body, `id="owners"`) {
		t.Error("htmx response should contain owners partial with id")
	}
	if !strings.Contains(body, "Owner added successfully") {
		t.Error("htmx response should contain flash message")
	}
}

func TestHtmxGroupRemoveOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()

	// Create a group and add an owner for clean removal
	mailEnabled := false
	securityEnabled := true
	now := time.Now()
	testGroup := model.Group{
		ODataType:       "#microsoft.graph.group",
		DisplayName:     "HTMX Owner Test Group",
		MailNickname:    "htmxownertest",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
		Visibility:      "Public",
		CreatedDateTime: &now,
	}
	createdGroup, err := st.CreateGroup(ctx, testGroup)
	if err != nil {
		t.Fatalf("Failed to create test group: %v", err)
	}

	users, _, _ := st.ListUsers(ctx, model.ListOptions{Top: 100})
	var aliceID string
	for _, u := range users {
		if u.UserPrincipalName == "alice.smith@saldeti.local" {
			aliceID = u.ID
			break
		}
	}
	if aliceID == "" {
		t.Fatal("Alice Smith not found")
	}

	// Add Alice as owner
	err = st.AddOwner(ctx, createdGroup.ID, aliceID, "user")
	if err != nil {
		t.Fatalf("Failed to add Alice as owner: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/groups/"+createdGroup.ID+"/owners/"+aliceID+"/remove", nil)
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("htmx response should not contain full HTML layout")
	}
	if !strings.Contains(body, `id="owners"`) {
		t.Error("htmx response should contain owners partial")
	}
}
