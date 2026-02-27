package event

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/gbaski/gbaski-shared/config"
)

func InvalidateStorage(objectPath string) error {

	objectPath = strings.TrimPrefix(objectPath, "https://storage.gbaski.app")

	cfg := config.GetAwsConfig()
	cfClient := cloudfront.NewFromConfig(*cfg)

	input := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String("E21GP6LQUH5H9A"),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String(time.Now().Format(time.RFC3339)),
			Paths: &types.Paths{
				Quantity: aws.Int32(1),
				Items:    []string{objectPath},
			},
		},
	}

	_, err := cfClient.CreateInvalidation(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to create CloudFront invalidation: %w", err)
	}

	fmt.Println("CloudFront invalidation created")

	return nil
}
