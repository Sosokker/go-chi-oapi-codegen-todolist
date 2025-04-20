package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Sosokker/todolist-backend/internal/api/models" // Generated models
	"github.com/Sosokker/todolist-backend/internal/auth"
	"github.com/Sosokker/todolist-backend/internal/config"
	"github.com/Sosokker/todolist-backend/internal/domain"
	"github.com/Sosokker/todolist-backend/internal/service"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"golang.org/x/oauth2"
)

// Compile-time check to ensure ApiHandler implements the interface
var _ ServerInterface = (*ApiHandler)(nil)

// ApiHandler holds dependencies for API handlers
type ApiHandler struct {
	services *service.ServiceRegistry
	cfg      *config.Config
	logger   *slog.Logger
	// Add other dependencies like cache if needed directly by handlers
}

// NewApiHandler creates a new handler instance
func NewApiHandler(services *service.ServiceRegistry, cfg *config.Config, logger *slog.Logger) *ApiHandler {
	return &ApiHandler{
		services: services,
		cfg:      cfg,
		logger:   logger,
	}
}

// --- Helper Functions ---

// SendJSONResponse sends a JSON response with a status code
func SendJSONResponse(w http.ResponseWriter, statusCode int, payload interface{}, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload != nil {
		err := json.NewEncoder(w).Encode(payload)
		if err != nil { // Log the error but response header is already sent
			logger.Error("Failed to encode JSON response", "error", err)
		}
	}
}

// SendJSONError sends a standardized JSON error response
func SendJSONError(w http.ResponseWriter, err error, defaultStatusCode int, logger *slog.Logger) {
	logger.Warn("API Error", "error", err)

	respErr := models.Error{
		Message: err.Error(),
	}
	statusCode := defaultStatusCode

	// Map domain errors to HTTP status codes
	switch {
	case errors.Is(err, domain.ErrNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, domain.ErrForbidden):
		statusCode = http.StatusForbidden
	case errors.Is(err, domain.ErrUnauthorized):
		statusCode = http.StatusUnauthorized
	case errors.Is(err, domain.ErrConflict):
		statusCode = http.StatusConflict
	case errors.Is(err, domain.ErrValidation), errors.Is(err, domain.ErrBadRequest):
		statusCode = http.StatusBadRequest
	case errors.Is(err, domain.ErrInternalServer):
		statusCode = http.StatusInternalServerError
		respErr.Message = "An internal error occurred."
	default:
		if statusCode < 500 {
			logger.Error("Unhandled error type in API mapping", "error", err, "defaultStatus", defaultStatusCode)
			statusCode = http.StatusInternalServerError
			respErr.Message = "An unexpected error occurred."
		}
	}

	respErr.Code = int32(statusCode)
	SendJSONResponse(w, statusCode, respErr, logger)
}

// parseAndValidateBody decodes JSON body and logs/sends error on failure.
func parseAndValidateBody(w http.ResponseWriter, r *http.Request, body interface{}, logger *slog.Logger) bool {
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		SendJSONError(w, fmt.Errorf("invalid request body: %w", domain.ErrBadRequest), http.StatusBadRequest, logger)
		return false
	}
	// TODO: Add struct validation here if needed (e.g., using go-Code Playground/validator)
	return true
}

// --- Mappers (Domain <-> Generated API Models) ---

func mapDomainUserToApi(user *domain.User) *models.User {
	if user == nil {
		return nil
	}
	userID := openapi_types.UUID(user.ID)
	email := openapi_types.Email(user.Email)
	emailVerified := user.EmailVerified
	createdAt := user.CreatedAt
	updatedAt := user.UpdatedAt

	return &models.User{
		Id:            &userID,
		Username:      user.Username,
		Email:         email,
		EmailVerified: &emailVerified,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt}
}

func mapDomainTagToApi(tag *domain.Tag) *models.Tag {
	if tag == nil {
		return nil
	}
	tagID := openapi_types.UUID(tag.ID)
	userID := openapi_types.UUID(tag.UserID)
	createdAt := tag.CreatedAt
	updatedAt := tag.UpdatedAt
	return &models.Tag{
		Id:        &tagID,
		UserId:    &userID,
		Name:      tag.Name,
		Color:     tag.Color,
		Icon:      tag.Icon,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt}
}

