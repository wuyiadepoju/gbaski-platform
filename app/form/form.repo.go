package form

import (
	"context"
	"database/sql"

	"github.com/gbaski/gbaski-shared/fee"
	"github.com/gbaski/gbaski-shared/repo"

	"github.com/google/uuid"
)

type Repository struct {
	*repo.BaseRepository
}

func NewRepository() *Repository {
	return &Repository{repo.NewBaseRepository()}
}

func (r *Repository) CreateForm(form FormModel) (*uuid.UUID, error) {

	tx, err := r.DB.BeginTx(context.Background(), nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	tx, formId, err := r.CreateFormTx(tx, form)

	if err != nil {
		return nil, err
	}

	tx.Commit()

	return &formId, nil
}

func (r *Repository) CreateFormTx(tx *sql.Tx, form FormModel) (*sql.Tx, uuid.UUID, error) {

	var formId uuid.UUID

	err := tx.QueryRow(`
	INSERT INTO forms (event_id, name, reg_type, schema, configs, price, currency, buyer_pays_fee, access_type, start_date, end_date, user_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING id`,
		form.EventId,
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
		form.UserId,
	).Scan(&formId)

	if err != nil {
		return nil, uuid.Nil, err
	}

	if form.IsPrimaryForm {
		_, err = tx.Exec(`UPDATE events SET form_id = $1 WHERE id = $2`, formId, form.EventId)

		if err != nil {
			return nil, uuid.Nil, err
		}
	}

	return tx, formId, nil
}

func (r *Repository) GetFormsByEventId(eventId uuid.UUID) ([]FormItem, error) {

	var forms []FormItem

	rows, err := r.DB.Queryx(`
		SELECT * FROM forms WHERE event_id = $1`,
		eventId,
	)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var form FormItem
		err = rows.StructScan(&form)
		if err != nil {
			return nil, err
		}

		form.FeeRate = fee.New(string(form.RegType)).PlatformFeeRate()
		forms = append(forms, form)
	}

	return forms, nil
}

func (r *Repository) UpdateForm(form FormModel, formId uuid.UUID) error {

	_, err := r.DB.Exec(`
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
		form.EventId,
		form.Name,
		form.RegType,
		form.Schema,
		form.Configs,
		form.Price,
		form.AccessType,
		form.StartDate,
		form.EndDate,
		formId,
	)

	if err != nil {
		return err
	}

	return nil
}
