package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
	"github.com/gbaski/gbaski-platform/internal/payout/domain"
	"github.com/gbaski/gbaski-shared/pager"
	"github.com/gbaski/gbaski-shared/repo"
	"github.com/gbaski/gbaski-shared/util"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// PayoutRepositoryImpl implements contracts.PayoutRepository
type PayoutRepositoryImpl struct {
	baseRepo *repo.BaseRepository
}

// NewPayoutRepositoryImpl creates a new PayoutRepositoryImpl
// If db is nil, uses BaseRepository's default DB connection
func NewPayoutRepositoryImpl(db *sqlx.DB) contracts.PayoutRepository {
	baseRepo := repo.NewBaseRepository()
	if db != nil {
		baseRepo.DB = db
	}
	return &PayoutRepositoryImpl{baseRepo: baseRepo}
}

// GetBanks returns the list of all banks
func (r *PayoutRepositoryImpl) GetBanks(ctx context.Context) ([]domain.Bank, error) {
	type bankRow struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
		Code string `db:"code"`
	}

	var rows []bankRow
	err := r.baseRepo.DB.SelectContext(ctx, &rows, `SELECT id, bank_name AS name, cbn_code AS code FROM payout.banks`)
	if err != nil {
		return nil, err
	}

	banks := make([]domain.Bank, 0, len(rows))
	for _, row := range rows {
		banks = append(banks, domain.NewBank(row.ID, row.Name, row.Code))
	}
	return banks, nil
}

// SavePayoutAccount upserts a payout account and sets the returned ID on the aggregate
func (r *PayoutRepositoryImpl) SavePayoutAccount(ctx context.Context, account *domain.PayoutAccount) error {
	config := struct {
		InstantPayoutWeeklyLimit float64 `json:"instantPayoutWeeklyLimit"`
		SuspendPayout            struct {
			Enabled bool      `json:"enabled"`
			Reason  string    `json:"reason"`
			Until   time.Time `json:"until"`
		} `json:"suspendPayout"`
	}{
		InstantPayoutWeeklyLimit: 200000,
	}

	sqlString := `
        INSERT INTO payout.accounts (
            bank_id, account_number, account_name, currency, user_id, config
        ) VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (currency, user_id) DO UPDATE SET
            bank_id = $1,
            account_number = $2,
            account_name = $3
        RETURNING id`

	var id int
	err := r.baseRepo.DB.QueryRowContext(
		ctx, sqlString,
		account.BankID(),
		account.AccountNumber(),
		account.AccountName(),
		account.Currency().Value(),
		account.UserID(),
		util.JsonStringify(config),
	).Scan(&id)
	if err != nil {
		return err
	}

	account.SetID(id)
	return nil
}

// FindPayoutAccount fetches a payout account by currency and user ID
func (r *PayoutRepositoryImpl) FindPayoutAccount(ctx context.Context, currency string, userID uuid.UUID) (*domain.PayoutAccount, error) {
	type accountRow struct {
		ID            int    `db:"id"`
		AccountNumber string `db:"account_number"`
		AccountName   string `db:"account_name"`
		BankName      string `db:"bank_name"`
		BankID        int    `db:"bank_id"`
		PayoutType    string `db:"payout_type"`
	}

	var row accountRow
	err := r.baseRepo.DB.QueryRowContext(ctx,
		`SELECT a.id, a.account_number, a.account_name, b.bank_name, a.bank_id, a.payout_type
         FROM (payout.accounts a JOIN payout.banks b ON b.id = a.bank_id)
         WHERE a.currency = $1 AND a.user_id = $2 LIMIT 1`,
		currency, userID,
	).Scan(&row.ID, &row.AccountNumber, &row.AccountName, &row.BankName, &row.BankID, &row.PayoutType)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	cur, err := domain.NewCurrency(currency)
	if err != nil {
		return nil, err
	}

	pt, err := domain.NewPayoutType(row.PayoutType)
	if err != nil {
		return nil, err
	}

	return domain.ReconstructPayoutAccount(row.ID, userID, row.AccountNumber, row.AccountName, row.BankName, row.BankID, pt, cur), nil
}

// ChangePayoutType updates the payout type for an account and returns the updated aggregate
func (r *PayoutRepositoryImpl) ChangePayoutType(ctx context.Context, accountID int, userID uuid.UUID, payoutType domain.PayoutType) (*domain.PayoutAccount, error) {
	var currency string
	err := r.baseRepo.DB.QueryRowContext(ctx,
		`UPDATE payout.accounts SET payout_type = $1 WHERE id = $2 AND user_id = $3 RETURNING currency`,
		payoutType.Value(), accountID, userID,
	).Scan(&currency)
	if err != nil {
		return nil, err
	}

	return r.FindPayoutAccount(ctx, currency, userID)
}

