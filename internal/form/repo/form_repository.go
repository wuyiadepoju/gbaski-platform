package repo

import (
	"context"
	"database/sql"

	"github.com/gbaski/gbaski-platform/internal/form/contracts"
	"github.com/gbaski/gbaski-platform/internal/form/domain"
	"github.com/gbaski/gbaski-shared/fee"
	sharedrepo "github.com/gbaski/gbaski-shared/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// FormRepositoryImpl is the concrete implementation of FormRepository
type FormRepositoryImpl struct {
	*sharedrepo.BaseRepository
}

// NewFormRepositoryImpl creates a new form repository using the provided DB connection.
// If db is nil, it falls back to the default BaseRepository connection.
func NewFormRepositoryImpl(db *sqlx.DB) contracts.FormRepository {
	base := sharedrepo.NewBaseRepository()
	if db != nil {
		base.DB = db
	}
	return &FormRepositoryImpl{BaseRepository: base}
}

// Ensure FormRepositoryImpl implements contracts.FormRepository
var _ contracts.FormRepository = (*FormRepositoryImpl)(nil)

func (r *FormRepositoryImpl) CreateForm(ctx context.Context, form *domain.FormAggregate) (*uuid.UUID, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	tx, formId, err := r.createFormTx(tx, form)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &formId, nil
}

func (r *FormRepositoryImpl) createFormTx(tx *sql.Tx, form *domain.FormAggregate) (*sql.Tx, uuid.UUID, error) {
	var formId uuid.UUID

	err := tx.QueryRow(`
	INSERT INTO forms (event_id, name, reg_type, schema, configs, price, currency, buyer_pays_fee, access_type, start_date, end_date, user_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING id`,
		form.EventID,
		form.Name,
		form.RegType,
		form.Schema,
		form.Configs,
		form.Price,
		form.Currency,
		form.BuyerPaysFee,
		form.AccessType,
		form.StartDate,
		form.EndDate,
		form.UserID,
	).Scan(&formId)

	if err != nil {
		return nil, uuid.Nil, err
	}

	if form.IsPrimaryForm {
		if _, err = tx.Exec(`UPDATE events SET form_id = $1 WHERE id = $2`, formId, form.EventID); err != nil {
			return nil, uuid.Nil, err
		}
	}

	return tx, formId, nil
}

func (r *FormRepositoryImpl) GetFormsByEventId(ctx context.Context, eventId uuid.UUID) ([]domain.FormQueryModel, error) {
	var forms []domain.FormQueryModel

	rows, err := r.DB.QueryxContext(ctx, `
		SELECT 
			f.id,
			f.event_id,
			f.name,
			f.reg_type,
			f.schema,
			f.configs,
			f.price,
			f.currency,
			e.name AS event_name,
			f.created_at,
			f.updated_at
		FROM forms f
		JOIN events e ON e.id = f.event_id
		WHERE f.event_id = $1`,
		eventId,
	)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var form domain.FormQueryModel
		if err = rows.StructScan(&form); err != nil {
			return nil, err
		}

		form.FeeRate = fee.New(string(form.RegType)).PlatformFeeRate()
		forms = append(forms, form)
	}

	return forms, nil
}

func (r *FormRepositoryImpl) UpdateForm(ctx context.Context, form *domain.FormAggregate) error {
	if form.ID == nil {
		return nil
	}

	_, err := r.DB.ExecContext(ctx, `
	UPDATE forms SET
		event_id = $1,
		name = $2,
		reg_type = $3,
		schema = $4,
		configs = $5,
		price = $6,
		access_type = $7,
		start_date = $8,
		end_date = $9,
		updated_at = NOW()
	WHERE id = $10`,
		form.EventID,
		form.Name,
		form.RegType,
		form.Schema,
		form.Configs,
		form.Price,
		form.AccessType,
		form.StartDate,
		form.EndDate,
		form.ID,
	)

	return err
}

