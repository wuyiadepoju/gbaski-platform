package repo

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/registration/contracts"
	"github.com/gbaski/gbaski-platform/internal/registration/domain"
	"github.com/gbaski/gbaski-shared/pager"
	"github.com/gbaski/gbaski-shared/repo"
	"github.com/google/uuid"
)

type RegistrationRepositoryImpl struct {
	*repo.BaseRepository
}

func NewRegistrationRepositoryImpl() contracts.RegistrationRepository {
	return &RegistrationRepositoryImpl{repo.NewBaseRepository()}
}

func (r *RegistrationRepositoryImpl) GetRegistrations(ctx context.Context, query domain.RegistrationQueryModel, formID uuid.UUID, regType string, userID uuid.UUID) (domain.RegistrationList, error) {
	sqlString := `SELECT 
   r.id,
   u.first_name, 
   u.last_name, 
   u.email, 
   u.phone, 
   r.created_at, 
   r.reg_ref
   FROM registrations r 
   JOIN forms f ON f.id = r.form_id 
   JOIN users u ON u.id = r.buyer_id
   WHERE f.id = :form_id AND f.user_id = :user_id AND f.reg_type = :reg_type
   ORDER BY r.created_at DESC`

	params := map[string]any{
		"form_id":  formID,
		"user_id":  userID,
		"reg_type": regType,
	}

	// This is a simplified version of the original query to fit the domain model
	// The original had complex fuzzy search which we can re-add if needed.
	p := pager.NewPager[domain.RegistrationItem]().
		SetQuery(sqlString).
		SetParams(params).
		SetPage(query.Page)

	result, err := p.Query()
	if err != nil {
		return domain.RegistrationList{}, err
	}

	return domain.RegistrationList{
		Items: result.GetItems(),
		Page:  result.GetPage(),
		Size:  result.GetSize(),
		Total: result.GetTotal(),
	}, nil
}

func (r *RegistrationRepositoryImpl) GetRegistrationStats(ctx context.Context, formID uuid.UUID, regType string, userID uuid.UUID) (domain.RegistrationStats, error) {
	var stats domain.RegistrationStats

	sqlString := `
		SELECT 
			COUNT(DISTINCT r.id) as total_registered,
			COUNT(DISTINCT r.id) FILTER (WHERE r.reg_status = 'checkedin') as total_checkedin,
			COUNT(DISTINCT r.id) FILTER (WHERE r.reg_status = 'cancelled') as total_cancelled
		FROM registrations r 
		JOIN forms f ON f.id = r.form_id
		WHERE r.form_id = $1 AND f.reg_type = $2 AND f.user_id = $3
	`

	err := r.DB.QueryRowContext(ctx, sqlString, formID, regType, userID).Scan(&stats.TotalRegistered, &stats.TotalCheckedIn, &stats.TotalCancelled)
	if err != nil {
		return domain.RegistrationStats{}, err
	}

	return stats, nil
}

func (r *RegistrationRepositoryImpl) GetTicketAttendees(ctx context.Context, query domain.RegistrationQueryModel, formID uuid.UUID, userID uuid.UUID) (domain.TicketAttendeeList, error) {
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
			a.checked_at
	  FROM attendees a 
	  JOIN registrations r ON r.id = a.reg_id 
	  JOIN forms f ON f.id = r.form_id 
	  JOIN tickets t ON t.id = a.ticket_id
	  WHERE r.form_id = :form_id AND f.user_id = :user_id
	  ORDER BY a.created_at DESC`

	params := map[string]any{
		"form_id": formID,
		"user_id": userID,
	}

	p := pager.NewPager[domain.Attendee]().
		SetQuery(sqlString).
		SetParams(params).
		SetPage(query.Page)

	result, err := p.Query()
	if err != nil {
		return domain.TicketAttendeeList{}, err
	}

	return domain.TicketAttendeeList{
		Items: result.GetItems(),
		Page:  result.GetPage(),
		Size:  result.GetSize(),
		Total: result.GetTotal(),
	}, nil
}

func (r *RegistrationRepositoryImpl) LookupTicketRegistration(ctx context.Context, formID uuid.UUID, regRef string) (*domain.Attendee, error) {
	sqlString := `SELECT 
		a.id,
		a.reg_id,
		a.first_name,
		a.last_name, 
		a.email,
		a.phone,
		a.reg_ref,
		t.name as ticket_name,
		a.checked_in,
		a.created_at,
		a.checked_at
	FROM attendees a 
	JOIN tickets t ON t.id = a.ticket_id
	JOIN registrations r ON r.id = a.reg_id 
	WHERE lower(a.reg_ref) = lower($1)`

	var attendee domain.Attendee
	err := r.DB.GetContext(ctx, &attendee, sqlString, regRef)
	if err != nil {
		return nil, domain.ErrAttendeeNotFound
	}

	return &attendee, nil
}

func (r *RegistrationRepositoryImpl) SaveAttendee(ctx context.Context, attendee *domain.Attendee) error {
	sqlString := `UPDATE attendees SET checked_in = :checked_in, checked_at = :checked_at, checked_by = :checked_by WHERE id = :id`
	_, err := r.DB.NamedExecContext(ctx, sqlString, attendee)
	return err
}
