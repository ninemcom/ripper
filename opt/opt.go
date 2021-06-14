package opt

import (
	"google.golang.org/api/androidpublisher/v3"
	"time"
)

type Option interface {
	Apply(list *androidpublisher.PurchasesVoidedpurchasesListCall)
}

type basicOption struct {
	apply func(list *androidpublisher.PurchasesVoidedpurchasesListCall)
}

func (opt *basicOption) Apply(list *androidpublisher.PurchasesVoidedpurchasesListCall) {
	opt.apply(list)
}

func WithStartTime(time time.Time) Option {
	return &basicOption{apply: func(list *androidpublisher.PurchasesVoidedpurchasesListCall) {
		list.StartTime(time.Unix())
	}}
}

func WithEndTime(time time.Time) Option {
	return &basicOption{apply: func(list *androidpublisher.PurchasesVoidedpurchasesListCall) {
		list.EndTime(time.Unix())
	}}
}

func WithMaxResults(max int64) Option {
	return &basicOption{apply: func(list *androidpublisher.PurchasesVoidedpurchasesListCall) {
		list.MaxResults(max)
	}}
}

func WithTokenFrom(token string) Option {
	return &basicOption{apply: func(list *androidpublisher.PurchasesVoidedpurchasesListCall) {
		list.Token(token)
	}}

}
