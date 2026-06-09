package store

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/saldeti/saldeti/internal/google/model"
)

var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
)

type clientEntry struct {
	clientID     string
	clientSecret string
}

type memoryStore struct {
	mu sync.RWMutex

	clients map[string]clientEntry

	users        map[string]*model.User
	usersByEmail map[string]string

	userAliases  map[string]string
	userPhotos   map[string]*model.UserPhoto
	deletedUsers map[string]*model.User

	groups        map[string]*model.Group
	groupsByEmail map[string]string
	groupAliases  map[string]string
	members       map[string]map[string]*model.Member

	orgUnits map[string]map[string]*model.OrgUnit

	roles           map[string]map[string]*model.Role
	roleAssignments map[string]map[string]*model.RoleAssignment

	customers    map[string]*model.Customer
	domains      map[string]map[string]*model.Domain
	domainAliases map[string]map[string]*model.DomainAlias

	chromeDevices  map[string]map[string]*model.ChromeOSDevice
	mobileDevices  map[string]map[string]*model.MobileDevice
	chromeCommands map[string]map[string]*model.ChromeOSCommandResult

	ciDevices   map[string]*model.CloudIdentityDevice
	deviceUsers map[string][]*model.DeviceUser

	ciGroups      map[string]*model.CloudIdentityGroup
	ciMemberships map[string]map[string]*model.Membership

	activities   []model.Activity
	usageReports map[string][]model.UsageReport

	tokens            map[string]map[string]*model.Token
	asps              map[string]map[string]*model.ASP
	verificationCodes map[string][]model.VerificationCode
	userInvitations   map[string]*model.UserInvitation

	schemas           map[string]map[string]*model.Schema
	calendarResources map[string]map[string]*model.CalendarResource
	buildings         map[string]map[string]*model.Building
	features          map[string]map[string]*model.Feature
	groupSettingsMap  map[string]*model.GroupSettings
	transferApps      []model.TransferApplication
	dataTransfers     map[string]*model.DataTransfer
	subscriptions     map[string]*model.Subscription
}

// MemoryStore implements Store with in-memory storage.
type MemoryStore struct {
	store *memoryStore
}

// NewMemoryStore creates a new in-memory store for Google Workspace data.
func NewMemoryStore() *MemoryStore {
	ms := &memoryStore{
		clients:           make(map[string]clientEntry),
		users:             make(map[string]*model.User),
		usersByEmail:      make(map[string]string),
		userAliases:       make(map[string]string),
		userPhotos:        make(map[string]*model.UserPhoto),
		deletedUsers:      make(map[string]*model.User),
		groups:            make(map[string]*model.Group),
		groupsByEmail:     make(map[string]string),
		groupAliases:      make(map[string]string),
		members:           make(map[string]map[string]*model.Member),
		orgUnits:          make(map[string]map[string]*model.OrgUnit),
		roles:             make(map[string]map[string]*model.Role),
		roleAssignments:   make(map[string]map[string]*model.RoleAssignment),
		customers:         make(map[string]*model.Customer),
		domains:           make(map[string]map[string]*model.Domain),
		domainAliases:     make(map[string]map[string]*model.DomainAlias),
		chromeDevices:     make(map[string]map[string]*model.ChromeOSDevice),
		mobileDevices:     make(map[string]map[string]*model.MobileDevice),
		chromeCommands:    make(map[string]map[string]*model.ChromeOSCommandResult),
		ciDevices:         make(map[string]*model.CloudIdentityDevice),
		deviceUsers:       make(map[string][]*model.DeviceUser),
		ciGroups:          make(map[string]*model.CloudIdentityGroup),
		ciMemberships:     make(map[string]map[string]*model.Membership),
		usageReports:      make(map[string][]model.UsageReport),
		tokens:            make(map[string]map[string]*model.Token),
		asps:              make(map[string]map[string]*model.ASP),
		verificationCodes: make(map[string][]model.VerificationCode),
		userInvitations:   make(map[string]*model.UserInvitation),
		schemas:           make(map[string]map[string]*model.Schema),
		calendarResources: make(map[string]map[string]*model.CalendarResource),
		buildings:         make(map[string]map[string]*model.Building),
		features:          make(map[string]map[string]*model.Feature),
		groupSettingsMap:  make(map[string]*model.GroupSettings),
		dataTransfers:     make(map[string]*model.DataTransfer),
		subscriptions:     make(map[string]*model.Subscription),
	}
	ms.transferApps = []model.TransferApplication{
		{Kind: "admin#datatransfer#application", Id: "55656082996", Name: "Google Drive"},
		{Kind: "admin#datatransfer#application", Id: "google-mail", Name: "Gmail"},
		{Kind: "admin#datatransfer#application", Id: "google-calendar", Name: "Google Calendar"},
	}
	return &MemoryStore{store: ms}
}

// Pagination helpers

func encodePageToken(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodePageToken(token string) int {
	if token == "" {
		return 0
	}
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	offset, err := strconv.Atoi(string(b))
	if err != nil {
		return 0
	}
	return offset
}

// Generic patch helper using JSON round-trip
func applyPatchMap(data interface{}, patch map[string]interface{}) error {
	original, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var originalMap map[string]interface{}
	if err := json.Unmarshal(original, &originalMap); err != nil {
		return err
	}
	for k, v := range patch {
		originalMap[k] = v
	}
	merged, err := json.Marshal(originalMap)
	if err != nil {
		return err
	}
	return json.Unmarshal(merged, data)
}

// User key resolver
func (ms *memoryStore) resolveUserKey(userKey string) (*model.User, error) {
	if u, ok := ms.users[userKey]; ok {
		cp := *u
		return &cp, nil
	}
	if id, ok := ms.usersByEmail[userKey]; ok {
		if u, ok := ms.users[id]; ok {
			cp := *u
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (ms *memoryStore) resolveGroupKey(groupKey string) (*model.Group, error) {
	if g, ok := ms.groups[groupKey]; ok {
		cp := *g
		return &cp, nil
	}
	if id, ok := ms.groupsByEmail[groupKey]; ok {
		if g, ok := ms.groups[id]; ok {
			cp := *g
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func maxResultsOrDefault(mr int) int {
	if mr <= 0 || mr > 500 {
		return 100
	}
	return mr
}

// Ping verifies the store is accessible.
func (s *MemoryStore) Ping(ctx context.Context) error {
	return nil
}

// Auth

func (s *MemoryStore) GetClient(ctx context.Context, clientID string) (string, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	entry, ok := s.store.clients[clientID]
	if !ok {
		return "", ErrNotFound
	}
	return entry.clientSecret, nil
}

func (s *MemoryStore) RegisterClient(ctx context.Context, clientID, clientSecret string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if _, ok := s.store.clients[clientID]; ok {
		return ErrAlreadyExists
	}
	s.store.clients[clientID] = clientEntry{clientID: clientID, clientSecret: clientSecret}
	return nil
}

// Users

func (s *MemoryStore) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.PrimaryEmail != "" {
		if _, ok := s.store.usersByEmail[user.PrimaryEmail]; ok {
			return model.User{}, ErrAlreadyExists
		}
	}
	user.CreationTime = time.Now().Format(time.RFC3339)
	user.Kind = "admin#directory#user"
	s.store.users[user.ID] = &user
	if user.PrimaryEmail != "" {
		s.store.usersByEmail[user.PrimaryEmail] = user.ID
	}
	return user, nil
}

func (s *MemoryStore) GetUser(ctx context.Context, userKey string) (*model.User, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	return s.store.resolveUserKey(userKey)
}

func (s *MemoryStore) ListUsers(ctx context.Context, opts model.ListOptions) ([]model.User, string, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var all []model.User
	for _, u := range s.store.users {
		if opts.Query != "" {
			if !strings.Contains(strings.ToLower(u.PrimaryEmail), strings.ToLower(opts.Query)) &&
				!strings.Contains(strings.ToLower(u.DisplayName), strings.ToLower(opts.Query)) &&
				!strings.Contains(strings.ToLower(u.FamilyName), strings.ToLower(opts.Query)) &&
				!strings.Contains(strings.ToLower(u.GivenName), strings.ToLower(opts.Query)) {
				continue
			}
		}
		all = append(all, *u)
	}
	offset := decodePageToken(opts.PageToken)
	mr := maxResultsOrDefault(opts.MaxResults)
	end := offset + mr
	if end > len(all) {
		end = len(all)
	}
	var nextPage string
	if end < len(all) {
		nextPage = encodePageToken(end)
	}
	return all[offset:end], nextPage, nil
}

func (s *MemoryStore) UpdateUser(ctx context.Context, userKey string, user model.User) (*model.User, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return nil, err
	}
	oldEmail := existing.PrimaryEmail
	id := existing.ID
	user.ID = id
	s.store.users[id] = &user
	if user.PrimaryEmail != oldEmail {
		delete(s.store.usersByEmail, oldEmail)
	}
	if user.PrimaryEmail != "" {
		s.store.usersByEmail[user.PrimaryEmail] = id
	}
	cp := user
	return &cp, nil
}

func (s *MemoryStore) PatchUser(ctx context.Context, userKey string, patch map[string]interface{}) (*model.User, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return nil, err
	}
	if err := applyPatchMap(existing, patch); err != nil {
		return nil, err
	}
	s.store.users[existing.ID] = existing
	if existing.PrimaryEmail != "" {
		s.store.usersByEmail[existing.PrimaryEmail] = existing.ID
	}
	return existing, nil
}

func (s *MemoryStore) DeleteUser(ctx context.Context, userKey string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return err
	}
	cp := *existing
	s.store.deletedUsers[existing.ID] = &cp
	delete(s.store.users, existing.ID)
	delete(s.store.usersByEmail, existing.PrimaryEmail)
	delete(s.store.userPhotos, existing.ID)
	return nil
}

func (s *MemoryStore) MakeAdmin(ctx context.Context, userKey string, isAdmin bool) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return err
	}
	existing.IsAdmin = isAdmin
	s.store.users[existing.ID] = existing
	return nil
}

func (s *MemoryStore) UndeleteUser(ctx context.Context, userKey string) (*model.User, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	// Try deletedUsers by ID
	if u, ok := s.store.deletedUsers[userKey]; ok {
		s.store.users[u.ID] = u
		if u.PrimaryEmail != "" {
			s.store.usersByEmail[u.PrimaryEmail] = u.ID
		}
		delete(s.store.deletedUsers, userKey)
		cp := *u
		return &cp, nil
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) SignOutUser(ctx context.Context, userKey string) error {
	return nil
}

// User Aliases

func (s *MemoryStore) AddUserAlias(ctx context.Context, userKey string, alias string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return err
	}
	s.store.userAliases[alias] = existing.ID
	existing.Aliases = append(existing.Aliases, alias)
	s.store.users[existing.ID] = existing
	return nil
}

func (s *MemoryStore) ListUserAliases(ctx context.Context, userKey string) ([]string, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return nil, err
	}
	if existing.Aliases == nil {
		return []string{}, nil
	}
	return existing.Aliases, nil
}

