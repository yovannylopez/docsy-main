package services

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

// AuditService provides centralized methods for recording audit operations
type AuditService struct {
	auditRepo ports.AuditRepository
}

// NewAuditService creates a new instance of AuditService
func NewAuditService(auditRepo ports.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// LogCreate records a create operation
func (s *AuditService) LogCreate(
	ctx context.Context,
	userID string,
	resource string,
	resourceID string,
	newData any,
	ipAddress *string,
	userAgent *string,
) error {
	newDataJSON, err := s.marshalData(newData)
	if err != nil {
		return fmt.Errorf("failed to marshal new data: %w", err)
	}

	log := &entities.AuditLog{
		ID:         uuid.NewString(),
		UserID:     &userID,
		Action:     domain.AuditActionCreate,
		Resource:   &resource,
		ResourceID: &resourceID,
		Result:     domain.AuditResultSuccess,
		Message:    s.ptrString(fmt.Sprintf("Created %s with ID %s", resource, resourceID)),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		NewData:    newDataJSON,
		CreatedAt:  time.Now(),
	}

	return s.auditRepo.LogAction(ctx, log)
}

// LogUpdate records an update operation
func (s *AuditService) LogUpdate(
	ctx context.Context,
	userID string,
	resource string,
	resourceID string,
	oldData any,
	newData any,
	ipAddress *string,
	userAgent *string,
) error {
	oldDataJSON, err := s.marshalData(oldData)
	if err != nil {
		return fmt.Errorf("failed to marshal old data: %w", err)
	}

	newDataJSON, err := s.marshalData(newData)
	if err != nil {
		return fmt.Errorf("failed to marshal new data: %w", err)
	}

	// Compare data to identify modified fields
	changedFields := s.compareData(oldData, newData)

	log := &entities.AuditLog{
		ID:            uuid.NewString(),
		UserID:        &userID,
		Action:        domain.AuditActionUpdate,
		Resource:      &resource,
		ResourceID:    &resourceID,
		Result:        domain.AuditResultSuccess,
		Message:       s.ptrString(fmt.Sprintf("Updated %s with ID %s", resource, resourceID)),
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		PreviousData:  oldDataJSON,
		NewData:       newDataJSON,
		ChangedFields: changedFields,
		CreatedAt:     time.Now(),
	}

	return s.auditRepo.LogAction(ctx, log)
}

// LogDelete records a delete operation
func (s *AuditService) LogDelete(
	ctx context.Context,
	userID string,
	resource string,
	resourceID string,
	oldData any,
	ipAddress *string,
	userAgent *string,
) error {
	oldDataJSON, err := s.marshalData(oldData)
	if err != nil {
		return fmt.Errorf("failed to marshal old data: %w", err)
	}

	log := &entities.AuditLog{
		ID:           uuid.NewString(),
		UserID:       &userID,
		Action:       domain.AuditActionDelete,
		Resource:     &resource,
		ResourceID:   &resourceID,
		Result:       domain.AuditResultSuccess,
		Message:      s.ptrString(fmt.Sprintf("Deleted %s with ID %s", resource, resourceID)),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		PreviousData: oldDataJSON,
		CreatedAt:    time.Now(),
	}

	return s.auditRepo.LogAction(ctx, log)
}

// LogRead records a read operation (optional, for complete audit)
func (s *AuditService) LogRead(
	ctx context.Context,
	userID string,
	resource string,
	resourceID *string,
	ipAddress *string,
	userAgent *string,
) error {
	message := fmt.Sprintf("Read %s", resource)
	if resourceID != nil {
		message = fmt.Sprintf("Read %s with ID %s", resource, *resourceID)
	}

	log := &entities.AuditLog{
		ID:         uuid.NewString(),
		UserID:     &userID,
		Action:     domain.AuditActionRead,
		Resource:   &resource,
		ResourceID: resourceID,
		Result:     domain.AuditResultSuccess,
		Message:    &message,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
	}

	return s.auditRepo.LogAction(ctx, log)
}

// LogAction records a custom action
func (s *AuditService) LogAction(
	ctx context.Context,
	userID string,
	action string,
	resource string,
	resourceID *string,
	result string,
	message string,
	ipAddress *string,
	userAgent *string,
	previousData any,
	newData any,
) error {
	var oldDataJSON, newDataJSON *map[string]any
	var err error

	if previousData != nil {
		oldDataJSON, err = s.marshalData(previousData)
		if err != nil {
			return fmt.Errorf("failed to marshal previous data: %w", err)
		}
	}

	if newData != nil {
		newDataJSON, err = s.marshalData(newData)
		if err != nil {
			return fmt.Errorf("failed to marshal new data: %w", err)
		}
	}

	var changedFields []string
	if oldDataJSON != nil && newDataJSON != nil {
		changedFields = s.compareData(previousData, newData)
	}

	log := &entities.AuditLog{
		ID:            uuid.NewString(),
		UserID:        &userID,
		Action:        action,
		Resource:      &resource,
		ResourceID:    resourceID,
		Result:        result,
		Message:       &message,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		PreviousData:  oldDataJSON,
		NewData:       newDataJSON,
		ChangedFields: changedFields,
		CreatedAt:     time.Now(),
	}

	return s.auditRepo.LogAction(ctx, log)
}

// marshalData converts data to a map for audit fields, excluding relational fields
//
//nolint:gocognit,gocyclo
func (s *AuditService) marshalData(data any) (*map[string]any, error) {
	if data == nil {
		return nil, nil
	}

	// If it is already a map, use it directly
	if dataMap, ok := data.(map[string]any); ok {
		return &dataMap, nil
	}

	// If it is a *map, dereference it
	if ptrMap, ok := data.(*map[string]any); ok {
		if ptrMap == nil {
			return nil, nil
		}
		return ptrMap, nil
	}

	// For structs, use reflection to exclude relational fields
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil
		}
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		// Create a map with only non-relational fields
		result := make(map[string]any)
		typ := val.Type()

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldVal := val.Field(i)

			// Ignore unexported fields
			if !fieldVal.CanInterface() {
				continue
			}

			// Ignore fields with db:"-" (relational fields)
			dbTag := field.Tag.Get("db")
			if dbTag == "-" {
				continue
			}

			// Ignore fields that are pointers to structs (relations)
			if fieldVal.Kind() == reflect.Ptr {
				if fieldVal.IsNil() {
					// Include nil as nil
					jsonTag := field.Tag.Get("json")
					if jsonTag != "" && jsonTag != "-" {
						jsonName := s.getJSONFieldName(jsonTag)
						if jsonName != "" {
							result[jsonName] = nil
						}
					}
					continue
				}
				// If it is a pointer to a struct, it could be a relation, but include it if it has a value
				elemType := fieldVal.Elem().Type()
				if elemType.Kind() == reflect.Struct {
					// Check if it is a common relation type (User, State, Dependency, etc.)
					typeName := elemType.Name()
					if s.isRelationType(typeName) {
						// For relations, only include the ID if it exists
						idField := fieldVal.Elem().FieldByName("ID")
						if idField.IsValid() && idField.CanInterface() {
							jsonTag := field.Tag.Get("json")
							if jsonTag != "" && jsonTag != "-" {
								jsonName := s.getJSONFieldName(jsonTag)
								if jsonName != "" {
									result[jsonName+"_id"] = idField.Interface()
								}
							}
						}
						continue
					}
				}
			}

			// Ignore slices of structs (relations such as Attachments, Comments)
			if fieldVal.Kind() == reflect.Slice {
				jsonTag := field.Tag.Get("json")
				if jsonTag != "" && jsonTag != "-" {
					// Include as empty array or IDs only
					jsonName := s.getJSONFieldName(jsonTag)
					if jsonName != "" {
						result[jsonName] = []any{}
					}
				}
				continue
			}

			// Include the field
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" && jsonTag != "-" {
				jsonName := s.getJSONFieldName(jsonTag)
				if jsonName != "" {
					// Handle time.Time correctly
					switch fieldVal.Type() {
					case reflect.TypeOf(time.Time{}):
						result[jsonName] = fieldVal.Interface().(time.Time).Format(time.RFC3339)
					case reflect.TypeOf((*time.Time)(nil)).Elem():
						if !fieldVal.IsNil() {
							result[jsonName] = fieldVal.Interface().(*time.Time).Format(time.RFC3339)
						} else {
							result[jsonName] = nil
						}
					default:
						result[jsonName] = fieldVal.Interface()
					}
				}
			}
		}

		return &result, nil
	}

	// For other types, use standard json.Marshal
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var m map[string]any
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return &m, nil
}

