package payout

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/gbaski/gbaski-shared/pager"
	"github.com/gbaski/gbaski-shared/repo"
	"github.com/gbaski/gbaski-shared/util"
)

type Repository struct {
	*repo.BaseRepository
}

func NewRepository() *Repository {
	return &Repository{repo.NewBaseRepository()}
}

func (r *Repository) GetPayoutAccount(currency string, userId uuid.UUID) (*PayoutAccount, error) {

	sqlString := `SELECT a.id, a.account_number, a.account_name, b.bank_name, a.bank_id, a.payout_type FROM (payout.accounts a JOIN payout.banks b ON b.id = a.bank_id) WHERE a.currency = $1 AND a.user_id = $2 LIMIT 1`

	var payoutAccount PayoutAccount

	payoutAccount.Currency = currency

	err := r.DB.QueryRow(sqlString, currency, userId).Scan(&payoutAccount.Id, &payoutAccount.AccountNumber, &payoutAccount.AccountName, &payoutAccount.BankName, &payoutAccount.BankId, &payoutAccount.PayoutType)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	payoutAccount.FeeRates.Instant = 2
	payoutAccount.FeeRates.Weekly = 0

	return &payoutAccount, nil
}

func (r *Repository) GetBanks() ([]Bank, error) {

	sqlString := `SELECT id,bank_name as name, cbn_code as code FROM payout.banks`

	var banks []Bank

	err := r.DB.Select(&banks, sqlString)

	if err != nil {
		return nil, err
	}

	return banks, nil
}

func (r *Repository) CreatePayoutAccount(data CreatePayoutAccountRequest) (*int16, error) {

	config := PayoutConfig{
		InstantPayoutWeeklyLimit: 200000,
		SuspendPayout: SuspendPayout{
			Enabled: false,
			Reason:  "",
			Until:   time.Time{},
		},
	}

	sqlString := `
        INSERT INTO payout.accounts (
			bank_id,
			account_number,
			account_name,
			currency,
			user_id,
			config
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) ON CONFLICT (currency, user_id) DO UPDATE SET
			bank_id = $1,
			account_number = $2,
			account_name = $3
		RETURNING id`

	var id int16

	err := r.DB.QueryRow(sqlString, data.BankId, data.AccountNumber, data.AccountName, data.Currency, data.UserId, util.JsonStringify(config)).Scan(&id)

	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (r *Repository) GetPayoutSummary(formId uuid.UUID, userId uuid.UUID) (map[string]PayoutSummary, error) {

	sqlString := `SELECT
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

	var items []PayoutSummary

	rows, err := r.DB.Query(sqlString, formId, userId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var summary PayoutSummary
		err := rows.Scan(&summary.PayoutType, &summary.NextPayoutDate, &summary.PendingAmount, &summary.ProcessingAmount)
		if err != nil {
			return nil, err
		}
		items = append(items, summary)
	}

	result := make(map[string]PayoutSummary)
	for _, payoutSummary := range items {
		result[string(payoutSummary.PayoutType)] = payoutSummary
	}

	if _, ok := result["instant"]; !ok {
		result["instant"] = PayoutSummary{
			PayoutType:       "instant",
			NextPayoutDate:   nil,
			PendingAmount:    0,
			ProcessingAmount: 0,
		}
	}

	if _, ok := result["weekly"]; !ok {
		result["weekly"] = PayoutSummary{
			PayoutType:       "weekly",
			NextPayoutDate:   nil,
			PendingAmount:    0,
			ProcessingAmount: 0,
		}
	}

	return result, nil

}

func (r *Repository) ChangePayoutType(data ChangePayoutTypeRequest) (*PayoutAccount, error) {

	sqlString := `UPDATE payout.accounts SET payout_type = $1 WHERE id = $2 AND user_id = $3 returning currency`

	var currency string

	err := r.DB.QueryRow(sqlString, data.PayoutType, data.PayoutAccountId, data.UserId).Scan(&currency)

	if err != nil {
		return nil, err
	}

	payoutAccount, _ := r.GetPayoutAccount(currency, data.UserId)

	return payoutAccount, nil
}

func (r *Repository) GetPayoutHistory(queryModel PayoutQueryModel, formId uuid.UUID, userId uuid.UUID) (PayoutList, error) {

	sqlString := `SELECT 
		p.amount,
		p.tx_ref,
		p.paid_at
	FROM (payout.payouts p 
		JOIN forms f ON f.id = p.form_id 
		JOIN users u ON u.id = f.user_id) 
	WHERE 
	f.id = p.form_id AND
	u.id = f.user_id AND
	p.form_id = :form_id AND
	u.id = :user_id 
	ORDER BY p.paid_at DESC`

	params := pager.Params{
		"form_id": formId,
		"user_id": userId,
	}

	result, err := pager.NewPager[PayoutItem]().
		SetQuery(sqlString).
		SetParams(params).
		SetPage(queryModel.Page).
		Query()

	if err != nil {
		return PayoutList{}, err
	}

	return PayoutList{
		Items: result.GetItems(),
		Page:  result.GetPage(),
		Size:  result.GetSize(),
		Total: result.GetTotal(),
	}, nil

}

func (r *Repository) CreatePaymentPayout(data CreatePaymentPayoutRequest) error {

	sqlString := `INSERT INTO payout.providers (provider_name, secret_key, public_key, user_id) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id) DO UPDATE SET provider_name = $1, secret_key = $2, public_key = $3`

	_, err := r.DB.Exec(sqlString, string(data.Provider), data.SecretKey, data.PublicKey, data.UserId)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetPaymentPayoutBySellerName(sellerName string) (*PaymentPayout, error) {

	sqlString := `SELECT p.provider_name, p.secret_key, p.public_key FROM (payout.providers p join users u on u.id = p.user_id) WHERE u.id = p.user_id AND u.name = $1`

	var paymentPayout PaymentPayout

	err := r.DB.QueryRow(sqlString, sellerName).Scan(&paymentPayout.Provider, &paymentPayout.SecretKey, &paymentPayout.PublicKey)

	if err != nil {
		return nil, err
	}

	return &paymentPayout, nil
}

func (r *Repository) GetPaymentPayoutByUserId(userId uuid.UUID) (*PaymentPayout, error) {

	sqlString := `SELECT u.id, p.provider_name, p.secret_key, p.public_key FROM (payout.providers p join users u on u.id = p.user_id) WHERE u.id = p.user_id AND u.id = $1`

	var paymentPayout PaymentPayout

	err := r.DB.QueryRow(sqlString, userId).Scan(&paymentPayout.UserId, &paymentPayout.Provider, &paymentPayout.SecretKey, &paymentPayout.PublicKey)

	if err != nil {
		return nil, err
	}

	return &paymentPayout, nil
}

func (r *Repository) GetTotalPayout(userId uuid.UUID, formId uuid.UUID) float64 {

	sqlString := `SELECT SUM(p.amount) FROM (payout.payouts p join forms f on f.id = p.form_id) WHERE f.user_id = $1 AND p.form_id = $2`

	var total float64

	err := r.DB.QueryRow(sqlString, userId, formId).Scan(&total)

	if err != nil {
		return 0
	}

	return total
}