func (s *MemoryStore) RemoveUserAlias(ctx context.Context, userKey string, alias string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return err
	}
	delete(s.store.userAliases, alias)
	filtered := existing.Aliases[:0]
	for _, a := range existing.Aliases {
		if a != alias {
			filtered = append(filtered, a)
		}
	}
	existing.Aliases = filtered
	s.store.users[existing.ID] = existing
	return nil
}

// User Photos

func (s *MemoryStore) GetUserPhoto(ctx context.Context, userKey string) (*model.UserPhoto, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return nil, err
	}
	photo, ok := s.store.userPhotos[existing.ID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *photo
	return &cp, nil
}

func (s *MemoryStore) UpdateUserPhoto(ctx context.Context, userKey string, photo model.UserPhoto) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return err
	}
	photo.PrimaryEmail = existing.PrimaryEmail
	s.store.userPhotos[existing.ID] = &photo
	return nil
}

func (s *MemoryStore) DeleteUserPhoto(ctx context.Context, userKey string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return err
	}
	delete(s.store.userPhotos, existing.ID)
	return nil
}

// Groups

func (s *MemoryStore) CreateGroup(ctx context.Context, group model.Group) (model.Group, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if group.ID == "" {
		group.ID = uuid.New().String()
	}
	if group.Email != "" {
		if _, ok := s.store.groupsByEmail[group.Email]; ok {
			return model.Group{}, ErrAlreadyExists
		}
	}
	group.Kind = "admin#directory#group"
	s.store.groups[group.ID] = &group
	if group.Email != "" {
		s.store.groupsByEmail[group.Email] = group.ID
	}
	s.store.members[group.ID] = make(map[string]*model.Member)
	return group, nil
}

func (s *MemoryStore) GetGroup(ctx context.Context, groupKey string) (*model.Group, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	return s.store.resolveGroupKey(groupKey)
}

func (s *MemoryStore) ListGroups(ctx context.Context, opts model.ListOptions) ([]model.Group, string, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var all []model.Group
	for _, g := range s.store.groups {
		if opts.Query != "" {
			if !strings.Contains(strings.ToLower(g.Email), strings.ToLower(opts.Query)) &&
				!strings.Contains(strings.ToLower(g.Name), strings.ToLower(opts.Query)) {
				continue
			}
		}
		all = append(all, *g)
	}
	offset := decodePageToken(opts.PageToken)
	mr := maxResultsOrDefault(opts.MaxResults)
	end := offset + mr
	if end > len(all) {
		end = len(all)
	}
	var nextPage string
	if end < len(all) {
		nextPage = encodePageToken(end)
	}
	return all[offset:end], nextPage, nil
}

func (s *MemoryStore) UpdateGroup(ctx context.Context, groupKey string, group model.Group) (*model.Group, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return nil, err
	}
	oldEmail := existing.Email
	id := existing.ID
	group.ID = id
	s.store.groups[id] = &group
	if group.Email != oldEmail {
		delete(s.store.groupsByEmail, oldEmail)
	}
	if group.Email != "" {
		s.store.groupsByEmail[group.Email] = id
	}
	cp := group
	return &cp, nil
}

func (s *MemoryStore) PatchGroup(ctx context.Context, groupKey string, patch map[string]interface{}) (*model.Group, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return nil, err
	}
	if err := applyPatchMap(existing, patch); err != nil {
		return nil, err
	}
	s.store.groups[existing.ID] = existing
	if existing.Email != "" {
		s.store.groupsByEmail[existing.Email] = existing.ID
	}
	return existing, nil
}

func (s *MemoryStore) DeleteGroup(ctx context.Context, groupKey string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return err
	}
	delete(s.store.groups, existing.ID)
	delete(s.store.groupsByEmail, existing.Email)
	delete(s.store.members, existing.ID)
	delete(s.store.groupSettingsMap, existing.Email)
	return nil
}

// Group Aliases

func (s *MemoryStore) AddGroupAlias(ctx context.Context, groupKey string, alias string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return err
	}
	s.store.groupAliases[alias] = existing.ID
	existing.Aliases = append(existing.Aliases, alias)
	s.store.groups[existing.ID] = existing
	return nil
}

func (s *MemoryStore) ListGroupAliases(ctx context.Context, groupKey string) ([]string, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return nil, err
	}
	if existing.Aliases == nil {
		return []string{}, nil
	}
	return existing.Aliases, nil
}

func (s *MemoryStore) RemoveGroupAlias(ctx context.Context, groupKey string, alias string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return err
	}
	delete(s.store.groupAliases, alias)
	filtered := existing.Aliases[:0]
	for _, a := range existing.Aliases {
		if a != alias {
			filtered = append(filtered, a)
		}
	}
	existing.Aliases = filtered
	s.store.groups[existing.ID] = existing
	return nil
}

// Members

func (s *MemoryStore) ListMembers(ctx context.Context, groupKey string, opts model.ListOptions) ([]model.Member, string, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return nil, "", err
	}
	mMap, ok := s.store.members[existing.ID]
	if !ok {
		return []model.Member{}, "", nil
	}
	var all []model.Member
	for _, m := range mMap {
		if opts.Query != "" {
			if !strings.Contains(strings.ToLower(m.Email), strings.ToLower(opts.Query)) &&
				!strings.Contains(strings.ToLower(m.Role), strings.ToLower(opts.Query)) {
				continue
			}
		}
		all = append(all, *m)
	}
	offset := decodePageToken(opts.PageToken)
	mr := maxResultsOrDefault(opts.MaxResults)
	end := offset + mr
	if end > len(all) {
		end = len(all)
	}
	var nextPage string
	if end < len(all) {
		nextPage = encodePageToken(end)
	}
	return all[offset:end], nextPage, nil
}

func (s *MemoryStore) GetMember(ctx context.Context, groupKey, memberKey string) (*model.Member, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return nil, err
	}
	mMap, ok := s.store.members[existing.ID]
	if !ok {
		return nil, ErrNotFound
	}
	m, ok := mMap[memberKey]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *m
	return &cp, nil
}