// getJSONFieldName extracts the JSON field name from the tag
func (s *AuditService) getJSONFieldName(jsonTag string) string {
	if jsonTag == "" || jsonTag == "-" {
		return ""
	}
	parts := strings.Split(jsonTag, ",")
	return parts[0]
}

// isRelationType determines if a type is a common relation type
func (s *AuditService) isRelationType(typeName string) bool {
	relationTypes := []string{
		"User", "CommunicationState", "Dependency", "CommunicationType",
		"Attachment", "CommunicationComment",
	}
	for _, rt := range relationTypes {
		if typeName == rt {
			return true
		}
	}
	return false
}

// compareData compares two structures and returns the names of modified fields
func (s *AuditService) compareData(oldData, newData any) []string {
	if oldData == nil || newData == nil {
		return []string{}
	}

	oldVal := reflect.ValueOf(oldData)
	newVal := reflect.ValueOf(newData)

	// If they are pointers, get the underlying value
	if oldVal.Kind() == reflect.Ptr {
		if oldVal.IsNil() {
			return []string{}
		}
		oldVal = oldVal.Elem()
	}

	if newVal.Kind() == reflect.Ptr {
		if newVal.IsNil() {
			return []string{}
		}
		newVal = newVal.Elem()
	}

	// Handle maps
	if oldVal.Kind() == reflect.Map && newVal.Kind() == reflect.Map {
		return s.compareMaps(oldVal, newVal)
	}

	// Handle structs
	if oldVal.Kind() == reflect.Struct && newVal.Kind() == reflect.Struct {
		return s.compareStructs(oldVal, newVal)
	}

	// If neither maps nor structs, return empty
	return []string{}
}

