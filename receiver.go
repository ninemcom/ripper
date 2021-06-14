package ripper

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ninemcom/ripper/opt"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
	"net/http"
	"time"
)

type Platform string

type configuration interface {
	ID() string
	Auth() string
}

type PlayStoreVoidedPurchasesListResponse androidpublisher.VoidedPurchasesListResponse

type PlayStoreHandler func(voided *PlayStoreVoidedPurchasesListResponse, err error) error

type AppStoreNotiType string

type AppStoreVoidedPurchasesListResponse struct {
	NotificationType  AppStoreNotiType          `json:"notification_type"`
	BundleID          string                    `json:"bid"`
	Password          string                    `json:"password"`
	LatestReceiptInfo AppStoreLatestReceiptInfo `json:"latest_receipt_info"`
}

type AppStoreLatestReceiptInfo struct {
	OriginalTransactionID string `json:"original_transaction_id"`
	ProductID             string `json:"product_id"`
	Reason                string `json:"cancellation_reason"` // 1: refund requested by user
	CancellationDateMS    string `json:"cancellation_ms"`
	InAppOwnershipType    string `json:"in_app_ownership_type"`
}

type AppStoreHandler func(voided *AppStoreVoidedPurchasesListResponse, err error) error

type Receiver struct {
	ctx    context.Context
	config configuration
}

func NewReceiver(ctx context.Context, config configuration) *Receiver {
	return &Receiver{
		ctx:    ctx,
		config: config,
	}
}

// NOTE: this function serves all voided purchases all the time. you need to check whether the order duplicated.
func (sub *Receiver) ReceivePlayStore(handler PlayStoreHandler, interval time.Duration, opts ...opt.Option) error {
	config := sub.config

	if config == nil {
		return ErrConfigNil
	}

	if interval == 0 {
		interval = 60 * time.Second
	}

	andPubService, err := androidpublisher.NewService(sub.ctx, option.WithCredentialsFile(config.Auth()))
	if err != nil {
		return err
	}

	go func() {
		defer defaultRecover()

		pubService := androidpublisher.NewPurchasesVoidedpurchasesService(andPubService)

		//
		for {
			select {
			case <-sub.ctx.Done():
				break
			default:
				voideds := []*PlayStoreVoidedPurchasesListResponse{}
				errs := []error{}

				// initial request if empty
				nextPageToken := ""
				for {
					listReq := pubService.List(config.ID())
					for i := range opts {
						opts[i].Apply(listReq)
					}

					if nextPageToken != "" {
						listReq.Token(nextPageToken)
					}

					res, err := listReq.Do()
					if err != nil {
						errs = append(errs, err)
						break
					}

					voided := (*PlayStoreVoidedPurchasesListResponse)(res)
					voideds = append(voideds, voided)

					if voided.TokenPagination != nil {
						nextPageToken = voided.TokenPagination.NextPageToken
					} else {
						break
					}
				}

				for i := range voideds {
					handler(voideds[i], nil)
				}

				for i := range errs {
					handler(nil, errs[i])
				}

				time.Sleep(interval)
			}
		}
	}()

	return nil
}

func (sub *Receiver) ReceiveAppStore(handler AppStoreHandler, addr string) {
	go func() {
		defer defaultRecover()

		http.HandleFunc("/appstore", func(w http.ResponseWriter, r *http.Request) {
			appStoreRes := AppStoreVoidedPurchasesListResponse{}
			err := json.NewDecoder(r.Body).Decode(&appStoreRes)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				_ = handler(nil, err)
				return
			}

			if err := handler(&appStoreRes, err); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		})
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic(err)
		}
	}()
}

func defaultRecover() {
	if v := recover(); v != nil {
		fmt.Println(v)
	}
}