func (s *MemoryStore) AddMember(ctx context.Context, groupKey string, member model.Member) (model.Member, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return model.Member{}, err
	}
	if member.ID == "" {
		member.ID = uuid.New().String()
	}
	member.Kind = "admin#directory#member"
	if s.store.members[existing.ID] == nil {
		s.store.members[existing.ID] = make(map[string]*model.Member)
	}
	s.store.members[existing.ID][member.Email] = &member
	return member, nil
}

func (s *MemoryStore) UpdateMember(ctx context.Context, groupKey, memberKey string, member model.Member) (*model.Member, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return nil, err
	}
	mMap, ok := s.store.members[existing.ID]
	if !ok {
		return nil, ErrNotFound
	}
	old, ok := mMap[memberKey]
	if !ok {
		return nil, ErrNotFound
	}
	member.ID = old.ID
	delete(mMap, memberKey)
	mMap[member.Email] = &member
	cp := member
	return &cp, nil
}

func (s *MemoryStore) RemoveMember(ctx context.Context, groupKey, memberKey string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return err
	}
	mMap, ok := s.store.members[existing.ID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := mMap[memberKey]; !ok {
		return ErrNotFound
	}
	delete(mMap, memberKey)
	return nil
}

func (s *MemoryStore) HasMember(ctx context.Context, groupKey, memberKey string) (bool, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	existing, err := s.store.resolveGroupKey(groupKey)
	if err != nil {
		return false, err
	}
	mMap, ok := s.store.members[existing.ID]
	if !ok {
		return false, nil
	}
	_, found := mMap[memberKey]
	return found, nil
}

// OrgUnits

func (s *MemoryStore) ListOrgUnits(ctx context.Context, customerID string) ([]model.OrgUnit, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	ouMap, ok := s.store.orgUnits[customerID]
	if !ok {
		return []model.OrgUnit{}, nil
	}
	var all []model.OrgUnit
	for _, ou := range ouMap {
		all = append(all, *ou)
	}
	return all, nil
}

