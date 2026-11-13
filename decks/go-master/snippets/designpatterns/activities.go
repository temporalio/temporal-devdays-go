package designpatterns

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"
)

// RFIAPI is the reporting service. The tests point it at a local server.
var RFIAPI = "https://rfi.example.com"

// emailService and db are in-memory stand-ins for the external services, so
// the tests can make them fail and count what they did.
var emailService = &fakeEmail{}

type fakeEmail struct {
	mu           sync.Mutex
	failuresLeft int
	sent         []string
}

func (e *fakeEmail) Send(ctx context.Context, to, body string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.failuresLeft > 0 {
		e.failuresLeft--
		return errors.New("SMTP timeout")
	}
	e.sent = append(e.sent, to+": "+body)
	return nil
}

func (e *fakeEmail) SendWelcomeEmail(ctx context.Context, to string) error {
	return e.Send(ctx, to, "Welcome!")
}

var db = &fakeDB{}

type fakeDB struct {
	mu   sync.Mutex
	rows []User
}

func (d *fakeDB) AddUser(ctx context.Context, data UserData) (User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	user := User{ID: len(d.rows) + 1, Email: data.Email}
	d.rows = append(d.rows, user)
	return user, nil
}

// part: 04-design-patterns.md "is worth testing in isolation" 1
// activities.go
var rfiID = regexp.MustCompile(`^RFI-\d{4}$`)

func ValidateEligibility(ctx context.Context, req RFIRequest) (bool, error) {
	rfiIDValid := rfiID.MatchString(req.RFIID)
	submittedRecently := req.SubmittedAt.After(time.Now().AddDate(0, 0, -30))
	withinCostThreshold := req.EstimatedCost < 50_000

	return rfiIDValid && submittedRecently && withinCostThreshold, nil
}

// end part

// part: 04-design-patterns.md "can fail on its own" 1
// activities.go
func NotifyReviewer(ctx context.Context, req RFIRequest, reviewer string) error {
	return emailService.Send(ctx, reviewer, fmt.Sprintf("RFI %s requires your review", req.RFIID))
}

// end part

// part: 04-design-patterns.md "takes significant time" 1
// activities.go
func GenerateAnnualReport(ctx context.Context, req RFIRequest) error {
	// compiles all RFIs for the year into a report
	resp, err := http.Get(fmt.Sprintf("%s/rfi/%s/annual", RFIAPI, req.ProjectID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// end part

func RunCreditCheck(ctx context.Context, applicant Applicant) (CreditCheck, error) {
	if applicant.Income > 50_000 {
		return CreditCheck{Score: 720}, nil
	}
	return CreditCheck{Score: 580}, nil
}

func CheckFraud(ctx context.Context, applicant Applicant, credit CreditCheck) (FraudCheck, error) {
	return FraudCheck{Flagged: credit.Score < 600}, nil
}

func MakeDecision(ctx context.Context, applicant Applicant, credit CreditCheck, fraud FraudCheck) (Decision, error) {
	if credit.Score >= 650 && !fraud.Flagged {
		return "approved", nil
	}
	return "declined", nil
}

func GetApproval(ctx context.Context, claim Claim) (Approval, error) {
	return Approval{ClaimID: claim.ClaimID, Approved: claim.Amount <= 10_000}, nil
}

// slide: 04-design-patterns.md "When to split an Activity"
// ❌ One Activity doing two things — not safe to retry
func CreateUser(ctx context.Context, data UserData) (User, error) {
	user, err := db.AddUser(ctx, data)
	if err != nil {
		return User{}, err
	}
	if err := emailService.SendWelcomeEmail(ctx, user.Email); err != nil {
		return User{}, err
	}

	return user, nil
}

// end slide

// StandardPricing and PremiumPricing back the "Make it an Activity" note that
// reading the environment in a Workflow breaks replay.
func StandardPricing(ctx context.Context) (string, error) { return "standard", nil }
func PremiumPricing(ctx context.Context) (string, error)  { return "premium", nil }
