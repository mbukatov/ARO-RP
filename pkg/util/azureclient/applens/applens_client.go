package applens

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.
// AppLens Client created from CosmosDB Client
// (https://github.com/Azure/azure-sdk-for-go/blob/3f7acd20691214ef2cb1f0132f82115f1df01a8c/sdk/data/azcosmos/cosmos_client.go)

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

// AppLens client is used to interact with the Azure AppLens service.
type Client struct {
	endpoint string
	pipeline runtime.Pipeline
}

type ResponseMessageEnvelope struct {
	Id         string      `json:"id,omitempty"`
	Name       string      `json:"name,omitempty"`
	Type       string      `json:"type,omitempty"`
	Location   string      `json:"location,omitempty"`
	Properties interface{} `json:"properties,omitempty"`
}

type ResponseMessageCollectionEnvelope struct {
	Value []ResponseMessageEnvelope `json:"value,omitempty"`
}

// ListDetectors obtains the list of detectors for a service from AppLens.
// ctx - The context for the request.
// o - Options for Read operation.
func (c *Client) ListDetectors(
	ctx context.Context,
	o *ListDetectorsOptions) (*ResponseMessageCollectionEnvelope, error) {
	if o == nil {
		o = &ListDetectorsOptions{}
	}

	azResponse, err := c.sendPostRequest(
		ctx,
		o,
		nil)
	if err != nil {
		return nil, err
	}

	defer azResponse.Body.Close()

	bodyJson, err := io.ReadAll(azResponse.Body)

	if err != nil {
		return nil, err
	}

	return newResponseMessageCollectionEnvelope(bodyJson, o.ResourceID, o.Location)
}

// GetDetector obtains detector information from AppLens.
// ctx - The context for the request.
// o - Options for Read operation.
func (c *Client) GetDetector(
	ctx context.Context,
	o *GetDetectorOptions) (*ResponseMessageEnvelope, error) {
	if o == nil {
		o = &GetDetectorOptions{}
	}

	azResponse, err := c.sendPostRequest(
		ctx,
		o,
		nil)
	if err != nil {
		return nil, err
	}

	defer azResponse.Body.Close()

	detectorJson, err := io.ReadAll(azResponse.Body)

	if err != nil {
		return nil, err
	}

	return newResponseMessageEnvelope(o.ResourceID, o.DetectorID, o.Location, detectorJson)
}

func (c *Client) sendPostRequest(
	ctx context.Context,
	requestOptions appLensRequestOptions,
	requestEnricher func(*policy.Request)) (*http.Response, error) {
	req, err := c.createRequest(ctx, http.MethodPost, requestOptions, requestEnricher)
	if err != nil {
		return nil, err
	}

	return c.executeAndEnsureSuccessResponse(req)
}

func (c *Client) createRequest(
	ctx context.Context,
	method string,
	requestOptions appLensRequestOptions,
	requestEnricher func(*policy.Request)) (*policy.Request, error) {
	if requestOptions != nil {
		header := requestOptions.toHeader()
		ctx = runtime.WithHTTPHeader(ctx, header)
	}

	req, err := runtime.NewRequest(ctx, method, c.endpoint)
	if err != nil {
		return nil, err
	}

	if requestEnricher != nil {
		requestEnricher(req)
	}

	return req, nil
}

func (c *Client) executeAndEnsureSuccessResponse(request *policy.Request) (*http.Response, error) {
	response, err := c.pipeline.Do(request)
	if err != nil {
		return nil, err
	}

	successResponse := (response.StatusCode >= 200 && response.StatusCode < 300) || response.StatusCode == 304
	if successResponse {
		return response, nil
	}

	return nil, newAppLensError(response)
}

func newResponseMessageCollectionEnvelope(valueJson []byte, resourceID, location string) (*ResponseMessageCollectionEnvelope, error) {
	var results []interface{}
	err := json.Unmarshal(valueJson, &results)

	if err != nil {
		return nil, err
	}

	listResult := ResponseMessageCollectionEnvelope{}
	for _, v := range results {
		if id := getDetectorID(v); len(id) > 0 {
			detector := ResponseMessageEnvelope{
				Id:         path.Join(resourceID, "detectors", id),
				Name:       id,
				Location:   location,
				Type:       "Microsoft.RedHatOpenShift/openShiftClusters/detectors",
				Properties: v,
			}
			listResult.Value = append(listResult.Value, detector)
		}
	}

	return &listResult, nil
}

func getDetectorID(detector interface{}) string {
	if propertyMap, ok := detector.(map[string]interface{}); ok {
		if metadataMap, ok := propertyMap["metadata"].(map[string]interface{}); ok {
			if idObj, ok := metadataMap["id"]; ok {
				if id, ok := idObj.(string); ok {
					return id
				}
			}
		}
	}
	return ""
}

func newResponseMessageEnvelope(resourceID, name, location string, propertiesJson []byte) (*ResponseMessageEnvelope, error) {
	var converted interface{}
	err := json.Unmarshal(propertiesJson, &converted)

	if err != nil {
		return nil, err
	}

	return &ResponseMessageEnvelope{
		Id:         path.Join(resourceID, "detectors", name),
		Name:       name,
		Type:       "Microsoft.RedHatOpenShift/openShiftClusters/detectors",
		Location:   location,
		Properties: converted,
	}, nil
}