func mapDomainTodoToApi(todo *domain.Todo) *models.Todo {
	if todo == nil {
		return nil
	}
	apiSubtasks := make([]models.Subtask, len(todo.Subtasks))
	for i, st := range todo.Subtasks {
		mappedSubtask := mapDomainSubtaskToApi(&st)
		if mappedSubtask != nil {
			apiSubtasks[i] = *mappedSubtask
		}
	}

	tagIDs := make([]openapi_types.UUID, len(todo.TagIDs))
	for i, domainID := range todo.TagIDs {
		tagIDs[i] = openapi_types.UUID(domainID)
	}

	todoID := openapi_types.UUID(todo.ID)
	userID := openapi_types.UUID(todo.UserID)
	createdAt := todo.CreatedAt
	updatedAt := todo.UpdatedAt

	return &models.Todo{
		Id:          &todoID,
		UserId:      &userID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      models.TodoStatus(todo.Status),
		Deadline:    todo.Deadline,
		TagIds:      tagIDs,
		Attachments: todo.Attachments,
		Subtasks:    &apiSubtasks,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt}
}

func mapDomainSubtaskToApi(subtask *domain.Subtask) *models.Subtask {
	if subtask == nil {
		return nil
	}
	subtaskID := openapi_types.UUID(subtask.ID)
	todoID := openapi_types.UUID(subtask.TodoID)
	createdAt := subtask.CreatedAt
	updatedAt := subtask.UpdatedAt

	return &models.Subtask{
		Id:          &subtaskID,
		TodoId:      &todoID,
		Description: subtask.Description,
		Completed:   subtask.Completed,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt}
}

func mapDomainAttachmentInfoToApi(info *domain.AttachmentInfo) *models.FileUploadResponse {
	if info == nil {
		return nil
	}
	return &models.FileUploadResponse{
		FileId:      info.FileID,
		FileName:    info.FileName,
		FileUrl:     info.FileURL,
		ContentType: info.ContentType,
		Size:        info.Size,
	}
}

// --- Auth Handlers ---

func (h *ApiHandler) SignupUserApi(w http.ResponseWriter, r *http.Request) {
	var body models.SignupRequest
	if !parseAndValidateBody(w, r, &body, h.logger) {
		return
	}

	creds := service.SignupCredentials{
		Username: body.Username,
		Email:    string(body.Email),
		Password: *body.Password,
	}

	user, err := h.services.Auth.Signup(r.Context(), creds)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	SendJSONResponse(w, http.StatusCreated, mapDomainUserToApi(user), h.logger)
}

func (h *ApiHandler) LoginUserApi(w http.ResponseWriter, r *http.Request) {
	var body models.LoginRequest
	if !parseAndValidateBody(w, r, &body, h.logger) {
		return
	}

	creds := service.LoginCredentials{
		Email:    string(body.Email),
		Password: *body.Password,
	}

	token, _, err := h.services.Auth.Login(r.Context(), creds)
	if err != nil {
		SendJSONError(w, err, http.StatusUnauthorized, h.logger)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.JWT.CookieName,
		Value:    token,
		Path:     h.cfg.JWT.CookiePath,
		Domain:   h.cfg.JWT.CookieDomain,
		Expires:  time.Now().Add(time.Duration(h.cfg.JWT.ExpiryMinutes) * time.Minute),
		HttpOnly: h.cfg.JWT.CookieHttpOnly,
		Secure:   h.cfg.JWT.CookieSecure,
		SameSite: parseSameSite(h.cfg.JWT.CookieSameSite),
	})

	resp := models.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	}
	SendJSONResponse(w, http.StatusOK, resp, h.logger)
}

func (h *ApiHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.JWT.CookieName,
		Value:    "",
		Path:     h.cfg.JWT.CookiePath,
		Domain:   h.cfg.JWT.CookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: h.cfg.JWT.CookieHttpOnly,
		Secure:   h.cfg.JWT.CookieSecure,
		SameSite: parseSameSite(h.cfg.JWT.CookieSameSite),
	})

	w.WriteHeader(http.StatusNoContent)
}

// Helper to parse SameSite string to http.SameSite type
func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

// --- Google OAuth Handlers ---

