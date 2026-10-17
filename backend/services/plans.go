package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"ats-backend/config"
	"ats-backend/models"

	"github.com/google/uuid"
)

// ErrPlanLimit marks a request refused because the company's plan doesn't allow it.
var ErrPlanLimit = errors.New("plan limit reached")

// DefaultTier is the free plan companies fall back to.
const DefaultTier = "starter"

// CompanyPlan loads the plan the company is entitled to right now. A tier bought through billing
// lapses to the default once its paid period ends; tiers set by a super admin (no subscription) don't.
func CompanyPlan(companyID uuid.UUID) (*models.SubscriptionPlan, error) {
	var company models.Company
	if err := config.DB.Select("subscription_tier").First(&company, "id = ?", companyID).Error; err != nil {
		return nil, err
	}
	tier := strings.ToLower(company.SubscriptionTier)
	if tier != DefaultTier {
		var latest models.Subscription
		err := config.DB.Where("company_id = ?", companyID).Order("current_period_end desc").Limit(1).Find(&latest).Error
		if err != nil {
			return nil, err
		}
		if latest.ID != uuid.Nil && latest.CurrentPeriodEnd.Before(time.Now()) {
			tier = DefaultTier
		}
	}
	var plan models.SubscriptionPlan
	if err := config.DB.Where("lower(name) = ?", tier).First(&plan).Error; err != nil {
		return nil, fmt.Errorf("plan %q: %w", tier, err)
	}
	return &plan, nil
}

// PlanUsage is what the company has used against its plan's limits.
type PlanUsage struct {
	OpenJobs          int64 `json:"open_jobs"`
	ApplicationsMonth int64 `json:"applications_this_month"`
	Admins            int64 `json:"admins"`
	PendingInvites    int64 `json:"pending_invites"`
}

func monthStart() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func CompanyUsage(companyID uuid.UUID) (PlanUsage, error) {
	var u PlanUsage
	err := config.DB.Raw(`select
		(select count(*) from jobs where company_id = @c and status = 'open' and deadline >= @today) as open_jobs,
		(select count(*) from applications where company_id = @c and applied_at >= @month) as applications_month,
		(select count(*) from admins where company_id = @c) as admins,
		(select count(*) from team_invites where company_id = @c and accepted_at is null and expires_at > now()) as pending_invites`,
		map[string]any{"c": companyID, "today": time.Now().UTC().Format("2006-01-02"), "month": monthStart()}).
		Scan(&u).Error
	return u, err
}

// The Check* helpers return ErrPlanLimit (wrapped with a user-facing reason) when the plan forbids the action.

func CheckCanOpenJob(companyID uuid.UUID) error {
	plan, err := CompanyPlan(companyID)
	if err != nil {
		return err
	}
	u, err := CompanyUsage(companyID)
	if err != nil {
		return err
	}
	if u.OpenJobs >= int64(plan.MaxJobs) {
		return fmt.Errorf("%w: your %s plan allows %d open jobs. Close a job or upgrade your plan", ErrPlanLimit, plan.Name, plan.MaxJobs)
	}
	return nil
}

func CheckCanAcceptApplication(companyID uuid.UUID) error {
	plan, err := CompanyPlan(companyID)
	if err != nil {
		return err
	}
	u, err := CompanyUsage(companyID)
	if err != nil {
		return err
	}
	if u.ApplicationsMonth >= int64(plan.MaxApplications) {
		return fmt.Errorf("%w: your %s plan allows %d applications per month. Upgrade to receive more", ErrPlanLimit, plan.Name, plan.MaxApplications)
	}
	return nil
}

func CheckCanScoreCVs(companyID uuid.UUID) error {
	plan, err := CompanyPlan(companyID)
	if err != nil {
		return err
	}
	if !plan.AIShortlisting {
		return fmt.Errorf("%w: CV scoring isn't included in the %s plan. Upgrade to use it", ErrPlanLimit, plan.Name)
	}
	return nil
}

func CheckCanAddMember(companyID uuid.UUID) error {
	plan, err := CompanyPlan(companyID)
	if err != nil {
		return err
	}
	u, err := CompanyUsage(companyID)
	if err != nil {
		return err
	}
	if u.Admins+u.PendingInvites >= int64(plan.MaxAdmins) {
		return fmt.Errorf("%w: your %s plan allows %d team member(s), including pending invites. Upgrade to add more", ErrPlanLimit, plan.Name, plan.MaxAdmins)
	}
	return nil
}
