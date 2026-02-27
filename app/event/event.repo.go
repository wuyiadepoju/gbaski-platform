package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gbaski/gbaski-event/pkg/event"
	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-shared/fee"
	"github.com/gbaski/gbaski-shared/util"

	"github.com/gbaski/gbaski-shared/pager"

	"github.com/google/uuid"
)

type RepositoryInterface interface {
	CreateEvent(event *EventModel) error
	GetEvent(id string) (*EventModel, error)
	UpdateEvent(event *EventModel) error
	DeleteEvent(id string) error
}

type Repository struct {
	event.Repository
}

func NewRepository() *Repository {
	return &Repository{*event.NewRepository()}
}

func (r *Repository) CreateEvent(event EventModel) (*EventItem, error) {

	var eventId uuid.UUID

	tx, err := r.DB.BeginTx(context.Background(), nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	err = tx.QueryRow(`
        INSERT INTO events (name, description, status, slug, category_id, mode_type, location, payment, duration_type, image_url, video_url, social_media, start_date, end_date, user_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
        RETURNING id`,
		event.Name,
		event.Description,
		event.Status,
		event.Slug,
		event.CategoryId,
		event.ModeType,
		event.Location,
		event.Payment,
		event.DurationType,
		event.ImageURL,
		event.VideoURL,
		event.SocialMedia,
		event.StartDate,
		event.EndDate,
		event.UserId,
	).Scan(&eventId)

	if err != nil {
		return nil, err
	}

	// log.PrettyPrint("event", event)

	tx, _, err = r.CreateForm(tx, FormModel{
		FormRequest: FormRequest{
			EventId:       eventId,
			Name:          event.Name,
			Schema:        event.Schema,
			Price:         0.0,
			Currency:      "NGN",
			FeePayer:      FeePayerSeller,
			AccessType:    event.AccessType,
			RegType:       event.RegType,
			IsPrimaryForm: true,
			StartDate:     event.StartDate,
			EndDate:       event.EndDate,
		},
		Configs: event.Configs,
		UserId:  event.UserId,
	})

	if err != nil {
		return nil, err
	}

	tx.Commit()

	return r.GetEvent(eventId)
}

func (r *Repository) CreateEventForm(form FormModel) (*uuid.UUID, error) {

	tx, err := r.DB.BeginTx(context.Background(), nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	tx, formId, err := r.CreateForm(tx, form)

	if err != nil {
		return nil, err
	}

	tx.Commit()

	return formId, nil

}

func (r *Repository) GetEventStats(eventId uuid.UUID, userId uuid.UUID) ([]EventStatsItem, error) {
	var stats []EventStatsItem

	rows, err := r.DB.Query(`
	SELECT
		f.reg_type AS name,
		f.name AS description,
		f.currency,
		SUM(t.qty) AS quantity,
		SUM(t.amount) AS amount
	FROM forms f
	JOIN registrations r ON f.id = r.form_id
	JOIN transactions t ON r.transaction_id = t.id
	WHERE f.event_id = $1
	  AND f.user_id = $2
	  AND f.reg_type IN ('ticket', 'entry', 'vote')
	GROUP BY f.reg_type, f.name, f.currency
	ORDER BY f.reg_type`,
		eventId, userId)

	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var statsItem EventStatsItem
		err = rows.Scan(&statsItem.Name, &statsItem.Description, &statsItem.Currency, &statsItem.Quantity, &statsItem.Amount)
		if err != nil {
			return stats, err
		}
		stats = append(stats, statsItem)
	}

	if err = rows.Err(); err != nil {
		return stats, err
	}

	return stats, nil
}

func (r *Repository) GetEventFeatures(eventId uuid.UUID) ([]EventFeature, error) {

	var features []EventFeature

	// Tweak this later
	rows, err := r.DB.Queryx(`SELECT id, name, 'description' as description, access_type, id as form_id, reg_type as tag FROM forms where event_id = $1;`, eventId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {

		var feature EventFeature

		err = rows.StructScan(&feature)
		if err != nil {
			return nil, err
		}

		features = append(features, feature)
	}

	return features, nil
}

func (r *Repository) GetEvents(queryModel EventQueryModel, userId uuid.UUID) (EventList, error) {

	sqlString := `SELECT 
   e.id,
   e.name,
   e.description,
   e.slug,
   u.name as brand_name,
   e.category_id,
   e.mode_type,
   e.location,
   e.payment,
   e.duration_type,
   e.status,
   COALESCE(NULLIF(e.image_url, ''), 'https:://image.gbaski.app/gbaski/event-image.webp') AS image_url,
   e.video_url,
   e.start_date,
   e.end_date,
   e.form_id,
   e.created_at,
   e.updated_at
FROM 
    events e 
    JOIN users u ON e.user_id = u.id
WHERE e.user_id = u.id
ORDER BY 
    e.created_at DESC, 
    e.start_date DESC`

	conditions := make(map[string]interface{})

	conditions["search"] = pager.SearchCondition{
		Search:  queryModel.Search,
		Columns: []string{"e.name", "e.description"},
	}

	conditions["filter"] = pager.FilterCondition{
		Filter:  queryModel.Filter,
		Columns: []string{"e.status"},
	}

	conditions["custom"] = pager.CustomCondition{
		Items:  []string{"e.user_id = :user_id"},
		Params: map[string]any{"user_id": userId},
	}

	pager := pager.NewPager[EventItem]().SetQuery(sqlString).SetConditions(conditions).SetPage(queryModel.Page)

	result, err := pager.Query()

	if err != nil {
		return EventList{}, err
	}

	if result.GetSize() == 0 && result.GetTotal() > 0 {
		result, err = pager.SetPage(queryModel.Page - 1).Query()

		if err != nil {
			return EventList{}, err
		}
	}

	if err != nil {
		fmt.Println("Error getting event stats: ", err.Error())
		return EventList{}, err
	}

	items := result.GetItems()

	for index, item := range items {
		item.Url = util.EventUrl(item.BrandName, item.Name)
		items[index] = item
	}

	return EventList{
		Items: items,
		Page:  result.GetPage(),
		Size:  result.GetSize(),
		Total: result.GetTotal(),
	}, nil
}

func (r *Repository) UpdateEventImage(eventId uuid.UUID, imageURL string) error {
	_, err := r.DB.Exec(`UPDATE events SET image_url = $1 WHERE id = $2`, imageURL, eventId)
	return err
}

func (r *Repository) UpdateEventVideo(eventId uuid.UUID, videoURL string) error {
	_, err := r.DB.Exec(`UPDATE events SET video_url = $1 WHERE id = $2`, videoURL, eventId)
	return err
}

func (r *Repository) UpdateEventStatus(eventId uuid.UUID, status EventStatus) (*EventItem, error) {
	_, err := r.DB.Exec(`UPDATE events SET status = $1 WHERE id = $2`, status, eventId)

	if err != nil {
		return nil, err
	}

	event, err := r.GetEvent(eventId)

	if err != nil {
		return nil, err
	}

	return event, nil
}

func (r *Repository) UpdateEvent(event EventModel) error {

	_, err := r.DB.NamedExec(`
		UPDATE events 
		SET name = :name, description = :description, slug = :slug, category_id = :category_id, mode_type = :mode_type, location = :location, payment = :payment, duration_type = :duration_type, form_id = :form_id, social_media = :social_media, start_date = :start_date, end_date = :end_date, image_url = :image_url, video_url = :video_url
		WHERE id = :id and user_id = :user_id`,
		event,
	)
	return err
}

func (r *Repository) DeleteEvent(eventId uuid.UUID) error {
	_, err := r.DB.Exec(`DELETE FROM events WHERE id = $1`, eventId)
	return err
}

func (r *Repository) CreateEventCategory(category EventCategory) (int, error) {

	var categoryId int

	err := r.DB.QueryRow(`
		INSERT INTO categories (name, event_type, public) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`,
		category.Name,
		category.EventType,
		category.Public,
	).Scan(&categoryId)

	if err != nil {
		return 0, err
	}

	return categoryId, nil
}

func (r *Repository) GetEventForms(userId uuid.UUID) ([]EventForm, error) {

	forms := []EventForm{}

	rows, err := r.DB.Queryx(`
        SELECT 
            f.id, 
			f.user_id,
            f.event_id, 
            f.name, 
            f.reg_type, 
            f.price, 
            f.currency, 
            f.fee_payer, 
            f.access_type, 
            f.configs, 
            f.start_date, 
            f.end_date, 
            e.name  AS event_name, 
            u.name  AS brand_name, 
            e.start_date AS event_date,
			CASE
				WHEN EXISTS (
					SELECT 1 
					FROM payout.providers pp 
					WHERE pp.user_id = u.id
				) AND EXISTS (
					SELECT 1 
					FROM payout.wallet_balance wb 
					WHERE wb.user_id = u.id AND wb.amount > 0
				) THEN true
				ELSE false
			END as direct_payment_enabled
        FROM 
            forms f 
            JOIN events e ON f.event_id = e.id 
            JOIN users u ON f.user_id = u.id
        WHERE 
            u.id = $1 
        ORDER BY 
            f.updated_at DESC, 
            f.created_at DESC, 
            e.status DESC
    `,
		userId,
	)

	if err != nil {
		log.Error("EventForm", "Event Forms Query", err)
		return nil, err
	}

	for rows.Next() {
		var form EventForm
		err = rows.StructScan(&form)
		if err != nil {
			log.Error("EventForm", "EventForm[] Data Mapping", err)
			return nil, err
		}

		schema, err := r.GetFormSchema(form.Id, form.RegType)

		if err != nil {
			log.Error("EventForm", "GetFormSchema", err)
			return nil, err
		}

		form.Schema = &schema

		form.FeeRate = fee.New(string(form.RegType)).PlatformFeeRate()

		forms = append(forms, form)
	}

	return forms, nil
}

func (r *Repository) CreateForm(tx *sql.Tx, form FormModel) (*sql.Tx, *uuid.UUID, error) {

	var formId uuid.UUID

	var err error

	if form.Id != nil {

		tx, formId, err = r.UpdateFormTx(tx, form)

		if err != nil {
			return nil, nil, err
		}

	} else {
		tx, formId, err = r.CreateFormTx(tx, form)

		if err != nil {
			return nil, nil, err
		}
	}

	if form.IsPrimaryForm {
		_, err = tx.Exec(`UPDATE events SET form_id = $1 WHERE id = $2`, formId, form.EventId)

		if err != nil {
			return nil, nil, err
		}
	}

	if form.RegType == RegistrationTypeTicket && form.Schema != nil {
		tx, err = CreateTicketTx(tx, formId, form.Schema)
		if err != nil {
			return nil, nil, err
		}
	}

	return tx, &formId, nil
}

func (r *Repository) UpdateFormTx(tx *sql.Tx, form FormModel) (*sql.Tx, uuid.UUID, error) {

	formId := *form.Id

	_, err := tx.Exec(`
	UPDATE forms SET
		event_id = $1,
		name = $2,
		reg_type = $3,
		schema = $4,
		configs = $5,
		price = $6,
		currency = $7,
		fee_payer = $8,
		access_type = $9,
		start_date = $10,
		end_date = $11,
		updated_at = NOW()
	WHERE id = $12`,
		form.EventId,
		form.Name,
		form.RegType,
		form.Schema,
		form.Configs,
		form.Price,
		form.Currency,
		form.FeePayer,
		form.AccessType,
		form.StartDate,
		form.EndDate,
		formId,
	)

	if err != nil {
		return tx, formId, err
	}

	return tx, formId, nil
}

func (r *Repository) CreateFormTx(tx *sql.Tx, form FormModel) (*sql.Tx, uuid.UUID, error) {

	var formId uuid.UUID

	err := tx.QueryRow(`
	INSERT INTO forms (event_id, name, reg_type, configs, access_type, price, currency, fee_payer, start_date, end_date, user_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id`,
		form.EventId,
		form.Name,
		form.RegType,
		form.Configs,
		form.AccessType,
		form.Price,
		form.Currency,
		form.FeePayer,
		form.StartDate,
		form.EndDate,
		form.UserId,
	).Scan(&formId)

	if err != nil {
		return nil, uuid.Nil, err
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
		forms = append(forms, form)
	}

	return forms, nil
}

func CreateTicketTx(tx *sql.Tx, formId uuid.UUID, schema json.RawMessage) (*sql.Tx, error) {

	if len(schema) == 0 {
		return tx, errors.New("form schema is empty")
	}

	var tickets []SchemaTicketItem
	if err := json.Unmarshal(schema, &tickets); err != nil {
		return nil, err
	}

	insertTickets := func(ticket SchemaTicketItem) error {

		_, err := tx.Exec(`
		INSERT INTO tickets 
		(name, description, payment, price, discount, capacity, "limit", index, form_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
			ticket.Name,
			ticket.Description,
			ticket.Payment,
			ticket.Price,
			ticket.Discount,
			ticket.Capacity,
			ticket.Limit,
			ticket.Index,
			formId,
		)

		if err != nil {
			return err
		}

		return nil
	}

	updateTickets := func(ticket SchemaTicketItem) error {
		_, err := tx.Exec(`
		UPDATE tickets SET
			name = $1,
			description = $2,
			payment = $3,
			price = $4,
			discount = $5,
			capacity = $6,
			"limit" = $7,
			index = $8,
			deleted = $9
		WHERE id = $10 AND form_id = $11
		`,
			ticket.Name,
			ticket.Description,
			ticket.Payment,
			ticket.Price,
			ticket.Discount,
			ticket.Capacity,
			ticket.Limit,
			ticket.Index,
			ticket.Deleted,
			ticket.Id,
			formId,
		)

		if err != nil {
			return err
		}

		return nil
	}

	for _, ticket := range tickets {

		if ticket.Id != 0 {

			updateTickets(ticket)
		} else {
			insertTickets(ticket)
		}

	}

	return tx, nil
}

func (r *Repository) GetEventsReport(userId uuid.UUID) (*EventsReport, error) {

	var reportJson json.RawMessage

	err := r.DB.QueryRow(`
		SELECT * FROM get_events_report($1)`,
		userId,
	).Scan(&reportJson)

	if err != nil {
		return nil, err
	}

	var report EventsReport

	err = json.Unmarshal(reportJson, &report)

	if err != nil {
		return nil, err
	}

	return &report, nil
}