func (h *ApiHandler) InitiateGoogleLogin(w http.ResponseWriter, r *http.Request) {
	oauthCfg := h.services.Auth.GetGoogleAuthConfig()
	state := uuid.NewString()

	signedState := auth.SignState(state, []byte(h.cfg.OAuth.Google.StateSecret))

	http.SetCookie(w, &http.Cookie{
		Name:     auth.StateCookieName,
		Value:    signedState,
		Path:     "/",
		Expires:  time.Now().Add(auth.StateExpiry + 1*time.Minute),
		HttpOnly: true,
		Secure:   h.cfg.JWT.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	redirectURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	h.logger.Debug("Redirecting to Google OAuth", "url", redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *ApiHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	receivedCode := r.URL.Query().Get("code")
	receivedState := r.URL.Query().Get("state")

	stateCookie, err := r.Cookie(auth.StateCookieName)
	if err != nil {
		h.logger.WarnContext(ctx, "OAuth state cookie missing or error", "error", err)
		http.Redirect(w, r, "/login?error=state_missing", http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.StateCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.JWT.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	originalState, err := auth.VerifyAndExtractState(stateCookie.Value, []byte(h.cfg.OAuth.Google.StateSecret))
	if err != nil {
		h.logger.WarnContext(ctx, "OAuth state verification failed", "error", err, "receivedState", receivedState)
		errorParam := "state_invalid"
		if errors.Is(err, auth.ErrStateExpired) {
			errorParam = "state_expired"
		}
		http.Redirect(w, r, "/login?error="+errorParam, http.StatusTemporaryRedirect)
		return
	}

	if receivedState == "" || receivedState != originalState {
		h.logger.WarnContext(ctx, "OAuth state mismatch", "received", receivedState, "expected", originalState)
		http.Redirect(w, r, "/login?error=state_mismatch", http.StatusTemporaryRedirect)
		return
	}

	if receivedCode == "" {
		errorDesc := r.URL.Query().Get("error_description")
		h.logger.WarnContext(ctx, "Missing OAuth code parameter in callback", "error_desc", errorDesc)
		errorParam := url.QueryEscape(r.URL.Query().Get("error"))
		if errorParam == "" {
			errorParam = "missing_code"
		}
		http.Redirect(w, r, "/login?error="+errorParam, http.StatusTemporaryRedirect)
		return
	}

	token, user, err := h.services.Auth.HandleGoogleCallback(ctx, receivedCode)
	if err != nil {
		h.logger.ErrorContext(ctx, "Google callback handling failed in service", "error", err)
		errorParam := "auth_failed"
		if errors.Is(err, domain.ErrConflict) {
			errorParam = "auth_conflict"
		}
		http.Redirect(w, r, "/login?error="+errorParam, http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.JWT.CookieName,
		Value:    token,
		Path:     h.cfg.JWT.CookiePath,
		Domain:   h.cfg.JWT.CookieDomain,
		Expires:  time.Now().Add(time.Duration(h.cfg.JWT.ExpiryMinutes) * time.Minute),
		HttpOnly: h.cfg.JWT.CookieHttpOnly,
		Secure:   h.cfg.JWT.CookieSecure,
		SameSite: parseSameSite(h.cfg.JWT.CookieSameSite),
	})

	redirectURL := fmt.Sprintf("%s/oauth/callback#access_token=%s", h.cfg.Frontend.Url, url.QueryEscape(token))
	h.logger.InfoContext(ctx, "Google OAuth login successful", "userId", user.ID, "email", user.Email, "redirectingTo", redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// --- User Handlers ---

func (h *ApiHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With(slog.String("handler", "GetCurrentUser"))
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, logger)
		return
	}
	logger = logger.With(slog.String("userId", userID.String()))

	user, err := h.services.User.GetUserByID(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch user from service", "error", err)
		SendJSONError(w, err, http.StatusInternalServerError, logger)
		return
	}

	apiUser := mapDomainUserToApi(user)
	if apiUser == nil {
		logger.ErrorContext(ctx, "Failed to map domain user to API model")
		SendJSONError(w, domain.ErrInternalServer, http.StatusInternalServerError, logger)
		return
	}

	logger.DebugContext(ctx, "Successfully retrieved current user")
	SendJSONResponse(w, http.StatusOK, apiUser, logger)
}

func (h *ApiHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With(slog.String("handler", "UpdateCurrentUser"))

	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, logger)
		return
	}
	logger = logger.With(slog.String("userId", userID.String()))

	var body models.UpdateUserRequest
	if !parseAndValidateBody(w, r, &body, logger) {
		return
	}

	updateInput := service.UpdateUserInput{
		Username: body.Username,
	}

	updatedUser, err := h.services.User.UpdateUser(ctx, userID, updateInput)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update user in service", "error", err)
		SendJSONError(w, err, http.StatusInternalServerError, logger)
		return
	}

	apiUser := mapDomainUserToApi(updatedUser)
	if apiUser == nil {
		logger.ErrorContext(ctx, "Failed to map updated domain user to API model")
		SendJSONError(w, domain.ErrInternalServer, http.StatusInternalServerError, logger)
		return
	}

	logger.InfoContext(ctx, "Successfully updated current user")
	SendJSONResponse(w, http.StatusOK, apiUser, logger)
}

// --- Tag Handlers ---

func (h *ApiHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	var body models.CreateTagRequest
	if !parseAndValidateBody(w, r, &body, h.logger) {
		// TODO: Add specific field validation checks here or rely on service layer
		return
	}

	input := service.CreateTagInput{
		Name:  body.Name,
		Color: body.Color,
		Icon:  body.Icon,
	}

	tag, err := h.services.Tag.CreateTag(r.Context(), userID, input)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	SendJSONResponse(w, http.StatusCreated, mapDomainTagToApi(tag), h.logger)
}

func (h *ApiHandler) ListUserTags(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	tags, err := h.services.Tag.ListUserTags(r.Context(), userID)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	apiTags := make([]models.Tag, len(tags))
	for i, tag := range tags {
		apiTags[i] = *mapDomainTagToApi(&tag)
	}

	SendJSONResponse(w, http.StatusOK, apiTags, h.logger)
}

func (h *ApiHandler) GetTagById(w http.ResponseWriter, r *http.Request, tagId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}
	domainTagID := uuid.UUID(tagId)

	tag, err := h.services.Tag.GetTagByID(r.Context(), domainTagID, userID)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}
	SendJSONResponse(w, http.StatusOK, mapDomainTagToApi(tag), h.logger)
}

