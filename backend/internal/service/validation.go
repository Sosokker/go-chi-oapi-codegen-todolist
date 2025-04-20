package service

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/Sosokker/todolist-backend/internal/domain"
)

const (
	MinUsernameLength    = 3
	MaxUsernameLength    = 50
	MinPasswordLength    = 6
	MinTagNameLength     = 1
	MaxTagNameLength     = 50
	MaxTagIconLength     = 30
	MinTodoTitleLength   = 1
	MinSubtaskDescLength = 1
)

// Regex for simple hex color validation (#RRGGBB)
var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// ValidateUsername checks username constraints.
func ValidateUsername(username string) error {
	if len(username) < MinUsernameLength || len(username) > MaxUsernameLength {
		return fmt.Errorf("username must be between %d and %d characters: %w", MinUsernameLength, MaxUsernameLength, domain.ErrValidation)
	}
	// Add other constraints like allowed characters if needed
	return nil
}

// ValidateEmail checks if the email format is valid.
func ValidateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format: %w", domain.ErrValidation)
	}
	return nil
}

// ValidatePassword checks password length.
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters: %w", MinPasswordLength, domain.ErrValidation)
	}
	return nil
}

// ValidateSignupInput validates the input for user registration.
func ValidateSignupInput(creds SignupCredentials) error {
	if err := ValidateUsername(creds.Username); err != nil {
		return err
	}
	if err := ValidateEmail(creds.Email); err != nil {
		return err
	}
	if err := ValidatePassword(creds.Password); err != nil {
		return err
	}
	return nil
}

// ValidateLoginInput validates the input for user login.
func ValidateLoginInput(creds LoginCredentials) error {
	if err := ValidateEmail(creds.Email); err != nil {
		return err
	}
	if creds.Password == "" { // Password presence check
		return fmt.Errorf("password is required: %w", domain.ErrValidation)
	}
	return nil
}

// IsValidHexColor checks if a string is a valid #RRGGBB hex color.
func IsValidHexColor(color string) bool {
	return hexColorRegex.MatchString(color)
}

// ValidateTagName checks tag name constraints.
func ValidateTagName(name string) error {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < MinTagNameLength || len(trimmed) > MaxTagNameLength {
		return fmt.Errorf("tag name must be between %d and %d characters: %w", MinTagNameLength, MaxTagNameLength, domain.ErrValidation)
	}
	return nil
}

// ValidateTagIcon checks tag icon constraints.
func ValidateTagIcon(icon *string) error {
	if icon != nil && len(*icon) > MaxTagIconLength {
		return fmt.Errorf("tag icon cannot be longer than %d characters: %w", MaxTagIconLength, domain.ErrValidation)
	}
	return nil
}

// ValidateCreateTagInput validates input for creating a tag.
func ValidateCreateTagInput(input CreateTagInput) error {
	if err := ValidateTagName(input.Name); err != nil {
		return err
	}
	if input.Color != nil && !IsValidHexColor(*input.Color) {
		return fmt.Errorf("invalid color format (must be #RRGGBB): %w", domain.ErrValidation)
	}
	if err := ValidateTagIcon(input.Icon); err != nil {
		return err
	}
	return nil
}

// ValidateUpdateTagInput validates input for updating a tag.
func ValidateUpdateTagInput(input UpdateTagInput) error {
	if input.Name != nil {
		if err := ValidateTagName(*input.Name); err != nil {
			return err
		}
	}
	if input.Color != nil && !IsValidHexColor(*input.Color) {
		return fmt.Errorf("invalid color format (must be #RRGGBB): %w", domain.ErrValidation)
	}
	if err := ValidateTagIcon(input.Icon); err != nil { // Check pointer directly
		return err
	}
	return nil
}

// ValidateTodoTitle checks title constraints.
func ValidateTodoTitle(title string) error {
	if len(strings.TrimSpace(title)) < MinTodoTitleLength {
		return fmt.Errorf("todo title cannot be empty: %w", domain.ErrValidation)
	}
	return nil
}

// ValidateCreateTodoInput validates input for creating a todo.
func ValidateCreateTodoInput(input CreateTodoInput) error {
	if err := ValidateTodoTitle(input.Title); err != nil {
		return err
	}
	// Optional: Validate Status enum value if needed
	// Optional: Validate Deadline is not in the past?
	return nil
}

// ValidateUpdateTodoInput validates input for updating a todo.
func ValidateUpdateTodoInput(input UpdateTodoInput) error {
	if input.Title != nil {
		if err := ValidateTodoTitle(*input.Title); err != nil {
			return err
		}
	}
	// Optional: Validate Status enum value if needed
	// Optional: Validate Deadline is not in the past?
	return nil
}

// ValidateSubtaskDescription checks description constraints.
func ValidateSubtaskDescription(desc string) error {
	if len(strings.TrimSpace(desc)) < MinSubtaskDescLength {
		return fmt.Errorf("subtask description cannot be empty: %w", domain.ErrValidation)
	}
	return nil
}

// ValidateCreateSubtaskInput validates input for creating a subtask.
func ValidateCreateSubtaskInput(input CreateSubtaskInput) error {
	return ValidateSubtaskDescription(input.Description)
}

// ValidateUpdateSubtaskInput validates input for updating a subtask.
func ValidateUpdateSubtaskInput(input UpdateSubtaskInput) error {
	if input.Description != nil {
		if err := ValidateSubtaskDescription(*input.Description); err != nil {
			return err
		}
	}
	return nil
}

// ValidateListParams checks basic pagination parameters.
func ValidateListParams(limit, offset int) error {
	if limit < 0 {
		return fmt.Errorf("limit cannot be negative: %w", domain.ErrValidation)
	}
	if offset < 0 {
		return fmt.Errorf("offset cannot be negative: %w", domain.ErrValidation)
	}
	// Add max limit check if desired
	// if limit > MaxListLimit { ... }
	return nil
}
