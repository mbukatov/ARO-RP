package applens

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
)

// AppLensClient is a minimal interface for azure AppLensClient
type AppLensClient interface {
	GetDetector(ctx context.Context, o *GetDetectorOptions) (*ResponseMessageEnvelope, error)
	ListDetectors(ctx context.Context, o *ListDetectorsOptions) (*ResponseMessageCollectionEnvelope, error)
}

type appLensClient struct {
	*Client
}
