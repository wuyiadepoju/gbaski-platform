package registration

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/gbaski/gbaski-shared/pager"
	"github.com/gbaski/gbaski-shared/repo"
)

type Repository struct {
	*repo.BaseRepository
}

func NewRepository() *Repository {
	return &Repository{repo.NewBaseRepository()}
}

func (r *Repository) GetRegistrations(queryModel RegistrationQueryModel, formId uuid.UUID, regType string, userId uuid.UUID) (RegistrationList, error) {

	sqlString := `SELECT 
   r.id,
   u.first_name, 
   u.last_name, 
   u.email, 
   u.phone, 
   u.picture,
   r.created_at, 
   t.qty, 
   t.amount, 
   t.ref, 
   t.tx_id,
   t.tx_pro, 
   t.tx_cur, 
   t.tx_date,
   r.reg_desc,
   r.reg_ref, 
   r.reg_status,
   r.metadata,
	setweight(to_tsvector('english', 
		coalesce(u.first_name, '') || ' ' || 
		coalesce(u.last_name, '')
	), 'A') ||
	setweight(to_tsvector('english', 
		coalesce(u.email, '') || ' ' || 
		coalesce(u.phone, '') || ' ' ||
		coalesce(t.ref, '') || ' ' || 
		coalesce(t.tx_pro, '') || ' ' || 
		coalesce(cast(t.amount as text), '')
	), 'B') ||
	setweight(to_tsvector('english', 
		coalesce(r.reg_desc, '') || ' ' || 
		coalesce(r.reg_ref, '') || ' ' || 
		coalesce(r.reg_status, '')
	), 'C') as search_vector
   FROM (registrations r left join transactions t on t.id = r.transaction_id join forms f on f.id = r.form_id join users u on u.id = r.buyer_id)
   WHERE f.id = r.form_id and u.id = r.buyer_id
   ORDER BY r.created_at DESC`

	conditions := make(pager.Conditions)

	conditions["search"] = pager.SearchCondition{
		Search: queryModel.Search,
	}

	conditions["filter"] = pager.FilterCondition{
		Filter:  queryModel.Filter,
		Columns: []string{"r.reg_ref", "r.reg_status"},
	}

	conditions["custom"] = pager.CustomCondition{
		Items:  []string{"r.form_id = :form_id", "f.user_id = :user_id", "f.reg_type = :reg_type"},
		Params: map[string]any{"form_id": formId, "user_id": userId, "reg_type": regType},
	}

	pager := pager.NewPager[RegistrationItem]().SetQuery(sqlString).SetConditions(conditions).SetPage(queryModel.Page)

	result, err := pager.Query()

	if err != nil {
		return RegistrationList{}, err
	}

	if result.GetSize() == 0 && result.GetTotal() > 0 {
		result, err = pager.SetPage(queryModel.Page - 1).Query()

		if err != nil {
			return RegistrationList{}, err
		}
	}

	return RegistrationList{
		Items: result.GetItems(),
		Page:  result.GetPage(),
		Size:  result.GetSize(),
		Total: result.GetTotal(),
	}, nil
}

func (r *Repository) GetRegistrationStats(formId uuid.UUID, regType string, userId uuid.UUID) (RegistrationStats, error) {

	itemStats, err := r.getItemStats(formId, regType, userId)

	if err != nil {
		return RegistrationStats{}, err
	}

	saleStats, err := r.getSaleStats(formId, regType, userId)

	if err != nil {
		return RegistrationStats{}, err
	}

	return RegistrationStats{
		ItemStats: itemStats,
		SaleStats: saleStats,
	}, nil
}

func (r *Repository) UpdateRegistrationStatus(request UpdateRegistrationStatusRequest) (bool, error) {

	sqlString := `UPDATE registrations SET reg_status = :reg_status, updated_at = now() WHERE form_id = :form_id and reg_code = :reg_code and id = :id`

	res, err := r.DB.NamedExec(sqlString, request)

	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return false, err
	}

	if rows == 0 {
		return false, errors.New("no rows affected")
	}

	return true, nil
}

