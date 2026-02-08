package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tokyoyuan/dvd-rental/internal/store/model"
	"github.com/tokyoyuan/dvd-rental/internal/store/repository/sqlcgen"
)

// StaffRepository defines the data access interface for staff members.
type StaffRepository interface {
	GetStaff(ctx context.Context, staffID int32) (model.Staff, error)
	GetStaffByUsername(ctx context.Context, username string) (model.Staff, error)
	ListStaff(ctx context.Context, limit, offset int32) ([]model.Staff, error)
	ListActiveStaff(ctx context.Context, limit, offset int32) ([]model.Staff, error)
	ListStaffByStore(ctx context.Context, storeID, limit, offset int32) ([]model.Staff, error)
	ListActiveStaffByStore(ctx context.Context, storeID, limit, offset int32) ([]model.Staff, error)
	CountStaff(ctx context.Context) (int64, error)
	CountActiveStaff(ctx context.Context) (int64, error)
	CountStaffByStore(ctx context.Context, storeID int32) (int64, error)
	CountActiveStaffByStore(ctx context.Context, storeID int32) (int64, error)
	CreateStaff(ctx context.Context, params CreateStaffParams) (model.Staff, error)
	UpdateStaff(ctx context.Context, params UpdateStaffParams) (model.Staff, error)
	DeactivateStaff(ctx context.Context, staffID int32) (model.Staff, error)
	UpdateStaffPassword(ctx context.Context, staffID int32, passwordHash string) error
}

// CreateStaffParams holds the parameters for creating a new staff member.
type CreateStaffParams struct {
	FirstName    string
	LastName     string
	AddressID    int32
	Email        string
	StoreID      int32
	Username     string
	PasswordHash string
}

// UpdateStaffParams holds the parameters for updating a staff member.
type UpdateStaffParams struct {
	StaffID   int32
	FirstName string
	LastName  string
	AddressID int32
	Email     string
	StoreID   int32
	Username  string
}

type staffRepository struct {
	q *sqlcgen.Queries
}

// NewStaffRepository creates a new StaffRepository backed by PostgreSQL.
func NewStaffRepository(pool *pgxpool.Pool) StaffRepository {
	return &staffRepository{q: sqlcgen.New(pool)}
}

func (r *staffRepository) GetStaff(ctx context.Context, staffID int32) (model.Staff, error) {
	row, err := r.q.GetStaff(ctx, staffID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Staff{}, ErrNotFound
		}
		return model.Staff{}, fmt.Errorf("get staff: %w", err)
	}
	return model.Staff{
		StaffID:    row.StaffID,
		FirstName:  row.FirstName,
		LastName:   row.LastName,
		AddressID:  row.AddressID,
		Email:      textToString(row.Email),
		StoreID:    row.StoreID,
		Active:     row.Active,
		Username:   row.Username,
		Picture:    row.Picture,
		LastUpdate: row.LastUpdate.Time,
	}, nil
}

func (r *staffRepository) GetStaffByUsername(ctx context.Context, username string) (model.Staff, error) {
	row, err := r.q.GetStaffByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Staff{}, ErrNotFound
		}
		return model.Staff{}, fmt.Errorf("get staff by username: %w", err)
	}
	return model.Staff{
		StaffID:      row.StaffID,
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		AddressID:    row.AddressID,
		Email:        textToString(row.Email),
		StoreID:      row.StoreID,
		Active:       row.Active,
		Username:     row.Username,
		PasswordHash: textToString(row.PasswordHash),
		LastUpdate:   row.LastUpdate.Time,
	}, nil
}

func (r *staffRepository) ListStaff(ctx context.Context, limit, offset int32) ([]model.Staff, error) {
	rows, err := r.q.ListStaff(ctx, sqlcgen.ListStaffParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list staff: %w", err)
	}
	return toStaffListFromListRows(rows), nil
}