func (h *ApiHandler) UpdateTagById(w http.ResponseWriter, r *http.Request, tagId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}
	domainTagID := uuid.UUID(tagId)

	var body models.UpdateTagRequest
	if !parseAndValidateBody(w, r, &body, h.logger) {
		return
	}

	input := service.UpdateTagInput{
		Name:  body.Name,
		Color: body.Color,
		Icon:  body.Icon,
	}

	tag, err := h.services.Tag.UpdateTag(r.Context(), domainTagID, userID, input)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	SendJSONResponse(w, http.StatusOK, mapDomainTagToApi(tag), h.logger)
}

func (h *ApiHandler) DeleteTagById(w http.ResponseWriter, r *http.Request, tagId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}
	domainTagID := uuid.UUID(tagId)

	err = h.services.Tag.DeleteTag(r.Context(), domainTagID, userID)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Todo Handlers ---

func (h *ApiHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	var body models.CreateTodoRequest
	if !parseAndValidateBody(w, r, &body, h.logger) {
		return
	}

	var domainTagIDs []uuid.UUID
	if body.TagIds != nil {
		domainTagIDs = make([]uuid.UUID, len(*body.TagIds))
		for i, apiID := range *body.TagIds {
			domainTagIDs[i] = uuid.UUID(apiID)
		}
	} else {
		domainTagIDs = []uuid.UUID{}
	}

	input := service.CreateTodoInput{
		Title:       body.Title,
		Description: body.Description,
		Deadline:    body.Deadline,
		TagIDs:      domainTagIDs,
	}
	if body.Status != nil {
		domainStatus := domain.TodoStatus(*body.Status)
		input.Status = &domainStatus
	}

	todo, err := h.services.Todo.CreateTodo(r.Context(), userID, input)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}
	apiTodo := mapDomainTodoToApi(todo)
	SendJSONResponse(w, http.StatusCreated, apiTodo, h.logger)
}

func (h *ApiHandler) ListTodos(w http.ResponseWriter, r *http.Request, params ListTodosParams) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	input := service.ListTodosInput{
		Limit:  20,
		Offset: 0,
	}
	if params.Limit != nil {
		input.Limit = *params.Limit
	}
	if params.Offset != nil {
		input.Offset = *params.Offset
	}
	if params.Status != nil {
		domainStatus := domain.TodoStatus(*params.Status)
		input.Status = &domainStatus
	}
	if params.TagId != nil {
		input.TagID = params.TagId
	}
	if params.DeadlineBefore != nil {
		input.DeadlineBefore = params.DeadlineBefore
	}
	if params.DeadlineAfter != nil {
		input.DeadlineAfter = params.DeadlineAfter
	}

	todos, err := h.services.Todo.ListUserTodos(r.Context(), userID, input)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	apiTodos := make([]models.Todo, len(todos))
	for i, todo := range todos {
		apiTodos[i] = *mapDomainTodoToApi(&todo)
	}

	SendJSONResponse(w, http.StatusOK, apiTodos, h.logger)
}