// func (r *Repository) UpdateTicketAttendeeStatus(request UpdateRegistrationStatusRequest) (bool, error) {

// 	sqlString := `
// 		UPDATE attendees
// 		SET
// 			checked_in = true,
// 			checked_at = now()
// 		WHERE id = $1 and lower(reg_ref) = lower($2)
// 	`

// 	res, err := r.DB.Exec(sqlString, request.RegId, request.RegRef)

// 	if err != nil {
// 		return false, err
// 	}

// 	rows, err := res.RowsAffected()

// 	if err != nil {
// 		return false, err
// 	}

// 	if rows == 0 {
// 		return false, errors.New("no rows affected")
// 	}

// 	return true, nil
// }

func (r *Repository) getItemStats(formId uuid.UUID, regType string, userId uuid.UUID) (ItemStats, error) {

	var itemStats ItemStats

	sqlString := `
		SELECT 
			COUNT(DISTINCT r.id) as total_registered,
			COUNT(DISTINCT r.id) FILTER (WHERE r.reg_status = 'checkedin') as total_checkedin,
			COUNT(DISTINCT r.id) FILTER (WHERE r.reg_status = 'cancelled') as total_cancelled
		FROM registrations r 
		JOIN forms f ON f.id = r.form_id
		WHERE 
		r.form_id = $1 AND
		f.reg_type = $2 AND
		f.user_id = $3
	`

	err := r.DB.QueryRow(sqlString, formId, regType, userId).Scan(&itemStats.TotalRegistered, &itemStats.TotalCheckedIn, &itemStats.TotalCancelled)

	if err != nil {

		fmt.Println("ItemStats Err: ", err.Error())
		return ItemStats{}, err
	}

	return itemStats, nil
}

func (r *Repository) getSaleStats(formId uuid.UUID, regType string, userId uuid.UUID) (SaleStats, error) {

	var saleStats SaleStats

	sqlString := `
		SELECT 
			SUM(t.amount) as total_amount,
			SUM(t.qty) as total_qty,
			SUM(t.refunded) as total_refunded
		FROM (registrations r JOIN transactions t ON t.id = r.transaction_id JOIN forms f ON f.id = r.form_id)
		WHERE 
		f.id = r.form_id AND
		t.id = r.transaction_id AND
		r.form_id = $1 AND
		f.reg_type = $2 AND
		f.user_id = $3
	`

	err := r.DB.QueryRow(sqlString, formId, regType, userId).Scan(&saleStats.TotalAmount, &saleStats.TotalQty, &saleStats.TotalRefunded)

	if err != nil {

		fmt.Println("SaleStats Err: ", err.Error())
		return SaleStats{}, err
	}

	return saleStats, nil

}

func (r *Repository) GetTicketAttendees(queryModel RegistrationQueryModel, formId uuid.UUID, userId uuid.UUID) (TicketAttendeeList, error) {

	sqlString := `SELECT 
			a.id,
			a.first_name,
			a.last_name,
			a.email,
			a.phone,
			a.reg_ref,
			t.name as ticket_name,
			a.created_at,
			a.checked_in,
			a.checked_at,
	setweight(to_tsvector('english', 
		coalesce(a.first_name, '') || ' ' || 
		coalesce(a.last_name, '')
	), 'A') ||
	setweight(to_tsvector('english', 
		coalesce(a.email, '') || ' ' || 
		coalesce(a.phone, '') || ' ' ||
		coalesce(a.reg_ref, '') || ' ' || 
		coalesce(t.name, '')
	), 'B')  as search_vector

  FROM (attendees a JOIN registrations r ON r.id = a.reg_id JOIN forms f ON f.id = r.form_id JOIN tickets t ON t.id = a.ticket_id) 
  WHERE 
  r.id = a.reg_id AND
  f.id = r.form_id AND
  t.id = a.ticket_id
  AND f.reg_type = 'ticket'
  ORDER BY a.created_at DESC`

	conditions := make(map[string]interface{})

	conditions["search"] = pager.SearchCondition{
		Search: queryModel.Search,
	}

	conditions["filter"] = pager.FilterCondition{
		Filter:  queryModel.Filter,
		Columns: []string{"a.reg_ref", "a.checked_in"},
	}

	conditions["custom"] = pager.CustomCondition{
		Items:  []string{"r.form_id = :form_id", "f.user_id = :user_id"},
		Params: map[string]any{"form_id": formId, "user_id": userId},
	}

	pager := pager.NewPager[TicketAttendee]().SetQuery(sqlString).SetConditions(conditions).SetPage(queryModel.Page)

	result, err := pager.Query()

	if err != nil {
		return TicketAttendeeList{}, err
	}

	return TicketAttendeeList{
		Items: result.GetItems(),
		Page:  result.GetPage(),
		Size:  result.GetSize(),
		Total: result.GetTotal(),
	}, nil
}

