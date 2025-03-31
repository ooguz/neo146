package models

// BuyMeACoffeeWebhook represents the webhook payload from Buy Me a Coffee
type BuyMeACoffeeWebhook struct {
	Type     string `json:"type"`
	LiveMode bool   `json:"live_mode"`
	Attempt  int    `json:"attempt"`
	Created  int64  `json:"created"`
	EventID  int    `json:"event_id"`
	Data     struct {
		ID                 int     `json:"id"`
		Amount             float64 `json:"amount"`
		Object             string  `json:"object"`
		Paused             string  `json:"paused"`
		Status             string  `json:"status"`
		Canceled           string  `json:"canceled"`
		Currency           string  `json:"currency"`
		PspID              string  `json:"psp_id"`
		DurationType       string  `json:"duration_type"`
		StartedAt          int64   `json:"started_at"`
		CanceledAt         *int64  `json:"canceled_at"`
		NoteHidden         bool    `json:"note_hidden"`
		SupportNote        *string `json:"support_note"`
		SupporterName      string  `json:"supporter_name"`
		SupporterID        int     `json:"supporter_id"`
		SupporterEmail     string  `json:"supporter_email"`
		CurrentPeriodEnd   int64   `json:"current_period_end"`
		CurrentPeriodStart int64   `json:"current_period_start"`
		SupporterFeedback  *string `json:"supporter_feedback"`
		CancelAtPeriodEnd  *string `json:"cancel_at_period_end"`
	} `json:"data"`
}

// PayPalIPN represents the PayPal IPN notification structure
type PayPalIPN struct {
	PaymentStatus    string  `json:"payment_status"`
	PaymentType      string  `json:"payment_type"`
	PaymentDate      string  `json:"payment_date"`
	PaymentGross     float64 `json:"mc_gross,string"`
	PaymentFee       float64 `json:"mc_fee,string"`
	Currency         string  `json:"mc_currency"`
	PayerEmail       string  `json:"payer_email"`
	PayerID          string  `json:"payer_id"`
	SubscriptionID   string  `json:"subscr_id"`
	SubscriptionDate string  `json:"subscr_date"`
	SubscriptionEnd  string  `json:"subscr_end"`
	Custom           string  `json:"custom"`
	IPNType          string  `json:"txn_type"`
}

// PayPalIPNResponse represents the response we send back to PayPal
type PayPalIPNResponse struct {
	Status string
}