func (h *ApiHandler) GetTodoById(w http.ResponseWriter, r *http.Request, todoId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	todo, err := h.services.Todo.GetTodoByID(r.Context(), todoId, userID)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	SendJSONResponse(w, http.StatusOK, mapDomainTodoToApi(todo), h.logger)
}

func (h *ApiHandler) UpdateTodoById(w http.ResponseWriter, r *http.Request, todoId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}
	domainTodoID := uuid.UUID(todoId)

	var body models.UpdateTodoRequest
	if !parseAndValidateBody(w, r, &body, h.logger) {
		return
	}

	input := service.UpdateTodoInput{
		Title:       body.Title,
		Description: body.Description,
		Deadline:    body.Deadline,
	}

	if body.Status != nil {
		domainStatus := domain.TodoStatus(*body.Status)
		input.Status = &domainStatus
	}

	if body.TagIds != nil {
		domainTagIDs := make([]uuid.UUID, len(*body.TagIds))
		for i, apiID := range *body.TagIds {
			domainTagIDs[i] = uuid.UUID(apiID)
		}
		input.TagIDs = &domainTagIDs
	}
	if body.Attachments != nil {
		input.Attachments = body.Attachments
	}

	todo, err := h.services.Todo.UpdateTodo(r.Context(), domainTodoID, userID, input)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	SendJSONResponse(w, http.StatusOK, mapDomainTodoToApi(todo), h.logger)
}

func (h *ApiHandler) DeleteTodoById(w http.ResponseWriter, r *http.Request, todoId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	err = h.services.Todo.DeleteTodo(r.Context(), todoId, userID)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Subtask Handlers ---

func (h *ApiHandler) CreateSubtaskForTodo(w http.ResponseWriter, r *http.Request, todoId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	var body models.CreateSubtaskRequest
	if !parseAndValidateBody(w, r, &body, h.logger) {
		return
	}
	input := service.CreateSubtaskInput{
		Description: body.Description,
	}

	subtask, err := h.services.Todo.CreateSubtask(r.Context(), todoId, userID, input)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	SendJSONResponse(w, http.StatusCreated, mapDomainSubtaskToApi(subtask), h.logger)
}

func (h *ApiHandler) ListSubtasksForTodo(w http.ResponseWriter, r *http.Request, todoId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	subtasks, err := h.services.Todo.ListSubtasks(r.Context(), todoId, userID)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	apiSubtasks := make([]models.Subtask, len(subtasks))
	for i, st := range subtasks {
		apiSubtasks[i] = *mapDomainSubtaskToApi(&st)
	}

	SendJSONResponse(w, http.StatusOK, apiSubtasks, h.logger)
}

func (h *ApiHandler) UpdateSubtaskById(w http.ResponseWriter, r *http.Request, todoId openapi_types.UUID, subtaskId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	var body models.UpdateSubtaskRequest
	if !parseAndValidateBody(w, r, &body, h.logger) {
		return
	}

	input := service.UpdateSubtaskInput{
		Description: body.Description,
		Completed:   body.Completed,
	}

	subtask, err := h.services.Todo.UpdateSubtask(r.Context(), todoId, subtaskId, userID, input)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	SendJSONResponse(w, http.StatusOK, mapDomainSubtaskToApi(subtask), h.logger)
}

func (h *ApiHandler) DeleteSubtaskById(w http.ResponseWriter, r *http.Request, todoId openapi_types.UUID, subtaskId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	err = h.services.Todo.DeleteSubtask(r.Context(), todoId, subtaskId, userID)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Attachment Handlers ---

func (h *ApiHandler) UploadTodoAttachment(w http.ResponseWriter, r *http.Request, todoId openapi_types.UUID) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		SendJSONError(w, fmt.Errorf("failed to parse multipart form: %w", domain.ErrBadRequest), http.StatusBadRequest, h.logger)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		SendJSONError(w, fmt.Errorf("error retrieving the file from form-data: %w", domain.ErrBadRequest), http.StatusBadRequest, h.logger)
		return
	}
	defer file.Close()

	fileName := handler.Filename
	fileSize := handler.Size

	attachmentInfo, err := h.services.Todo.AddAttachment(r.Context(), todoId, userID, fileName, fileSize, file)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	SendJSONResponse(w, http.StatusCreated, mapDomainAttachmentInfoToApi(attachmentInfo), h.logger)
}

func (h *ApiHandler) DeleteTodoAttachment(w http.ResponseWriter, r *http.Request, todoId openapi_types.UUID, attachmentId string) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	err = h.services.Todo.DeleteAttachment(r.Context(), todoId, userID, attachmentId)
	if err != nil {
		SendJSONError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