func (r *Repository) CountTicketCheckInsByAgent(formId uuid.UUID, agentName string) int {

	sqlString := `
			SELECT COUNT(DISTINCT r.id)
			FROM attendees a
			JOIN attendees a ON r.id = a.reg_id
			JOIN registrations r ON r.id = a.reg_id
			JOIN forms f ON f.id = r.form_id
			WHERE r.form_id = $1 AND a.checked_by = $2
	`

	var count int

	err := r.DB.QueryRow(sqlString, formId, agentName).Scan(&count)

	if err != nil {
		return 0
	}

	return count
}

func (r *Repository) CheckInTicketRegistration(regRef string, agentName string) (TicketAttendee, error) {

	var attendee TicketAttendee
	var count int

	sqlString := `SELECT COUNT(DISTINCT id) FROM attendees WHERE lower(reg_ref) = lower($1) AND checked_in = true`

	err := r.DB.QueryRow(sqlString, regRef).Scan(&count)

	if err != nil {
		return TicketAttendee{}, err
	}

	if count > 0 {
		return TicketAttendee{}, errors.New("already_checked_in")
	}

	sqlString = `UPDATE attendees a 
		SET checked_in = true, checked_at = now(), checked_by = $1 
		FROM tickets t
		WHERE lower(a.reg_ref) = lower($2) 
			AND t.id = a.ticket_id
		RETURNING a.first_name, a.email, a.phone, a.reg_ref, t.name as ticket_name, a.checked_in, a.checked_at`

	err = r.DB.QueryRow(sqlString, agentName, regRef).Scan(&attendee.FirstName, &attendee.Email, &attendee.Phone, &attendee.RegRef, &attendee.TicketName, &attendee.CheckedIn, &attendee.CheckedAt)

	if err != nil {
		return TicketAttendee{}, err
	}

	return attendee, nil
}

func (r *Repository) LookupTicketRegistration(formId uuid.UUID, regRef string) (*TicketAttendee, error) {

	sqlString := `SELECT 
		a.id,
		a.first_name,
		a.last_name, 
		a.email,
		a.phone,
		a.reg_ref,
		r.reg_status,
		t.name as ticket_name,
		a.checked_in,
		a.created_at,
		a.checked_at
	FROM attendees a 
	JOIN tickets t ON t.id = a.ticket_id
	JOIN registrations r ON r.id = a.reg_id 
	JOIN forms f ON f.id = r.form_id 
	WHERE lower(a.reg_ref) = lower($1) 
		AND t.id = a.ticket_id
		AND f.id = r.form_id 
		AND r.form_id = $2
		AND f.reg_type = 'ticket'`

	var attendee TicketAttendee

	err := r.DB.QueryRow(sqlString, regRef, formId).Scan(&attendee.Id, &attendee.FirstName, &attendee.LastName, &attendee.Email, &attendee.Phone, &attendee.RegRef, &attendee.RegStatus, &attendee.TicketName, &attendee.CheckedIn, &attendee.CreatedAt, &attendee.CheckedAt)

	if err != nil {
		return nil, err
	}

	return &attendee, nil
}
