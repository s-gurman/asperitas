package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"asperitas/internal/errs"

	"go.uber.org/zap"
)

// Логирует выполненные действия и пишет http ответ
func WriteAndLogData(w http.ResponseWriter, v interface{}, logger *zap.SugaredLogger, logString string) {
	data, err := json.Marshal(v)
	if err != nil {
		logger.Errorf("encode data to json err: %s", err)
		return
	}
	logger.Infof(logString)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err = w.Write(data); err != nil {
		logger.Errorf("write resp err: %s", err)
	}
}

// Логирует входящие ошибки (в том числе и кастомные) и пишет http ответ
func WriteAndLogErr(w http.ResponseWriter, err error, logger *zap.SugaredLogger, logPrefix string) {
	var (
		detailErrs errs.DetailErrors
		msgErr     errs.MsgError
		resp       string
		code       int
	)
	switch {
	case errors.As(err, &detailErrs):
		code = detailErrs.Status
		resp = detailErrs.Error()
		logger.Infof("%s: code=%d errs=%#v", logPrefix, code, detailErrs.Errors)
	case errors.As(err, &msgErr):
		code = msgErr.Status
		resp = msgErr.Error()
		logger.Infof("%s: code=%d msg=%s", logPrefix, code, msgErr.Msg)
	default:
		code = http.StatusInternalServerError
		resp = `{"message":"internal server error"}`
		logger.Errorf("%s: code=%d msg=%s", logPrefix, code, err)
	}
	w.WriteHeader(code)
	if _, err = fmt.Fprint(w, resp); err != nil {
		logger.Errorf("write resp err: %s", err)
	}
}
