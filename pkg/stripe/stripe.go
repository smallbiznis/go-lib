package stripe

import (
	"context"
	"errors"

	"github.com/smallbiznis/go-lib/pkg/env"
	st "github.com/stripe/stripe-go/v80"
	"github.com/stripe/stripe-go/v80/customer"
	"github.com/stripe/stripe-go/v80/subscription"
	"go.uber.org/fx"
)

var (
	Stripe = fx.Module("stripe", fx.Options(
		fx.Provide(New),
	))
)

func init() {
	st.Key = env.Lookup("STRIPE_SECRET_KEY", "sk_test_51LlDyWCIG3hXfWuZUvt4U1eu4ywFzTPj8eWePweMnXH9Bx96L1kCWbfmeFL0VsUI2TKgUUALxztYOcvbnHHyLyFB00fY5bYD4W")
}

type IStripe interface {
	CreateCustomer(context.Context, *st.CustomerParams) (*st.Customer, error)
	GetCustomer(context.Context, string, *st.CustomerParams) (*st.Customer, error)
	DeleteCustomer(context.Context, string, *st.CustomerParams) error

	CreateBillingSession(context.Context, *st.BillingPortalSession) (*st.BillingPortalSession, error)

	CreateSubscription(context.Context, *st.SubscriptionParams) (*st.Subscription, error)
	GetSubscription(context.Context, string, *st.SubscriptionParams) (*st.Subscription, error)
	ResumeSubscription(context.Context, string, *st.SubscriptionResumeParams) (*st.Subscription, error)
	CancelSubscription(context.Context, string, *st.SubscriptionCancelParams) error
}

type stripe struct{}

func New() IStripe {
	return &stripe{}
}

func (s *stripe) CreateCustomer(ctx context.Context, req *st.CustomerParams) (*st.Customer, error) {
	return customer.New(req)
}

func (s *stripe) GetCustomer(ctx context.Context, id string, params *st.CustomerParams) (*st.Customer, error) {
	return customer.Get(id, params)
}

func (s *stripe) DeleteCustomer(ctx context.Context, id string, params *st.CustomerParams) error {
	if _, err := customer.Del(id, params); err != nil {
		return err
	}
	return nil
}

func (s *stripe) CreateBillingSession(ctx context.Context, req *st.BillingPortalSession) (*st.BillingPortalSession, error) {
	return nil, errors.New("unimplement")
}

func (s *stripe) CreateSubscription(ctx context.Context, req *st.SubscriptionParams) (*st.Subscription, error) {
	return subscription.New(req)
}

func (s *stripe) GetSubscription(ctx context.Context, id string, params *st.SubscriptionParams) (*st.Subscription, error) {
	return subscription.Get(id, params)
}

func (s *stripe) ResumeSubscription(ctx context.Context, id string, params *st.SubscriptionResumeParams) (*st.Subscription, error) {
	return subscription.Resume(id, params)
}

func (s *stripe) CancelSubscription(ctx context.Context, id string, params *st.SubscriptionCancelParams) error {
	if _, err := subscription.Cancel(id, params); err != nil {
		return err
	}
	return nil
}
