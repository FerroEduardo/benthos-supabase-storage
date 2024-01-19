package output

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/benthosdev/benthos/v4/public/service"
)

func init() {
	err := service.RegisterOutput(
		"supabase_storage", service.NewConfigSpec().
			Field(service.NewStringField("bucket")).
			Field(service.NewStringField("baseUrl")).
			Field(service.NewStringField("token")),
		func(conf *service.ParsedConfig, mgr *service.Resources) (out service.Output, maxInFlight int, err error) {
			output, err := parseConfig(conf)
			if err != nil {
				return nil, 0, err
			}

			return output, 1, nil
		})
	if err != nil {
		panic(err)
	}
}

//------------------------------------------------------------------------------

type supabaseStorageOutput struct {
	bucket  string
	baseUrl string
	token   string
}

func parseConfig(conf *service.ParsedConfig) (*supabaseStorageOutput, error) {
	bucket, err := conf.FieldString("bucket")
	if err != nil {
		return nil, err
	}

	baseUrl, err := conf.FieldString("baseUrl")
	if err != nil {
		return nil, err
	}

	token, err := conf.FieldString("token")
	if err != nil {
		return nil, err
	}

	return &supabaseStorageOutput{
		bucket:  bucket,
		baseUrl: baseUrl,
		token:   token,
	}, nil
}

func (b *supabaseStorageOutput) Connect(ctx context.Context) error {
	return nil
}

type RequestData struct {
	data        []byte
	filename    string
	contentType string
}

func (b *supabaseStorageOutput) Write(ctx context.Context, msg *service.Message) error {
	raw, err := msg.AsStructured()

	if err != nil {
		return err
	}
	rawMap := raw.(map[string]any)
	content := RequestData{
		data:        rawMap["data"].([]byte),
		filename:    rawMap["filename"].(string),
		contentType: rawMap["contentType"].(string),
	}

	err = b.upload(content.filename, content.data, content.contentType)
	if err != nil {
		return err
	}

	fmt.Printf("[supabase_storage]: %s\n", content)
	return nil
}

func (b *supabaseStorageOutput) Close(ctx context.Context) error {
	return nil
}

func (b *supabaseStorageOutput) upload(wildcard string, file []byte, contentType string) error {
	url := fmt.Sprintf("%s/object/%s/%s", b.baseUrl, b.bucket, wildcard)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(file))
	if err != nil {
		fmt.Printf("Failed to create request: %s\n", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+b.token)
	req.Header.Set("Content-Type", contentType)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to make request: %s\n", err)
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Failed to read request body: %s\n", err)
		return err
	}

	fmt.Printf("client: status code: %d\n", res.StatusCode)
	if res.StatusCode != 200 {
		msg := fmt.Sprintf("Failed to save to storage - Status: %d - Response: %s", res.StatusCode, body)
		return errors.New(msg)
	}

	return nil
}
