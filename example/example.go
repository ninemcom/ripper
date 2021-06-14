package example

import (
	"context"
	"fmt"
	"github.com/ninemcom/ripper"
	"github.com/ninemcom/ripper/opt"
	"log"
	"os"
	"strconv"
	"time"
)

const GCloudGooglePlayServiceAccountPath = "/etc/secret-data/googleplay-api-cred.json"
const GCloudServiceAccountPath = "/etc/secret-data/gcloud-cred.json"
const ResourcesBucket = "resources"

var CouchbaseAddr = os.ExpandEnv("$COUCHBASE_URL")
var BigQuerySet = os.ExpandEnv("$BQ_DATASET")
var Project = os.ExpandEnv("$PROJECT")
var Region = os.ExpandEnv("$REGION")

func main() {
	gpPackage := "com..."
	Interval := "600"

	interval, err := strconv.Atoi(Interval)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	obs := ripper.NewReceiver(ctx, &ripper.PlayStoreConfiguration{
		PackageName:                  gpPackage,
		GCloudServiceAccountFilePath: GCloudGooglePlayServiceAccountPath,
	})

	dupMap := map[string]bool{}
	err = obs.ReceivePlayStore(func(voided *ripper.PlayStoreVoidedPurchasesListResponse, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		log.Println("got", len(voided.VoidedPurchases), "voided")
		for i := range voided.VoidedPurchases {
			if dupMap[voided.VoidedPurchases[i].OrderId] {
				continue
			}

			// Block the user
		}

		return nil
	}, time.Duration(interval)*time.Second, opt.WithMaxResults(1000))

	obs.ReceiveAppStore(func(voided *ripper.AppStoreVoidedPurchasesListResponse, err error) error {
		if voided.NotificationType == "REFUND" && voided.LatestReceiptInfo.InAppOwnershipType == "PURCHASED" {
			// Block the user
		}

		return nil
	}, ":8888")

	if err != nil {
		panic(err)
	}

	<-ctx.Done()
}