func (r *staffRepository) ListActiveStaff(ctx context.Context, limit, offset int32) ([]model.Staff, error) {
	rows, err := r.q.ListActiveStaff(ctx, sqlcgen.ListActiveStaffParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list active staff: %w", err)
	}
	return toStaffListFromActiveRows(rows), nil
}

func (r *staffRepository) ListStaffByStore(ctx context.Context, storeID, limit, offset int32) ([]model.Staff, error) {
	rows, err := r.q.ListStaffByStore(ctx, sqlcgen.ListStaffByStoreParams{
		StoreID: storeID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list staff by store: %w", err)
	}
	return toStaffListFromByStoreRows(rows), nil
}

func (r *staffRepository) ListActiveStaffByStore(ctx context.Context, storeID, limit, offset int32) ([]model.Staff, error) {
	rows, err := r.q.ListActiveStaffByStore(ctx, sqlcgen.ListActiveStaffByStoreParams{
		StoreID: storeID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list active staff by store: %w", err)
	}
	return toStaffListFromActiveByStoreRows(rows), nil
}

func (r *staffRepository) CountStaff(ctx context.Context) (int64, error) {
	count, err := r.q.CountStaff(ctx)
	if err != nil {
		return 0, fmt.Errorf("count staff: %w", err)
	}
	return count, nil
}

func (r *staffRepository) CountActiveStaff(ctx context.Context) (int64, error) {
	count, err := r.q.CountActiveStaff(ctx)
	if err != nil {
		return 0, fmt.Errorf("count active staff: %w", err)
	}
	return count, nil
}

func (r *staffRepository) CountStaffByStore(ctx context.Context, storeID int32) (int64, error) {
	count, err := r.q.CountStaffByStore(ctx, storeID)
	if err != nil {
		return 0, fmt.Errorf("count staff by store: %w", err)
	}
	return count, nil
}

func (r *staffRepository) CountActiveStaffByStore(ctx context.Context, storeID int32) (int64, error) {
	count, err := r.q.CountActiveStaffByStore(ctx, storeID)
	if err != nil {
		return 0, fmt.Errorf("count active staff by store: %w", err)
	}
	return count, nil
}

func (r *staffRepository) CreateStaff(ctx context.Context, params CreateStaffParams) (model.Staff, error) {
	row, err := r.q.CreateStaff(ctx, sqlcgen.CreateStaffParams{
		FirstName:    params.FirstName,
		LastName:     params.LastName,
		AddressID:    params.AddressID,
		Email:        stringToText(params.Email),
		StoreID:      params.StoreID,
		Username:     params.Username,
		PasswordHash: stringToText(params.PasswordHash),
	})
	if err != nil {
		return model.Staff{}, fmt.Errorf("create staff: %w", err)
	}
	return staffFromCreateRow(row), nil
}

func (r *staffRepository) UpdateStaff(ctx context.Context, params UpdateStaffParams) (model.Staff, error) {
	row, err := r.q.UpdateStaff(ctx, sqlcgen.UpdateStaffParams{
		StaffID:   params.StaffID,
		FirstName: params.FirstName,
		LastName:  params.LastName,
		AddressID: params.AddressID,
		Email:     stringToText(params.Email),
		StoreID:   params.StoreID,
		Username:  params.Username,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Staff{}, ErrNotFound
		}
		return model.Staff{}, fmt.Errorf("update staff: %w", err)
	}
	return staffFromUpdateRow(row), nil
}

func (r *staffRepository) DeactivateStaff(ctx context.Context, staffID int32) (model.Staff, error) {
	row, err := r.q.DeactivateStaff(ctx, staffID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Staff{}, ErrNotFound
		}
		return model.Staff{}, fmt.Errorf("deactivate staff: %w", err)
	}
	return staffFromDeactivateRow(row), nil
}

func (r *staffRepository) UpdateStaffPassword(ctx context.Context, staffID int32, passwordHash string) error {
	if err := r.q.UpdateStaffPassword(ctx, sqlcgen.UpdateStaffPasswordParams{
		StaffID:      staffID,
		PasswordHash: stringToText(passwordHash),
	}); err != nil {
		return fmt.Errorf("update staff password: %w", err)
	}
	return nil
}

// --- conversion helpers ---

func textToString(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func stringToText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func toStaffListFromListRows(rows []sqlcgen.ListStaffRow) []model.Staff {
	staff := make([]model.Staff, len(rows))
	for i, r := range rows {
		staff[i] = model.Staff{
			StaffID:    r.StaffID,
			FirstName:  r.FirstName,
			LastName:   r.LastName,
			AddressID:  r.AddressID,
			Email:      textToString(r.Email),
			StoreID:    r.StoreID,
			Active:     r.Active,
			Username:   r.Username,
			LastUpdate: r.LastUpdate.Time,
		}
	}
	return staff
}

func toStaffListFromActiveRows(rows []sqlcgen.ListActiveStaffRow) []model.Staff {
	staff := make([]model.Staff, len(rows))
	for i, r := range rows {
		staff[i] = model.Staff{
			StaffID:    r.StaffID,
			FirstName:  r.FirstName,
			LastName:   r.LastName,
			AddressID:  r.AddressID,
			Email:      textToString(r.Email),
			StoreID:    r.StoreID,
			Active:     r.Active,
			Username:   r.Username,
			LastUpdate: r.LastUpdate.Time,
		}
	}
	return staff
}

func toStaffListFromByStoreRows(rows []sqlcgen.ListStaffByStoreRow) []model.Staff {
	staff := make([]model.Staff, len(rows))
	for i, r := range rows {
		staff[i] = model.Staff{
			StaffID:    r.StaffID,
			FirstName:  r.FirstName,
			LastName:   r.LastName,
			AddressID:  r.AddressID,
			Email:      textToString(r.Email),
			StoreID:    r.StoreID,
			Active:     r.Active,
			Username:   r.Username,
			LastUpdate: r.LastUpdate.Time,
		}
	}
	return staff
}

func toStaffListFromActiveByStoreRows(rows []sqlcgen.ListActiveStaffByStoreRow) []model.Staff {
	staff := make([]model.Staff, len(rows))
	for i, r := range rows {
		staff[i] = model.Staff{
			StaffID:    r.StaffID,
			FirstName:  r.FirstName,
			LastName:   r.LastName,
			AddressID:  r.AddressID,
			Email:      textToString(r.Email),
			StoreID:    r.StoreID,
			Active:     r.Active,
			Username:   r.Username,
			LastUpdate: r.LastUpdate.Time,
		}
	}
	return staff
}

func staffFromCreateRow(r sqlcgen.CreateStaffRow) model.Staff {
	return model.Staff{
		StaffID:    r.StaffID,
		FirstName:  r.FirstName,
		LastName:   r.LastName,
		AddressID:  r.AddressID,
		Email:      textToString(r.Email),
		StoreID:    r.StoreID,
		Active:     r.Active,
		Username:   r.Username,
		LastUpdate: r.LastUpdate.Time,
	}
}

func staffFromUpdateRow(r sqlcgen.UpdateStaffRow) model.Staff {
	return model.Staff{
		StaffID:    r.StaffID,
		FirstName:  r.FirstName,
		LastName:   r.LastName,
		AddressID:  r.AddressID,
		Email:      textToString(r.Email),
		StoreID:    r.StoreID,
		Active:     r.Active,
		Username:   r.Username,
		LastUpdate: r.LastUpdate.Time,
	}
}

func staffFromDeactivateRow(r sqlcgen.DeactivateStaffRow) model.Staff {
	return model.Staff{
		StaffID:    r.StaffID,
		FirstName:  r.FirstName,
		LastName:   r.LastName,
		AddressID:  r.AddressID,
		Email:      textToString(r.Email),
		StoreID:    r.StoreID,
		Active:     r.Active,
		Username:   r.Username,
		LastUpdate: r.LastUpdate.Time,
	}
}