// compareMaps compares two maps and returns the keys that changed
func (s *AuditService) compareMaps(oldVal, newVal reflect.Value) []string {
	changedFields := []string{}

	// Get all unique keys from both maps
	allKeys := make(map[any]bool)
	for _, key := range oldVal.MapKeys() {
		allKeys[key.Interface()] = true
	}
	for _, key := range newVal.MapKeys() {
		allKeys[key.Interface()] = true
	}

	// Compare values for each key
	for key := range allKeys {
		keyVal := reflect.ValueOf(key)
		oldValue := oldVal.MapIndex(keyVal)
		newValue := newVal.MapIndex(keyVal)

		// If the key does not exist in one of the maps, consider it a change
		if !oldValue.IsValid() || !newValue.IsValid() {
			if keyStr, ok := key.(string); ok {
				if !s.shouldIgnoreField(keyStr) {
					changedFields = append(changedFields, keyStr)
				}
			}
			continue
		}

		// Compare values
		if !reflect.DeepEqual(oldValue.Interface(), newValue.Interface()) {
			if keyStr, ok := key.(string); ok {
				if !s.shouldIgnoreField(keyStr) {
					changedFields = append(changedFields, keyStr)
				}
			}
		}
	}

	return changedFields
}

// compareStructs compares two structs and returns the fields that changed
func (s *AuditService) compareStructs(oldVal, newVal reflect.Value) []string {
	changedFields := []string{}
	oldType := oldVal.Type()
	newType := newVal.Type()

	// Compare common fields
	for i := 0; i < oldType.NumField(); i++ {
		oldField := oldType.Field(i)
		oldFieldVal := oldVal.Field(i)

		// Ignore unexported fields
		if !oldFieldVal.CanInterface() {
			continue
		}

		// Find corresponding field in newData
		_, found := newType.FieldByName(oldField.Name)
		if !found {
			continue
		}

		newFieldVal := newVal.FieldByName(oldField.Name)
		if !newFieldVal.CanInterface() {
			continue
		}

		// Ignore specific fields that should not be compared
		if s.shouldIgnoreField(oldField.Name) {
			continue
		}

		// Compare values
		if !reflect.DeepEqual(oldFieldVal.Interface(), newFieldVal.Interface()) {
			changedFields = append(changedFields, oldField.Name)
		}
	}

	return changedFields
}

// shouldIgnoreField determines if a field should be ignored in the comparison
func (s *AuditService) shouldIgnoreField(fieldName string) bool {
	ignoreFields := []string{
		"UpdatedAt",
		"UpdatedBy",
		"PasswordHash",
		"Password",
		"Token",
		"Secret",
	}

	for _, ignore := range ignoreFields {
		if fieldName == ignore {
			return true
		}
	}

	return false
}

// ptrString converts a string to *string
func (s *AuditService) ptrString(str string) *string {
	return &str
}
