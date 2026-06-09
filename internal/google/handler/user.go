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

// listUsersHandler handles GET /admin/directory/v1/users
func listUsersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		opts := parseGoogleListOptions(c)
		users, nextPageToken, err := st.ListUsers(c.Request.Context(), opts)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list users")
			return
		}
		if users == nil {
			users = []model.User{}
		}
		resp := gin.H{
			"kind":  "admin#directory#users",
			"etag":  "\"placeholder\"",
			"users": users,
		}
		if nextPageToken != "" {
			resp["nextPageToken"] = nextPageToken
		}
		writeJSON(c, http.StatusOK, resp)
	}
}

// getUserHandler handles GET /admin/directory/v1/users/:userKey
func getUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		user, err := st.GetUser(c.Request.Context(), userKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get user")
			}
			return
		}
		writeJSON(c, http.StatusOK, user)
	}
}

// createUserHandler handles POST /admin/directory/v1/users
func createUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&user); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateUser(c.Request.Context(), user)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(c, http.StatusConflict, "duplicate", "User already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create user")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// updateUserHandler handles PUT /admin/directory/v1/users/:userKey
func updateUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		var user model.User
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&user); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateUser(c.Request.Context(), userKey, user)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update user")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// patchUserHandler handles PATCH /admin/directory/v1/users/:userKey
func patchUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchUser(c.Request.Context(), userKey, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch user")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// deleteUserHandler handles DELETE /admin/directory/v1/users/:userKey
func deleteUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		if err := st.DeleteUser(c.Request.Context(), userKey); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete user")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// makeAdminHandler handles POST /admin/directory/v1/users/:userKey/makeAdmin
func makeAdminHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		var req struct {
			Status bool `json:"status"`
		}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		if err := st.MakeAdmin(c.Request.Context(), userKey, req.Status); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to make admin")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// undeleteUserHandler handles POST /admin/directory/v1/users/:userKey/undelete
func undeleteUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		if _, err := st.UndeleteUser(c.Request.Context(), userKey); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found in deleted users")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to undelete user")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// signOutUserHandler handles POST /admin/directory/v1/users/:userKey/signOut
func signOutUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		if err := st.SignOutUser(c.Request.Context(), userKey); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to sign out user")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// addUserAliasHandler handles POST /admin/directory/v1/users/:userKey/aliases
func addUserAliasHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		var req struct {
			Alias string `json:"alias"`
		}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		if req.Alias == "" {
			writeError(c, http.StatusBadRequest, "invalid", "alias is required")
			return
		}
		if err := st.AddUserAlias(c.Request.Context(), userKey, req.Alias); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to add alias")
			}
			return
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#alias",
			"etag":  "\"placeholder\"",
			"alias": req.Alias,
		})
	}
}

// listUserAliasesHandler handles GET /admin/directory/v1/users/:userKey/aliases
func listUserAliasesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		aliases, err := st.ListUserAliases(c.Request.Context(), userKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list aliases")
			}
			return
		}
		if aliases == nil {
			aliases = []string{}
		}
		aliasItems := []gin.H{}
		for _, a := range aliases {
			aliasItems = append(aliasItems, gin.H{
				"kind":  "admin#directory#alias",
				"etag":  "\"placeholder\"",
				"alias": a,
			})
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":    "admin#directory#aliases",
			"etag":    "\"placeholder\"",
			"aliases": aliasItems,
		})
	}
}

// removeUserAliasHandler handles DELETE /admin/directory/v1/users/:userKey/aliases/:alias
func removeUserAliasHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		alias := c.Param("alias")
		if err := st.RemoveUserAlias(c.Request.Context(), userKey, alias); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Resource not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to remove alias")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// getUserPhotoHandler handles GET /admin/directory/v1/users/:userKey/photos/thumbnail
func getUserPhotoHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		photo, err := st.GetUserPhoto(c.Request.Context(), userKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Photo not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get photo")
			}
			return
		}
		writeJSON(c, http.StatusOK, photo)
	}
}

// updateUserPhotoHandler handles PUT /admin/directory/v1/users/:userKey/photos/thumbnail
func updateUserPhotoHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		var photo model.UserPhoto
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&photo); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		photo.Kind = "admin#directory#userPhoto"
		if err := st.UpdateUserPhoto(c.Request.Context(), userKey, photo); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update photo")
			}
			return
		}
		writeJSON(c, http.StatusOK, photo)
	}
}

// patchUserPhotoHandler handles PATCH /admin/directory/v1/users/:userKey/photos/thumbnail
func patchUserPhotoHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		var photo model.UserPhoto
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&photo); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		photo.Kind = "admin#directory#userPhoto"
		if err := st.UpdateUserPhoto(c.Request.Context(), userKey, photo); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update photo")
			}
			return
		}
		writeJSON(c, http.StatusOK, photo)
	}
}

// deleteUserPhotoHandler handles DELETE /admin/directory/v1/users/:userKey/photos/thumbnail
func deleteUserPhotoHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		if err := st.DeleteUserPhoto(c.Request.Context(), userKey); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "User or photo not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete photo")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// createGuestUserHandler handles POST /admin/directory/v1/users:createGuest
func createGuestUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&user); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateUser(c.Request.Context(), user)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(c, http.StatusConflict, "duplicate", "User already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create guest user")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// watchUsersHandler handles POST /admin/directory/v1/users:watch (stub)
func watchUsersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		writeJSON(c, http.StatusOK, gin.H{
			"kind": "admin#directory#channel",
			"id":   "",
			"etag": "\"placeholder\"",
		})
	}
}
