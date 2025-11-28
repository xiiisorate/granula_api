// Package repository handles data access for Workspace Service.
package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkspaceRepository handles workspace database operations.
type WorkspaceRepository struct {
	db *gorm.DB
}

// NewWorkspaceRepository creates a new WorkspaceRepository.
func NewWorkspaceRepository(db *gorm.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

// Create creates a new workspace.
func (r *WorkspaceRepository) Create(workspace *Workspace) error {
	return r.db.Create(workspace).Error
}

// FindByID finds a workspace by ID.
func (r *WorkspaceRepository) FindByID(id uuid.UUID) (*Workspace, error) {
	var workspace Workspace
	err := r.db.Preload("Members").Where("id = ?", id).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

// FindByOwnerID finds workspaces by owner ID.
func (r *WorkspaceRepository) FindByOwnerID(ownerID uuid.UUID) ([]Workspace, error) {
	var workspaces []Workspace
	err := r.db.Preload("Members").Where("owner_id = ?", ownerID).Find(&workspaces).Error
	return workspaces, err
}

// FindByUserID finds workspaces where user is a member.
func (r *WorkspaceRepository) FindByUserID(userID uuid.UUID, page, pageSize int) ([]Workspace, int64, error) {
	var workspaces []Workspace
	var total int64

	// Count total
	subQuery := r.db.Model(&WorkspaceMember{}).Select("workspace_id").Where("user_id = ?", userID)
	r.db.Model(&Workspace{}).Where("id IN (?) OR owner_id = ?", subQuery, userID).Count(&total)

	// Fetch with pagination
	offset := (page - 1) * pageSize
	err := r.db.Preload("Members").
		Where("id IN (?) OR owner_id = ?", subQuery, userID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&workspaces).Error

	return workspaces, total, err
}

// Update updates a workspace.
func (r *WorkspaceRepository) Update(workspace *Workspace) error {
	return r.db.Save(workspace).Error
}

// UpdateFields updates specific workspace fields.
func (r *WorkspaceRepository) UpdateFields(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&Workspace{}).Where("id = ?", id).Updates(updates).Error
}

// Delete soft deletes a workspace.
func (r *WorkspaceRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Workspace{}, "id = ?", id).Error
}

// MemberRepository handles workspace member database operations.
type MemberRepository struct {
	db *gorm.DB
}

// NewMemberRepository creates a new MemberRepository.
func NewMemberRepository(db *gorm.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

// Create adds a member to workspace.
func (r *MemberRepository) Create(member *WorkspaceMember) error {
	return r.db.Create(member).Error
}

// FindByWorkspaceAndUser finds a member by workspace and user IDs.
func (r *MemberRepository) FindByWorkspaceAndUser(workspaceID, userID uuid.UUID) (*WorkspaceMember, error) {
	var member WorkspaceMember
	err := r.db.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// FindByWorkspaceID finds all members of a workspace.
func (r *MemberRepository) FindByWorkspaceID(workspaceID uuid.UUID) ([]WorkspaceMember, error) {
	var members []WorkspaceMember
	err := r.db.Where("workspace_id = ?", workspaceID).Find(&members).Error
	return members, err
}

// UpdateRole updates member role.
func (r *MemberRepository) UpdateRole(workspaceID, userID uuid.UUID, role string) error {
	return r.db.Model(&WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Update("role", role).Error
}

// Delete removes a member from workspace.
func (r *MemberRepository) Delete(workspaceID, userID uuid.UUID) error {
	return r.db.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Delete(&WorkspaceMember{}).Error
}

// IsMember checks if user is a member of workspace.
func (r *MemberRepository) IsMember(workspaceID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(&count).Error
	return count > 0, err
}

