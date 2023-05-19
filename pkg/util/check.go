package util

import (
	"fmt"

	nautescrd "github.com/nautes-labs/pkg/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IsLegal check resource is ready to used.
// When function a returns an error, it means that this resource should not be used anymore.
// This may be because the resource does not belong to the product
func IsLegal(res client.Object, productName string) error {
	if !res.GetDeletionTimestamp().IsZero() {
		return fmt.Errorf("resouce %s is terminating", res.GetName())
	}

	if !IsBelongsToProduct(res, productName) {
		return fmt.Errorf("resource %s is not belongs to product", res.GetName())
	}
	return nil
}

// IsBelongsToProduct check resouces is maintain by nautes
func IsBelongsToProduct(res client.Object, productName string) bool {
	if res == nil {
		return false
	}

	labels := res.GetLabels()
	name, ok := labels[nautescrd.LABEL_FROM_PRODUCT]
	if !ok || name != productName {
		return false
	}
	return true
}