// GetPayoutSummary returns a map of payout summaries grouped by payout type
func (r *PayoutRepositoryImpl) GetPayoutSummary(ctx context.Context, formID, userID uuid.UUID) (map[string]contracts.PayoutSummaryDTO, error) {
	sqlString := `
    SELECT
      le.payout_type AS "payoutType",
      MIN(CASE WHEN le.payout_status = 'pending' THEN le.payout_date END) AS "nextPayoutDate",
      SUM(CASE WHEN le.payout_status = 'pending' THEN le.amount ELSE 0 END) AS "pendingAmount",
      SUM(CASE WHEN le.payout_status = 'processing' THEN le.amount ELSE 0 END) AS "processingAmount"
    FROM payout.ledger le
    JOIN forms f ON f.id = le.form_id
    JOIN users u ON u.id = f.user_id
    WHERE le.form_id = $1
      AND u.id = $2 AND u.account_type = 'organizer'
      AND le.entry_type = 'seller'
    GROUP BY le.payout_type`

	type summaryRow struct {
		PayoutType       string     `db:"payoutType"`
		NextPayoutDate   *time.Time `db:"nextPayoutDate"`
		PendingAmount    float32    `db:"pendingAmount"`
		ProcessingAmount float32    `db:"processingAmount"`
	}

	rows, err := r.baseRepo.DB.QueryContext(ctx, sqlString, formID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]contracts.PayoutSummaryDTO)
	for rows.Next() {
		var row summaryRow
		if err := rows.Scan(&row.PayoutType, &row.NextPayoutDate, &row.PendingAmount, &row.ProcessingAmount); err != nil {
			return nil, err
		}
		pt, _ := domain.NewPayoutType(row.PayoutType)
		result[row.PayoutType] = contracts.PayoutSummaryDTO{
			PayoutType:       pt,
			NextPayoutDate:   row.NextPayoutDate,
			PendingAmount:    row.PendingAmount,
			ProcessingAmount: row.ProcessingAmount,
		}
	}

	// Guarantee both types are present
	for _, t := range []string{"instant", "weekly"} {
		if _, ok := result[t]; !ok {
			pt, _ := domain.NewPayoutType(t)
			result[t] = contracts.PayoutSummaryDTO{
				PayoutType:       pt,
				NextPayoutDate:   nil,
				PendingAmount:    0,
				ProcessingAmount: 0,
			}
		}
	}

	return result, nil
}

// GetPayoutHistory returns a paginated list of payout history items
func (r *PayoutRepositoryImpl) GetPayoutHistory(ctx context.Context, params contracts.QueryParams, formID, userID uuid.UUID) (contracts.PayoutList, error) {
	sqlString := `
    SELECT p.amount, p.tx_ref, p.paid_at
    FROM (payout.payouts p
        JOIN forms f ON f.id = p.form_id
        JOIN users u ON u.id = f.user_id)
    WHERE
        f.id = p.form_id AND
        u.id = f.user_id AND
        p.form_id = :form_id AND
        u.id = :user_id
    ORDER BY p.paid_at DESC`

	pg := pager.Params{
		"form_id": formID,
		"user_id": userID,
	}

	result, err := pager.NewPager[contracts.PayoutItem]().
		SetQuery(sqlString).
		SetParams(pg).
		SetPage(params.Page).
		Query()
	if err != nil {
		return contracts.PayoutList{}, err
	}

	return contracts.PayoutList{
		Items: result.GetItems(),
		Page:  result.GetPage(),
		Size:  result.GetSize(),
		Total: result.GetTotal(),
	}, nil
}

// GetTotalPayout returns the total paid out for a form and user
func (r *PayoutRepositoryImpl) GetTotalPayout(ctx context.Context, userID, formID uuid.UUID) float64 {
	var total float64
	err := r.baseRepo.DB.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(p.amount), 0) FROM (payout.payouts p join forms f on f.id = p.form_id) WHERE f.user_id = $1 AND p.form_id = $2`,
		userID, formID,
	).Scan(&total)
	if err != nil {
		return 0
	}
	return total
}

// SavePaymentProvider upserts a payment provider record
func (r *PayoutRepositoryImpl) SavePaymentProvider(ctx context.Context, provider *domain.PaymentProvider) error {
	_, err := r.baseRepo.DB.ExecContext(ctx,
		`INSERT INTO payout.providers (provider_name, secret_key, public_key, user_id)
         VALUES ($1, $2, $3, $4)
         ON CONFLICT (user_id) DO UPDATE SET provider_name = $1, secret_key = $2, public_key = $3`,
		provider.Provider(), provider.SecretKey(), provider.PublicKey(), provider.UserID(),
	)
	return err
}

// FindPaymentProviderBySellerName fetches a payment provider by seller (user) name
func (r *PayoutRepositoryImpl) FindPaymentProviderBySellerName(ctx context.Context, sellerName string) (*domain.PaymentProvider, error) {
	type providerRow struct {
		Provider  string `db:"provider_name"`
		SecretKey string `db:"secret_key"`
		PublicKey string `db:"public_key"`
	}

	var row providerRow
	err := r.baseRepo.DB.QueryRowContext(ctx,
		`SELECT p.provider_name, p.secret_key, p.public_key
         FROM (payout.providers p join users u on u.id = p.user_id)
         WHERE u.id = p.user_id AND u.name = $1`,
		sellerName,
	).Scan(&row.Provider, &row.SecretKey, &row.PublicKey)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return domain.ReconstructPaymentProvider(uuid.Nil, row.Provider, row.PublicKey, row.SecretKey), nil
}

// FindPaymentProviderByUserID fetches a payment provider by user ID
func (r *PayoutRepositoryImpl) FindPaymentProviderByUserID(ctx context.Context, userID uuid.UUID) (*domain.PaymentProvider, error) {
	type providerRow struct {
		UserID    uuid.UUID `db:"id"`
		Provider  string    `db:"provider_name"`
		SecretKey string    `db:"secret_key"`
		PublicKey string    `db:"public_key"`
	}

	var row providerRow
	err := r.baseRepo.DB.QueryRowContext(ctx,
		`SELECT u.id, p.provider_name, p.secret_key, p.public_key
         FROM (payout.providers p join users u on u.id = p.user_id)
         WHERE u.id = p.user_id AND u.id = $1`,
		userID,
	).Scan(&row.UserID, &row.Provider, &row.SecretKey, &row.PublicKey)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return domain.ReconstructPaymentProvider(row.UserID, row.Provider, row.PublicKey, row.SecretKey), nil
}
