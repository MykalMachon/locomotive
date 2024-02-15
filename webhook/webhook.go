package webhook

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/graphql"
	"github.com/ferretcode/locomotive/logline"
)

var acceptedStatusCodes = []int{
	http.StatusOK,
	http.StatusNoContent,
	http.StatusAccepted,
	http.StatusCreated,
}

func SendGenericWebhook(log *graphql.EnvironmentLog, cfg *config.Config) (err error) {
	if len(log.MessageRaw) == 0 {
		return nil
	}

	jsonObject, err := logline.ReconstructLogLine(log)	

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.IngestUrl, bytes.NewBuffer(jsonObject))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-ndjson")

	for key, value := range cfg.AdditionalHeaders {
		req.Header.Set(key, value)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}

	defer res.Body.Close()

	if !slices.Contains(acceptedStatusCodes, res.StatusCode) {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("non success status code: %d", res.StatusCode)
		}

		return errors.New(string(body))
	}

	return nil
}