func (s *MemoryStore) GetOrgUnit(ctx context.Context, customerID, orgUnitPath string) (*model.OrgUnit, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	ouMap, ok := s.store.orgUnits[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	ou, ok := ouMap[orgUnitPath]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *ou
	return &cp, nil
}

func (s *MemoryStore) CreateOrgUnit(ctx context.Context, customerID string, ou model.OrgUnit) (model.OrgUnit, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if ou.OrgUnitId == "" {
		ou.OrgUnitId = "id:" + uuid.New().String()
	}
	ou.Kind = "admin#directory#orgUnit"
	if s.store.orgUnits[customerID] == nil {
		s.store.orgUnits[customerID] = make(map[string]*model.OrgUnit)
	}
	s.store.orgUnits[customerID][ou.OrgUnitPath] = &ou
	return ou, nil
}

func (s *MemoryStore) UpdateOrgUnit(ctx context.Context, customerID, orgUnitPath string, ou model.OrgUnit) (*model.OrgUnit, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	ouMap, ok := s.store.orgUnits[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	old, ok := ouMap[orgUnitPath]
	if !ok {
		return nil, ErrNotFound
	}
	ou.OrgUnitId = old.OrgUnitId
	delete(ouMap, orgUnitPath)
	ouMap[ou.OrgUnitPath] = &ou
	cp := ou
	return &cp, nil
}

func (s *MemoryStore) PatchOrgUnit(ctx context.Context, customerID, orgUnitPath string, patch map[string]interface{}) (*model.OrgUnit, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	ouMap, ok := s.store.orgUnits[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	ou, ok := ouMap[orgUnitPath]
	if !ok {
		return nil, ErrNotFound
	}
	if err := applyPatchMap(ou, patch); err != nil {
		return nil, err
	}
	delete(ouMap, orgUnitPath)
	ouMap[ou.OrgUnitPath] = ou
	return ou, nil
}

func (s *MemoryStore) DeleteOrgUnit(ctx context.Context, customerID, orgUnitPath string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	ouMap, ok := s.store.orgUnits[customerID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := ouMap[orgUnitPath]; !ok {
		return ErrNotFound
	}
	delete(ouMap, orgUnitPath)
	return nil
}

// Roles

func (s *MemoryStore) ListRoles(ctx context.Context, customerID string) ([]model.Role, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	rMap, ok := s.store.roles[customerID]
	if !ok {
		return []model.Role{}, nil
	}
	var all []model.Role
	for _, r := range rMap {
		all = append(all, *r)
	}
	return all, nil
}

func (s *MemoryStore) GetRole(ctx context.Context, customerID, roleID string) (*model.Role, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	rMap, ok := s.store.roles[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	r, ok := rMap[roleID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *r
	return &cp, nil
}

func (s *MemoryStore) CreateRole(ctx context.Context, customerID string, role model.Role) (model.Role, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if role.RoleId == "" {
		role.RoleId = uuid.New().String()
	}
	if s.store.roles[customerID] == nil {
		s.store.roles[customerID] = make(map[string]*model.Role)
	}
	s.store.roles[customerID][role.RoleId] = &role
	return role, nil
}

func (s *MemoryStore) UpdateRole(ctx context.Context, customerID, roleID string, role model.Role) (*model.Role, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	rMap, ok := s.store.roles[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	if _, ok := rMap[roleID]; !ok {
		return nil, ErrNotFound
	}
	role.RoleId = roleID
	rMap[roleID] = &role
	cp := role
	return &cp, nil
}

func (s *MemoryStore) PatchRole(ctx context.Context, customerID, roleID string, patch map[string]interface{}) (*model.Role, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	rMap, ok := s.store.roles[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	r, ok := rMap[roleID]
	if !ok {
		return nil, ErrNotFound
	}
	if err := applyPatchMap(r, patch); err != nil {
		return nil, err
	}
	rMap[roleID] = r
	return r, nil
}

func (s *MemoryStore) DeleteRole(ctx context.Context, customerID, roleID string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	rMap, ok := s.store.roles[customerID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := rMap[roleID]; !ok {
		return ErrNotFound
	}
	delete(rMap, roleID)
	return nil
}

// Role Assignments

func (s *MemoryStore) ListRoleAssignments(ctx context.Context, customerID string) ([]model.RoleAssignment, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	raMap, ok := s.store.roleAssignments[customerID]
	if !ok {
		return []model.RoleAssignment{}, nil
	}
	var all []model.RoleAssignment
	for _, ra := range raMap {
		all = append(all, *ra)
	}
	return all, nil
}

func (s *MemoryStore) GetRoleAssignment(ctx context.Context, customerID, assignmentID string) (*model.RoleAssignment, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	raMap, ok := s.store.roleAssignments[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	ra, ok := raMap[assignmentID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *ra
	return &cp, nil
}

func (s *MemoryStore) CreateRoleAssignment(ctx context.Context, customerID string, ra model.RoleAssignment) (model.RoleAssignment, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if ra.RoleAssignmentId == "" {
		ra.RoleAssignmentId = uuid.New().String()
	}
	ra.Kind = "admin#directory#roleAssignment"
	if s.store.roleAssignments[customerID] == nil {
		s.store.roleAssignments[customerID] = make(map[string]*model.RoleAssignment)
	}
	s.store.roleAssignments[customerID][ra.RoleAssignmentId] = &ra
	return ra, nil
}

func (s *MemoryStore) DeleteRoleAssignment(ctx context.Context, customerID, assignmentID string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	raMap, ok := s.store.roleAssignments[customerID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := raMap[assignmentID]; !ok {
		return ErrNotFound
	}
	delete(raMap, assignmentID)
	return nil
}

// Privileges

func (s *MemoryStore) ListPrivileges(ctx context.Context, customerID string) ([]model.Privilege, error) {
	return []model.Privilege{
		{Kind: "admin#directory#privilege", ServiceName: "admin", PrivilegeName: "ADMIN_CONSOLE_ALL", ServiceId: "serviceAdmin", IsOuScopable: true},
		{Kind: "admin#directory#privilege", ServiceName: "admin", PrivilegeName: "USERS_MANAGE", ServiceId: "serviceAdmin", IsOuScopable: true},
		{Kind: "admin#directory#privilege", ServiceName: "admin", PrivilegeName: "GROUPS_MANAGE", ServiceId: "serviceAdmin", IsOuScopable: true},
		{Kind: "admin#directory#privilege", ServiceName: "admin", PrivilegeName: "ORG_UNITS_MANAGE", ServiceId: "serviceAdmin", IsOuScopable: true},
		{Kind: "admin#directory#privilege", ServiceName: "admin", PrivilegeName: "DEVICES_MANAGE", ServiceId: "serviceAdmin", IsOuScopable: true},
		{Kind: "admin#directory#privilege", ServiceName: "admin", PrivilegeName: "LICENCES_MANAGE", ServiceId: "serviceAdmin", IsOuScopable: true},
		{Kind: "admin#directory#privilege", ServiceName: "admin", PrivilegeName: "REPORTS_MANAGE", ServiceId: "serviceAdmin", IsOuScopable: false},
	}, nil
}

// Customers

func (s *MemoryStore) GetCustomer(ctx context.Context, customerKey string) (*model.Customer, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	c, ok := s.store.customers[customerKey]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (s *MemoryStore) UpdateCustomer(ctx context.Context, customerKey string, customer model.Customer) (*model.Customer, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if _, ok := s.store.customers[customerKey]; !ok {
		return nil, ErrNotFound
	}
	customer.ID = customerKey
	s.store.customers[customerKey] = &customer
	cp := customer
	return &cp, nil
}

func (s *MemoryStore) PatchCustomer(ctx context.Context, customerKey string, patch map[string]interface{}) (*model.Customer, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	c, ok := s.store.customers[customerKey]
	if !ok {
		return nil, ErrNotFound
	}
	if err := applyPatchMap(c, patch); err != nil {
		return nil, err
	}
	s.store.customers[customerKey] = c
	return c, nil
}

// Domains

func (s *MemoryStore) ListDomains(ctx context.Context, customerID string) ([]model.Domain, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	dMap, ok := s.store.domains[customerID]
	if !ok {
		return []model.Domain{}, nil
	}
	var all []model.Domain
	for _, d := range dMap {
		all = append(all, *d)
	}
	return all, nil
}

func (s *MemoryStore) GetDomain(ctx context.Context, customerID, domainName string) (*model.Domain, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	dMap, ok := s.store.domains[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	d, ok := dMap[domainName]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (s *MemoryStore) AddDomain(ctx context.Context, customerID string, domain model.Domain) (model.Domain, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	domain.Kind = "admin#directory#domain"
	domain.CreationTime = time.Now().Format(time.RFC3339)
	if s.store.domains[customerID] == nil {
		s.store.domains[customerID] = make(map[string]*model.Domain)
	}
	s.store.domains[customerID][domain.DomainName] = &domain
	return domain, nil
}

func (s *MemoryStore) DeleteDomain(ctx context.Context, customerID, domainName string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	dMap, ok := s.store.domains[customerID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := dMap[domainName]; !ok {
		return ErrNotFound
	}
	delete(dMap, domainName)
	return nil
}

// Domain Aliases

func (s *MemoryStore) ListDomainAliases(ctx context.Context, customerID string) ([]model.DomainAlias, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	daMap, ok := s.store.domainAliases[customerID]
	if !ok {
		return []model.DomainAlias{}, nil
	}
	var all []model.DomainAlias
	for _, da := range daMap {
		all = append(all, *da)
	}
	return all, nil
}

func (s *MemoryStore) GetDomainAlias(ctx context.Context, customerID, aliasName string) (*model.DomainAlias, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	daMap, ok := s.store.domainAliases[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	da, ok := daMap[aliasName]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *da
	return &cp, nil
}

func (s *MemoryStore) CreateDomainAlias(ctx context.Context, customerID string, da model.DomainAlias) (model.DomainAlias, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	da.Kind = "admin#directory#domainAlias"
	if s.store.domainAliases[customerID] == nil {
		s.store.domainAliases[customerID] = make(map[string]*model.DomainAlias)
	}
	s.store.domainAliases[customerID][da.DomainAliasName] = &da
	return da, nil
}

func (s *MemoryStore) DeleteDomainAlias(ctx context.Context, customerID, aliasName string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	daMap, ok := s.store.domainAliases[customerID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := daMap[aliasName]; !ok {
		return ErrNotFound
	}
	delete(daMap, aliasName)
	return nil
}

// ChromeOS Devices

func (s *MemoryStore) ListChromeOSDevices(ctx context.Context, customerID string, opts model.ListOptions) ([]model.ChromeOSDevice, string, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	dMap, ok := s.store.chromeDevices[customerID]
	if !ok {
		return []model.ChromeOSDevice{}, "", nil
	}
	var all []model.ChromeOSDevice
	for _, d := range dMap {
		if opts.Query != "" {
			if !strings.Contains(strings.ToLower(d.SerialNumber), strings.ToLower(opts.Query)) &&
				!strings.Contains(strings.ToLower(d.AnnotatedUser), strings.ToLower(opts.Query)) &&
				!strings.Contains(strings.ToLower(d.Notes), strings.ToLower(opts.Query)) {
				continue
			}
		}
		all = append(all, *d)
	}
	offset := decodePageToken(opts.PageToken)
	mr := maxResultsOrDefault(opts.MaxResults)
	end := offset + mr
	if end > len(all) {
		end = len(all)
	}
	var nextPage string
	if end < len(all) {
		nextPage = encodePageToken(end)
	}
	return all[offset:end], nextPage, nil
}

func (s *MemoryStore) GetChromeOSDevice(ctx context.Context, customerID, deviceID string) (*model.ChromeOSDevice, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	dMap, ok := s.store.chromeDevices[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	d, ok := dMap[deviceID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (s *MemoryStore) PatchChromeOSDevice(ctx context.Context, customerID, deviceID string, patch map[string]interface{}) (*model.ChromeOSDevice, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	dMap, ok := s.store.chromeDevices[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	d, ok := dMap[deviceID]
	if !ok {
		return nil, ErrNotFound
	}
	if err := applyPatchMap(d, patch); err != nil {
		return nil, err
	}
	dMap[deviceID] = d
	return d, nil
}

func (s *MemoryStore) UpdateChromeOSDevice(ctx context.Context, customerID, deviceID string, device model.ChromeOSDevice) (*model.ChromeOSDevice, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if s.store.chromeDevices[customerID] == nil {
		return nil, ErrNotFound
	}
	dMap := s.store.chromeDevices[customerID]
	if _, ok := dMap[deviceID]; !ok {
		return nil, ErrNotFound
	}
	device.DeviceID = deviceID
	dMap[deviceID] = &device
	cp := device
	return &cp, nil
}

func (s *MemoryStore) MoveChromeOSDevices(ctx context.Context, customerID string, deviceIDs []string, orgUnitPath string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	dMap, ok := s.store.chromeDevices[customerID]
	if !ok {
		return ErrNotFound
	}
	for _, id := range deviceIDs {
		if d, ok := dMap[id]; ok {
			d.OrgUnitPath = orgUnitPath
		}
	}
	return nil
}

func (s *MemoryStore) BatchChangeChromeOSStatus(ctx context.Context, customerID string, deviceIDs []string, action string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	dMap, ok := s.store.chromeDevices[customerID]
	if !ok {
		return ErrNotFound
	}
	for _, id := range deviceIDs {
		if d, ok := dMap[id]; ok {
			d.Status = action
		}
	}
	return nil
}

func (s *MemoryStore) CountChromeOSDevices(ctx context.Context, customerID string) (int64, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	dMap, ok := s.store.chromeDevices[customerID]
	if !ok {
		return 0, nil
	}
	return int64(len(dMap)), nil
}

func (s *MemoryStore) IssueChromeOSCommand(ctx context.Context, customerID, deviceID string, cmd model.ChromeOSCommand) (model.ChromeOSCommandResult, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	result := model.ChromeOSCommandResult{
		CommandId:   uuid.New().String(),
		CommandType: cmd.CommandType,
		DeviceId:    deviceID,
		State:       "IN_PROGRESS",
		ExecuteTime: time.Now().Format(time.RFC3339),
	}
	if s.store.chromeCommands[customerID] == nil {
		s.store.chromeCommands[customerID] = make(map[string]*model.ChromeOSCommandResult)
	}
	s.store.chromeCommands[customerID][result.CommandId] = &result
	return result, nil
}

func (s *MemoryStore) GetChromeOSCommand(ctx context.Context, customerID, deviceID, commandID string) (*model.ChromeOSCommandResult, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	cMap, ok := s.store.chromeCommands[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	r, ok := cMap[commandID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *r
	return &cp, nil
}

// Mobile Devices

func (s *MemoryStore) ListMobileDevices(ctx context.Context, customerID string, opts model.ListOptions) ([]model.MobileDevice, string, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	dMap, ok := s.store.mobileDevices[customerID]
	if !ok {
		return []model.MobileDevice{}, "", nil
	}
	var all []model.MobileDevice
	for _, d := range dMap {
		if opts.Query != "" {
			if !strings.Contains(strings.ToLower(d.SerialNumber), strings.ToLower(opts.Query)) &&
				!strings.Contains(strings.ToLower(d.Model), strings.ToLower(opts.Query)) {
				continue
			}
		}
		all = append(all, *d)
	}
	offset := decodePageToken(opts.PageToken)
	mr := maxResultsOrDefault(opts.MaxResults)
	end := offset + mr
	if end > len(all) {
		end = len(all)
	}
	var nextPage string
	if end < len(all) {
		nextPage = encodePageToken(end)
	}
	return all[offset:end], nextPage, nil
}

func (s *MemoryStore) GetMobileDevice(ctx context.Context, customerID, resourceID string) (*model.MobileDevice, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	dMap, ok := s.store.mobileDevices[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	d, ok := dMap[resourceID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (s *MemoryStore) DeleteMobileDevice(ctx context.Context, customerID, resourceID string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	dMap, ok := s.store.mobileDevices[customerID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := dMap[resourceID]; !ok {
		return ErrNotFound
	}
	delete(dMap, resourceID)
	return nil
}

func (s *MemoryStore) MobileDeviceAction(ctx context.Context, customerID, resourceID string, action model.MobileDeviceAction) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	dMap, ok := s.store.mobileDevices[customerID]
	if !ok {
		return ErrNotFound
	}
	d, ok := dMap[resourceID]
	if !ok {
		return ErrNotFound
	}
	switch action.Action {
	case "admin_account_wipe":
		d.Status = "account_wiped"
	case "admin_remote_wipe":
		d.Status = "remote_wiped"
	case "approve":
		d.Status = "approved"
	case "block":
		d.Status = "blocked"
	case "cancel_remote_wipe_then_enable":
		d.Status = "active"
	default:
		d.Status = action.Action
	}
	return nil
}

// Cloud Identity Devices

func (s *MemoryStore) ListCIDevices(ctx context.Context, opts model.ListOptions) ([]model.CloudIdentityDevice, string, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var all []model.CloudIdentityDevice
	for _, d := range s.store.ciDevices {
		if opts.Query != "" {
			if !strings.Contains(strings.ToLower(d.DisplayName), strings.ToLower(opts.Query)) &&
				!strings.Contains(strings.ToLower(d.SerialNumber), strings.ToLower(opts.Query)) {
				continue
			}
		}
		all = append(all, *d)
	}
	offset := decodePageToken(opts.PageToken)
	mr := maxResultsOrDefault(opts.MaxResults)
	end := offset + mr
	if end > len(all) {
		end = len(all)
	}
	var nextPage string
	if end < len(all) {
		nextPage = encodePageToken(end)
	}
	return all[offset:end], nextPage, nil
}

func (s *MemoryStore) GetCIDevice(ctx context.Context, name string) (*model.CloudIdentityDevice, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	d, ok := s.store.ciDevices[name]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (s *MemoryStore) CreateCIDevice(ctx context.Context, device model.CloudIdentityDevice) (model.CloudIdentityDevice, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if device.Name == "" {
		device.Name = "devices/" + uuid.New().String()
	}
	now := time.Now().Format(time.RFC3339)
	device.CreateTime = now
	device.UpdateTime = now
	if device.State == "" {
		device.State = "ACTIVE"
	}
	s.store.ciDevices[device.Name] = &device
	return device, nil
}

func (s *MemoryStore) DeleteCIDevice(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if _, ok := s.store.ciDevices[name]; !ok {
		return ErrNotFound
	}
	delete(s.store.ciDevices, name)
	delete(s.store.deviceUsers, name)
	return nil
}

func (s *MemoryStore) WipeCIDevice(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	d, ok := s.store.ciDevices[name]
	if !ok {
		return ErrNotFound
	}
	d.State = "WIPED"
	return nil
}

func (s *MemoryStore) CancelWipeCIDevice(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	d, ok := s.store.ciDevices[name]
	if !ok {
		return ErrNotFound
	}
	d.State = "ACTIVE"
	return nil
}

// Device Users

func (s *MemoryStore) ListDeviceUsers(ctx context.Context, parent string) ([]model.DeviceUser, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	users, ok := s.store.deviceUsers[parent]
	if !ok {
		return []model.DeviceUser{}, nil
	}
	var result []model.DeviceUser
	for _, du := range users {
		result = append(result, *du)
	}
	return result, nil
}

func (s *MemoryStore) GetDeviceUser(ctx context.Context, name string) (*model.DeviceUser, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	for _, users := range s.store.deviceUsers {
		for _, du := range users {
			if du.Name == name {
				cp := *du
				return &cp, nil
			}
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) DeleteDeviceUser(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	for parent, users := range s.store.deviceUsers {
		for i, du := range users {
			if du.Name == name {
				s.store.deviceUsers[parent] = append(users[:i], users[i+1:]...)
				return nil
			}
		}
	}
	return ErrNotFound
}

func (s *MemoryStore) ApproveDeviceUser(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	for _, users := range s.store.deviceUsers {
		for _, du := range users {
			if du.Name == name {
				du.ManagementState = "APPROVED"
				return nil
			}
		}
	}
	return ErrNotFound
}

func (s *MemoryStore) BlockDeviceUser(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	for _, users := range s.store.deviceUsers {
		for _, du := range users {
			if du.Name == name {
				du.ManagementState = "BLOCKED"
				return nil
			}
		}
	}
	return ErrNotFound
}

func (s *MemoryStore) WipeDeviceUser(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	for _, users := range s.store.deviceUsers {
		for _, du := range users {
			if du.Name == name {
				du.CompromisedState = "WIPED"
				return nil
			}
		}
	}
	return ErrNotFound
}

func (s *MemoryStore) CancelWipeDeviceUser(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	for _, users := range s.store.deviceUsers {
		for _, du := range users {
			if du.Name == name {
				du.CompromisedState = ""
				return nil
			}
		}
	}
	return ErrNotFound
}

func (s *MemoryStore) LookupDeviceUser(ctx context.Context, parent string) ([]model.DeviceUser, error) {
	return s.ListDeviceUsers(ctx, parent)
}

// Cloud Identity Groups

func (s *MemoryStore) ListCIGroups(ctx context.Context, opts model.ListOptions) ([]model.CloudIdentityGroup, string, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var all []model.CloudIdentityGroup
	for _, g := range s.store.ciGroups {
		if opts.Query != "" {
			if !strings.Contains(strings.ToLower(g.DisplayName), strings.ToLower(opts.Query)) &&
				!strings.Contains(strings.ToLower(g.Description), strings.ToLower(opts.Query)) {
				continue
			}
		}
		all = append(all, *g)
	}
	offset := decodePageToken(opts.PageToken)
	mr := maxResultsOrDefault(opts.MaxResults)
	end := offset + mr
	if end > len(all) {
		end = len(all)
	}
	var nextPage string
	if end < len(all) {
		nextPage = encodePageToken(end)
	}
	return all[offset:end], nextPage, nil
}

func (s *MemoryStore) GetCIGroup(ctx context.Context, name string) (*model.CloudIdentityGroup, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	g, ok := s.store.ciGroups[name]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *g
	return &cp, nil
}

func (s *MemoryStore) CreateCIGroup(ctx context.Context, group model.CloudIdentityGroup) (model.CloudIdentityGroup, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if group.Name == "" {
		group.Name = "groups/" + uuid.New().String()
	}
	now := time.Now().Format(time.RFC3339)
	group.CreateTime = now
	group.UpdateTime = now
	s.store.ciGroups[group.Name] = &group
	s.store.ciMemberships[group.Name] = make(map[string]*model.Membership)
	return group, nil
}

func (s *MemoryStore) UpdateCIGroup(ctx context.Context, name string, group model.CloudIdentityGroup) (*model.CloudIdentityGroup, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if _, ok := s.store.ciGroups[name]; !ok {
		return nil, ErrNotFound
	}
	group.Name = name
	group.UpdateTime = time.Now().Format(time.RFC3339)
	s.store.ciGroups[name] = &group
	cp := group
	return &cp, nil
}

func (s *MemoryStore) DeleteCIGroup(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if _, ok := s.store.ciGroups[name]; !ok {
		return ErrNotFound
	}
	delete(s.store.ciGroups, name)
	delete(s.store.ciMemberships, name)
	return nil
}

func (s *MemoryStore) LookupCIGroup(ctx context.Context, key model.EntityKey) (*model.CloudIdentityGroup, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	for _, g := range s.store.ciGroups {
		if g.GroupKey != nil && g.GroupKey.ID == key.ID {
			cp := *g
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) SearchCIGroups(ctx context.Context, query string) ([]model.CloudIdentityGroup, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var results []model.CloudIdentityGroup
	lq := strings.ToLower(query)
	for _, g := range s.store.ciGroups {
		if strings.Contains(strings.ToLower(g.DisplayName), lq) ||
			strings.Contains(strings.ToLower(g.Description), lq) {
			results = append(results, *g)
		}
	}
	return results, nil
}

func (s *MemoryStore) GetCIGroupSecuritySettings(ctx context.Context, name string) (*model.SecuritySettings, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	if _, ok := s.store.ciGroups[name]; !ok {
		return nil, ErrNotFound
	}
	return &model.SecuritySettings{
		Name:                name + "/securitySettings",
		WhoCanJoin:          "ANYONE_CAN_JOIN",
		WhoCanViewMembership: "ALL_MEMBERS_CAN_VIEW",
		WhoCanDiscoverGroup: "ALL_IN_DOMAIN_CAN_DISCOVER",
	}, nil
}

func (s *MemoryStore) UpdateCIGroupSecuritySettings(ctx context.Context, name string, settings model.SecuritySettings) (*model.SecuritySettings, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if _, ok := s.store.ciGroups[name]; !ok {
		return nil, ErrNotFound
	}
	settings.Name = name + "/securitySettings"
	return &settings, nil
}

// Cloud Identity Memberships

func (s *MemoryStore) ListCIMemberships(ctx context.Context, parent string) ([]model.Membership, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	mMap, ok := s.store.ciMemberships[parent]
	if !ok {
		return []model.Membership{}, nil
	}
	var all []model.Membership
	for _, m := range mMap {
		all = append(all, *m)
	}
	return all, nil
}

func (s *MemoryStore) GetCIMembership(ctx context.Context, name string) (*model.Membership, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	for _, mMap := range s.store.ciMemberships {
		if m, ok := mMap[name]; ok {
			cp := *m
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) CreateCIMembership(ctx context.Context, parent string, membership model.Membership) (model.Membership, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if membership.Name == "" {
		membership.Name = parent + "/memberships/" + uuid.New().String()
	}
	now := time.Now().Format(time.RFC3339)
	membership.CreateTime = now
	membership.UpdateTime = now
	if s.store.ciMemberships[parent] == nil {
		s.store.ciMemberships[parent] = make(map[string]*model.Membership)
	}
	s.store.ciMemberships[parent][membership.Name] = &membership
	return membership, nil
}

func (s *MemoryStore) DeleteCIMembership(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	for parent, mMap := range s.store.ciMemberships {
		if _, ok := mMap[name]; ok {
			delete(mMap, name)
			_ = parent
			return nil
		}
	}
	return ErrNotFound
}

func (s *MemoryStore) LookupCIMembership(ctx context.Context, parent string, key model.EntityKey) (*model.Membership, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	mMap, ok := s.store.ciMemberships[parent]
	if !ok {
		return nil, ErrNotFound
	}
	for _, m := range mMap {
		if m.MemberKey != nil && m.MemberKey.ID == key.ID {
			cp := *m
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ModifyMembershipRoles(ctx context.Context, name string, roles model.ModifyMembershipRolesRequest) (*model.Membership, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	for _, mMap := range s.store.ciMemberships {
		if m, ok := mMap[name]; ok {
			// Remove roles
			if len(roles.RemoveRoles) > 0 {
				removeSet := make(map[string]bool)
				for _, r := range roles.RemoveRoles {
					removeSet[r] = true
				}
				var filtered []model.MembershipRole
				for _, r := range m.Roles {
					if !removeSet[r.Name] {
						filtered = append(filtered, r)
					}
				}
				m.Roles = filtered
			}
			// Add roles
			m.Roles = append(m.Roles, roles.AddRoles...)
			m.UpdateTime = time.Now().Format(time.RFC3339)
			cp := *m
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) CheckTransitiveMembership(ctx context.Context, parent string, key model.EntityKey) (bool, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	mMap, ok := s.store.ciMemberships[parent]
	if !ok {
		return false, nil
	}
	for _, m := range mMap {
		if m.MemberKey != nil && m.MemberKey.ID == key.ID {
			return true, nil
		}
	}
	return false, nil
}

func (s *MemoryStore) GetMembershipGraph(ctx context.Context, parent string, query string) (*model.MembershipGraph, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	mMap, ok := s.store.ciMemberships[parent]
	if !ok {
		return &model.MembershipGraph{}, nil
	}
	var adjList []model.MembershipAdjacency
	for _, m := range mMap {
		adj := model.MembershipAdjacency{
			Edge: &model.MembershipEdge{
				SourceMember: parent,
				TargetMember: m.Name,
			},
			Members: []model.Membership{*m},
		}
		adjList = append(adjList, adj)
	}
	return &model.MembershipGraph{AdjacencyList: adjList}, nil
}

func (s *MemoryStore) SearchTransitiveGroups(ctx context.Context, parent string, query string) ([]model.CloudIdentityGroup, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var results []model.CloudIdentityGroup
	visited := make(map[string]bool)
	var bfs func(groupName string)
	bfs = func(groupName string) {
		if visited[groupName] {
			return
		}
		visited[groupName] = true
		mMap, ok := s.store.ciMemberships[groupName]
		if !ok {
			return
		}
		for _, m := range mMap {
			if m.MemberKey != nil {
				// Check if member is a group
				for _, g := range s.store.ciGroups {
					if g.GroupKey != nil && g.GroupKey.ID == m.MemberKey.ID {
						if query == "" || strings.Contains(strings.ToLower(g.DisplayName), strings.ToLower(query)) {
							results = append(results, *g)
						}
						bfs(g.Name)
					}
				}
			}
		}
	}
	bfs(parent)
	return results, nil
}

func (s *MemoryStore) SearchTransitiveMemberships(ctx context.Context, parent string, query string) ([]model.Membership, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var results []model.Membership
	visited := make(map[string]bool)
	var collect func(groupName string)
	collect = func(groupName string) {
		if visited[groupName] {
			return
		}
		visited[groupName] = true
		mMap, ok := s.store.ciMemberships[groupName]
		if !ok {
			return
		}
		for _, m := range mMap {
			results = append(results, *m)
			// Recurse into nested groups
			if m.MemberKey != nil {
				for _, g := range s.store.ciGroups {
					if g.GroupKey != nil && g.GroupKey.ID == m.MemberKey.ID {
						collect(g.Name)
					}
				}
			}
		}
	}
	collect(parent)
	return results, nil
}

func (s *MemoryStore) SearchDirectGroups(ctx context.Context, parent string, query string) ([]model.Membership, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var results []model.Membership
	// Find memberships where the parent group contains the entity described by query
	for groupName, mMap := range s.store.ciMemberships {
		if !strings.Contains(strings.ToLower(groupName), strings.ToLower(query)) {
			continue
		}
		for _, m := range mMap {
			results = append(results, *m)
		}
	}
	return results, nil
}

// Reports

func (s *MemoryStore) ListActivities(ctx context.Context, userKey, applicationName string) ([]model.Activity, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var results []model.Activity
	for _, a := range s.store.activities {
		if userKey != "" {
			if a.Actor.Email != userKey {
				continue
			}
		}
		if applicationName != "" {
			if a.Id == nil || a.Id.ApplicationName != applicationName {
				continue
			}
		}
		results = append(results, a)
	}
	return results, nil
}

func (s *MemoryStore) ListUsageReports(ctx context.Context, date, userKey, entityType, entityKey string) ([]model.UsageReport, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var results []model.UsageReport
	for _, reports := range s.store.usageReports {
		for _, r := range reports {
			if date != "" && r.Date != date {
				continue
			}
			if userKey != "" && r.Entity != nil && r.Entity.UserEmail != userKey {
				continue
			}
			if entityType != "" && r.Entity != nil && r.Entity.Type != entityType {
				continue
			}
			if entityKey != "" && r.Entity != nil && r.Entity.EntityId != entityKey {
				continue
			}
			results = append(results, r)
		}
	}
	return results, nil
}

// Security - Tokens

func (s *MemoryStore) ListTokens(ctx context.Context, userKey string) ([]model.Token, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	tMap, ok := s.store.tokens[userKey]
	if !ok {
		return []model.Token{}, nil
	}
	var all []model.Token
	for _, t := range tMap {
		all = append(all, *t)
	}
	return all, nil
}

func (s *MemoryStore) GetToken(ctx context.Context, userKey, clientID string) (*model.Token, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	tMap, ok := s.store.tokens[userKey]
	if !ok {
		return nil, ErrNotFound
	}
	t, ok := tMap[clientID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (s *MemoryStore) DeleteToken(ctx context.Context, userKey, clientID string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	tMap, ok := s.store.tokens[userKey]
	if !ok {
		return ErrNotFound
	}
	if _, ok := tMap[clientID]; !ok {
		return ErrNotFound
	}
	delete(tMap, clientID)
	return nil
}

// Security - ASPs

func (s *MemoryStore) ListASPs(ctx context.Context, userKey string) ([]model.ASP, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	aMap, ok := s.store.asps[userKey]
	if !ok {
		return []model.ASP{}, nil
	}
	var all []model.ASP
	for _, a := range aMap {
		all = append(all, *a)
	}
	return all, nil
}

func (s *MemoryStore) GetASP(ctx context.Context, userKey, codeID string) (*model.ASP, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	aMap, ok := s.store.asps[userKey]
	if !ok {
		return nil, ErrNotFound
	}
	a, ok := aMap[codeID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (s *MemoryStore) DeleteASP(ctx context.Context, userKey, codeID string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	aMap, ok := s.store.asps[userKey]
	if !ok {
		return ErrNotFound
	}
	if _, ok := aMap[codeID]; !ok {
		return ErrNotFound
	}
	delete(aMap, codeID)
	return nil
}

// Security - Verification Codes

func (s *MemoryStore) ListVerificationCodes(ctx context.Context, userKey string) ([]model.VerificationCode, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	codes, ok := s.store.verificationCodes[userKey]
	if !ok {
		return []model.VerificationCode{}, nil
	}
	return codes, nil
}

func (s *MemoryStore) GenerateVerificationCodes(ctx context.Context, userKey string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	code := model.VerificationCode{
		Kind:                  "admin#directory#verificationCode",
		UserId:                userKey,
		VerificationCode:      fmt.Sprintf("%06d", rand.Intn(1000000)),
		VerificationMethod:    "sms",
		VerificationTimestamp: time.Now().Format(time.RFC3339),
	}
	s.store.verificationCodes[userKey] = append(s.store.verificationCodes[userKey], code)
	return nil
}

func (s *MemoryStore) InvalidateVerificationCodes(ctx context.Context, userKey string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	s.store.verificationCodes[userKey] = nil
	return nil
}

func (s *MemoryStore) TurnOff2SV(ctx context.Context, userKey string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	existing, err := s.store.resolveUserKey(userKey)
	if err != nil {
		return err
	}
	existing.IsEnrolledIn2Sv = false
	existing.IsEnforcedIn2Sv = false
	existing.Is2svEnrolled = false
	s.store.users[existing.ID] = existing
	return nil
}

// User Invitations

func (s *MemoryStore) ListUserInvitations(ctx context.Context, parent string) ([]model.UserInvitation, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var results []model.UserInvitation
	for _, inv := range s.store.userInvitations {
		if strings.HasPrefix(inv.Name, parent) {
			results = append(results, *inv)
		}
	}
	return results, nil
}

func (s *MemoryStore) GetUserInvitation(ctx context.Context, name string) (*model.UserInvitation, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	inv, ok := s.store.userInvitations[name]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *inv
	return &cp, nil
}

func (s *MemoryStore) IsInvitableUser(ctx context.Context, name string) (bool, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	_, ok := s.store.userInvitations[name]
	return ok, nil
}

func (s *MemoryStore) SendUserInvitation(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	inv, ok := s.store.userInvitations[name]
	if !ok {
		return ErrNotFound
	}
	inv.State = "SENT"
	s.store.userInvitations[name] = inv
	return nil
}

func (s *MemoryStore) CancelUserInvitation(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	inv, ok := s.store.userInvitations[name]
	if !ok {
		return ErrNotFound
	}
	inv.State = "CANCELLED"
	s.store.userInvitations[name] = inv
	return nil
}

// Custom Schemas

func (s *MemoryStore) ListSchemas(ctx context.Context, customerID string) ([]model.Schema, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	sMap, ok := s.store.schemas[customerID]
	if !ok {
		return []model.Schema{}, nil
	}
	var all []model.Schema
	for _, sc := range sMap {
		all = append(all, *sc)
	}
	return all, nil
}

func (s *MemoryStore) GetSchema(ctx context.Context, customerID, schemaKey string) (*model.Schema, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	sMap, ok := s.store.schemas[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	sc, ok := sMap[schemaKey]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *sc
	return &cp, nil
}

func (s *MemoryStore) CreateSchema(ctx context.Context, customerID string, schema model.Schema) (model.Schema, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if schema.SchemaId == "" {
		schema.SchemaId = uuid.New().String()
	}
	schema.Kind = "admin#directory#schema"
	if s.store.schemas[customerID] == nil {
		s.store.schemas[customerID] = make(map[string]*model.Schema)
	}
	s.store.schemas[customerID][schema.SchemaName] = &schema
	return schema, nil
}

func (s *MemoryStore) UpdateSchema(ctx context.Context, customerID, schemaKey string, schema model.Schema) (*model.Schema, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	sMap, ok := s.store.schemas[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	old, ok := sMap[schemaKey]
	if !ok {
		return nil, ErrNotFound
	}
	schema.SchemaId = old.SchemaId
	delete(sMap, schemaKey)
	sMap[schema.SchemaName] = &schema
	cp := schema
	return &cp, nil
}

func (s *MemoryStore) PatchSchema(ctx context.Context, customerID, schemaKey string, patch map[string]interface{}) (*model.Schema, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	sMap, ok := s.store.schemas[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	sc, ok := sMap[schemaKey]
	if !ok {
		return nil, ErrNotFound
	}
	if err := applyPatchMap(sc, patch); err != nil {
		return nil, err
	}
	delete(sMap, schemaKey)
	sMap[sc.SchemaName] = sc
	return sc, nil
}

func (s *MemoryStore) DeleteSchema(ctx context.Context, customerID, schemaKey string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	sMap, ok := s.store.schemas[customerID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := sMap[schemaKey]; !ok {
		return ErrNotFound
	}
	delete(sMap, schemaKey)
	return nil
}

// Calendar Resources

func (s *MemoryStore) ListCalendarResources(ctx context.Context, customerID string) ([]model.CalendarResource, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	rMap, ok := s.store.calendarResources[customerID]
	if !ok {
		return []model.CalendarResource{}, nil
	}
	var all []model.CalendarResource
	for _, r := range rMap {
		all = append(all, *r)
	}
	return all, nil
}

func (s *MemoryStore) GetCalendarResource(ctx context.Context, customerID, resourceID string) (*model.CalendarResource, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	rMap, ok := s.store.calendarResources[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	r, ok := rMap[resourceID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *r
	return &cp, nil
}

func (s *MemoryStore) CreateCalendarResource(ctx context.Context, customerID string, resource model.CalendarResource) (model.CalendarResource, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if resource.ResourceId == "" {
		resource.ResourceId = uuid.New().String()
	}
	resource.Kind = "admin#directory#resources#calendar#CalendarResource"
	if s.store.calendarResources[customerID] == nil {
		s.store.calendarResources[customerID] = make(map[string]*model.CalendarResource)
	}
	s.store.calendarResources[customerID][resource.ResourceId] = &resource
	return resource, nil
}

func (s *MemoryStore) UpdateCalendarResource(ctx context.Context, customerID, resourceID string, resource model.CalendarResource) (*model.CalendarResource, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	rMap, ok := s.store.calendarResources[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	if _, ok := rMap[resourceID]; !ok {
		return nil, ErrNotFound
	}
	resource.ResourceId = resourceID
	rMap[resourceID] = &resource
	cp := resource
	return &cp, nil
}

func (s *MemoryStore) PatchCalendarResource(ctx context.Context, customerID, resourceID string, patch map[string]interface{}) (*model.CalendarResource, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	rMap, ok := s.store.calendarResources[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	r, ok := rMap[resourceID]
	if !ok {
		return nil, ErrNotFound
	}
	if err := applyPatchMap(r, patch); err != nil {
		return nil, err
	}
	rMap[resourceID] = r
	return r, nil
}

func (s *MemoryStore) DeleteCalendarResource(ctx context.Context, customerID, resourceID string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	rMap, ok := s.store.calendarResources[customerID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := rMap[resourceID]; !ok {
		return ErrNotFound
	}
	delete(rMap, resourceID)
	return nil
}

// Buildings

func (s *MemoryStore) ListBuildings(ctx context.Context, customerID string) ([]model.Building, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	bMap, ok := s.store.buildings[customerID]
	if !ok {
		return []model.Building{}, nil
	}
	var all []model.Building
	for _, b := range bMap {
		all = append(all, *b)
	}
	return all, nil
}

func (s *MemoryStore) GetBuilding(ctx context.Context, customerID, buildingID string) (*model.Building, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	bMap, ok := s.store.buildings[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	b, ok := bMap[buildingID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *b
	return &cp, nil
}

func (s *MemoryStore) CreateBuilding(ctx context.Context, customerID string, building model.Building) (model.Building, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if building.BuildingId == "" {
		building.BuildingId = uuid.New().String()
	}
	building.Kind = "admin#directory#resources#buildings#Building"
	if s.store.buildings[customerID] == nil {
		s.store.buildings[customerID] = make(map[string]*model.Building)
	}
	s.store.buildings[customerID][building.BuildingId] = &building
	return building, nil
}

func (s *MemoryStore) UpdateBuilding(ctx context.Context, customerID, buildingID string, building model.Building) (*model.Building, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	bMap, ok := s.store.buildings[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	if _, ok := bMap[buildingID]; !ok {
		return nil, ErrNotFound
	}
	building.BuildingId = buildingID
	bMap[buildingID] = &building
	cp := building
	return &cp, nil
}

func (s *MemoryStore) PatchBuilding(ctx context.Context, customerID, buildingID string, patch map[string]interface{}) (*model.Building, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	bMap, ok := s.store.buildings[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	b, ok := bMap[buildingID]
	if !ok {
		return nil, ErrNotFound
	}
	if err := applyPatchMap(b, patch); err != nil {
		return nil, err
	}
	bMap[buildingID] = b
	return b, nil
}

func (s *MemoryStore) DeleteBuilding(ctx context.Context, customerID, buildingID string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	bMap, ok := s.store.buildings[customerID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := bMap[buildingID]; !ok {
		return ErrNotFound
	}
	delete(bMap, buildingID)
	return nil
}

// Features

func (s *MemoryStore) ListFeatures(ctx context.Context, customerID string) ([]model.Feature, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	fMap, ok := s.store.features[customerID]
	if !ok {
		return []model.Feature{}, nil
	}
	var all []model.Feature
	for _, f := range fMap {
		all = append(all, *f)
	}
	return all, nil
}

func (s *MemoryStore) GetFeature(ctx context.Context, customerID, featureKey string) (*model.Feature, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	fMap, ok := s.store.features[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	f, ok := fMap[featureKey]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *f
	return &cp, nil
}

func (s *MemoryStore) CreateFeature(ctx context.Context, customerID string, feature model.Feature) (model.Feature, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	feature.Kind = "admin#directory#resources#features#Feature"
	if s.store.features[customerID] == nil {
		s.store.features[customerID] = make(map[string]*model.Feature)
	}
	s.store.features[customerID][feature.FeatureName] = &feature
	return feature, nil
}

func (s *MemoryStore) UpdateFeature(ctx context.Context, customerID, featureKey string, feature model.Feature) (*model.Feature, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	fMap, ok := s.store.features[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	if _, ok := fMap[featureKey]; !ok {
		return nil, ErrNotFound
	}
	delete(fMap, featureKey)
	fMap[feature.FeatureName] = &feature
	cp := feature
	return &cp, nil
}

func (s *MemoryStore) PatchFeature(ctx context.Context, customerID, featureKey string, patch map[string]interface{}) (*model.Feature, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	fMap, ok := s.store.features[customerID]
	if !ok {
		return nil, ErrNotFound
	}
	f, ok := fMap[featureKey]
	if !ok {
		return nil, ErrNotFound
	}
	if err := applyPatchMap(f, patch); err != nil {
		return nil, err
	}
	delete(fMap, featureKey)
	fMap[f.FeatureName] = f
	return f, nil
}

func (s *MemoryStore) DeleteFeature(ctx context.Context, customerID, featureKey string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	fMap, ok := s.store.features[customerID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := fMap[featureKey]; !ok {
		return ErrNotFound
	}
	delete(fMap, featureKey)
	return nil
}

func (s *MemoryStore) RenameFeature(ctx context.Context, customerID, oldName, newName string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	fMap, ok := s.store.features[customerID]
	if !ok {
		return ErrNotFound
	}
	f, ok := fMap[oldName]
	if !ok {
		return ErrNotFound
	}
	f.FeatureName = newName
	delete(fMap, oldName)
	fMap[newName] = f
	return nil
}

// Groups Settings

func (s *MemoryStore) GetGroupSettings(ctx context.Context, groupUniqueId string) (*model.GroupSettings, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	gs, ok := s.store.groupSettingsMap[groupUniqueId]
	if !ok {
		// Return default settings
		return &model.GroupSettings{
			Kind:  "groupsSettings#groups",
			Email: groupUniqueId,
		}, nil
	}
	cp := *gs
	return &cp, nil
}

func (s *MemoryStore) UpdateGroupSettings(ctx context.Context, groupUniqueId string, settings model.GroupSettings) (*model.GroupSettings, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	settings.Kind = "groupsSettings#groups"
	settings.Email = groupUniqueId
	s.store.groupSettingsMap[groupUniqueId] = &settings
	cp := settings
	return &cp, nil
}

func (s *MemoryStore) PatchGroupSettings(ctx context.Context, groupUniqueId string, patch map[string]interface{}) (*model.GroupSettings, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	gs, ok := s.store.groupSettingsMap[groupUniqueId]
	if !ok {
		gs = &model.GroupSettings{
			Kind:  "groupsSettings#groups",
			Email: groupUniqueId,
		}
	}
	if err := applyPatchMap(gs, patch); err != nil {
		return nil, err
	}
	s.store.groupSettingsMap[groupUniqueId] = gs
	return gs, nil
}

// Data Transfer

func (s *MemoryStore) ListTransferApplications(ctx context.Context) ([]model.TransferApplication, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	result := make([]model.TransferApplication, len(s.store.transferApps))
	copy(result, s.store.transferApps)
	return result, nil
}

func (s *MemoryStore) GetTransferApplication(ctx context.Context, applicationID string) (*model.TransferApplication, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	for _, app := range s.store.transferApps {
		if app.Id == applicationID {
			cp := app
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListTransfers(ctx context.Context) ([]model.DataTransfer, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var all []model.DataTransfer
	for _, dt := range s.store.dataTransfers {
		all = append(all, *dt)
	}
	return all, nil
}

func (s *MemoryStore) GetTransfer(ctx context.Context, transferID string) (*model.DataTransfer, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	dt, ok := s.store.dataTransfers[transferID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *dt
	return &cp, nil
}

func (s *MemoryStore) CreateTransfer(ctx context.Context, transfer model.DataTransfer) (model.DataTransfer, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if transfer.Id == "" {
		transfer.Id = uuid.New().String()
	}
	transfer.Kind = "admin#datatransfer#DataTransfer"
	transfer.RequestTime = time.Now().Format(time.RFC3339)
	transfer.OverallTransferStatusCode = "pending"
	s.store.dataTransfers[transfer.Id] = &transfer
	return transfer, nil
}

// Subscriptions

func (s *MemoryStore) ListSubscriptions(ctx context.Context) ([]model.Subscription, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	var all []model.Subscription
	for _, sub := range s.store.subscriptions {
		all = append(all, *sub)
	}
	return all, nil
}

func (s *MemoryStore) GetSubscription(ctx context.Context, name string) (*model.Subscription, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	sub, ok := s.store.subscriptions[name]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *sub
	return &cp, nil
}

func (s *MemoryStore) CreateSubscription(ctx context.Context, sub model.Subscription) (model.Subscription, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if sub.Name == "" {
		sub.Name = "subscriptions/" + uuid.New().String()
	}
	if sub.Uid == "" {
		sub.Uid = uuid.New().String()
	}
	now := time.Now().Format(time.RFC3339)
	sub.CreateTime = now
	sub.UpdateTime = now
	if sub.State == "" {
		sub.State = "ACTIVE"
	}
	s.store.subscriptions[sub.Name] = &sub
	return sub, nil
}

func (s *MemoryStore) UpdateSubscription(ctx context.Context, name string, sub model.Subscription) (*model.Subscription, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if _, ok := s.store.subscriptions[name]; !ok {
		return nil, ErrNotFound
	}
	sub.Name = name
	sub.UpdateTime = time.Now().Format(time.RFC3339)
	s.store.subscriptions[name] = &sub
	cp := sub
	return &cp, nil
}

func (s *MemoryStore) DeleteSubscription(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if _, ok := s.store.subscriptions[name]; !ok {
		return ErrNotFound
	}
	delete(s.store.subscriptions, name)
	return nil
}

func (s *MemoryStore) ReactivateSubscription(ctx context.Context, name string) error {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	sub, ok := s.store.subscriptions[name]
	if !ok {
		return ErrNotFound
	}
	sub.State = "ACTIVE"
	return nil
}
