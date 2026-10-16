package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"bookshop/frontend/internal/adapters/httpcore"
	"bookshop/frontend/internal/logcontext"

	"github.com/stretchr/testify/require"
)

func TestTranslateFieldErrorCodes(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		code, field, message string
	}{
		{"book_author_required", "author", "Укажите автора."},
		{"book_name_required", "name", "Укажите название книги."},
		{"book_year_invalid", "year", "Укажите год выхода от 1 до 9999."},
		{"book_price_negative", "price", "Цена не может быть отрицательной."},
		{"book_publisher_required", "publisherId", "Выберите издательство."},
		{"book_publication_year_invalid", "publicationYear", "Укажите год выпуска версии от 1 до 9999."},
		{"book_publication_year_before_year", "publicationYear", "Год выпуска версии не может быть раньше года выхода."},
		{"book_quantity_negative", "quantity", "Остаток не может быть отрицательным."},
		{"publisher_name_required", "name", "Укажите название издательства."},
		{"publisher_name_duplicate", "name", "Издательство с таким названием уже существует."},
		{"book_duplicate", "name", "Такая книга уже существует."},
		{"book_publisher_not_found", "publisherId", "Выберите существующее издательство."},
		{"invalid_author", "author", "Укажите автора текстом."},
		{"invalid_name", "name", "Укажите название текстом."},
		{"invalid_year", "year", "Укажите год выхода целым числом от 1 до 9999."},
		{"invalid_price", "price", "Укажите цену целым числом в рублях. Дробные и слишком большие значения не поддерживаются."},
		{"invalid_publisher_id", "publisherId", "Выберите издательство из списка."},
		{"invalid_publication_year", "publicationYear", "Укажите год выпуска версии целым числом от 1 до 9999."},
		{"invalid_quantity", "quantity", "Укажите остаток целым числом. Дробные и слишком большие значения не поддерживаются."},
	} {
		t.Run(tc.code, func(t *testing.T) {
			t.Parallel()
			message, fields := translateErrorCode(tc.code)
			require.Equal(t, "Проверьте заполнение полей.", message)
			require.Equal(t, map[string]string{tc.field: tc.message}, fields)
		})
	}
}

func TestTranslateGeneralErrorCodes(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		code, message string
	}{
		{"invalid_input", "Проверьте заполнение полей."},
		{"conflict", "Не удалось сохранить изменения: запись уже существует или используется."},
		{"not_found", "Запись не найдена. Возможно, она была удалена."},
		{"invalid_request", "Некорректный формат запроса."},
		{"invalid_json", "Проверьте формат JSON."},
		{"trailing_json_data", "Удалите лишние данные после JSON."},
		{"invalid_id", "Некорректный идентификатор записи."},
		{"publisher_referenced", "Нельзя удалить издательство, пока на него ссылаются книги."},
		{"new_unknown_code", "Некорректный формат запроса."},
	} {
		t.Run(tc.code, func(t *testing.T) {
			t.Parallel()
			message, fields := translateErrorCode(tc.code)
			require.Equal(t, tc.message, message)
			require.Empty(t, fields)
		})
	}
}

func TestClassifyPublicErrorUsesStatusAndLocalMessages(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		status, wantStatus int
		code               string
		message            string
	}{
		{400, 400, "invalid_input", "Проверьте заполнение полей."},
		{409, 409, "conflict", "Не удалось сохранить изменения: запись уже существует или используется."},
		{404, 404, "private SQL", "Запись не найдена. Возможно, она была удалена."},
		{500, 502, "internal_error", "Не удалось выполнить запрос к серверу. Попробуйте позже."},
		{400, 502, "invalid_response", "Не удалось выполнить запрос к серверу. Попробуйте позже."},
	} {
		t.Run(fmt.Sprintf("%d/%s", tc.status, tc.code), func(t *testing.T) {
			t.Parallel()
			err := fmt.Errorf("save: %w", &httpcore.APIError{StatusCode: tc.status, Code: tc.code})
			status, message, fields := classifyPublicError(err)
			require.Equal(t, tc.wantStatus, status)
			require.Equal(t, tc.message, message)
			require.Empty(t, fields)
		})
	}
}

func TestRenderUnknownBackendCodeWithoutExposingIt(t *testing.T) {
	t.Parallel()
	for _, hx := range []string{"", "true"} {
		t.Run("htmx="+hx, func(t *testing.T) {
			t.Parallel()
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"private SQL"}`))
			}))
			t.Cleanup(api.Close)
			req := httptest.NewRequest("POST", "/publishers", strings.NewReader(url.Values{"name": {"Издательство"}}.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("HX-Request", hx)
			res := httptest.NewRecorder()
			createTestHandler(api.URL).ServeHTTP(res, req)

			require.Equal(t, http.StatusBadRequest, res.Code)
			require.Contains(t, res.Body.String(), "Некорректный формат запроса.")
			require.NotContains(t, res.Body.String(), "private SQL")
			require.Contains(t, res.Body.String(), `value="Издательство"`)
		})
	}
}

func TestLogUnexpectedError(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		cause  error
		logged bool
	}{
		{name: "upstream failure", status: http.StatusBadGateway, cause: fmt.Errorf("list books: %w", &httpcore.APIError{StatusCode: http.StatusInternalServerError, Code: "internal_error"}), logged: true},
		{name: "bad request", status: http.StatusBadRequest, cause: &httpcore.APIError{StatusCode: http.StatusBadRequest, Code: "invalid_input"}},
		{name: "not found", status: http.StatusNotFound, cause: &httpcore.APIError{StatusCode: http.StatusNotFound, Code: "not_found"}},
		{name: "conflict", status: http.StatusConflict, cause: &httpcore.APIError{StatusCode: http.StatusConflict, Code: "conflict"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			server := Server{Logger: slog.New(logcontext.NewHandler(slog.NewJSONHandler(&output, nil)))}
			ctx := logcontext.StoreRequest(context.Background(), testID, http.MethodGet, "GET /books")
			server.logUnexpectedError(ctx, tc.status, tc.cause)
			if !tc.logged {
				require.Empty(t, output.String())
				return
			}
			var record map[string]any
			require.NoError(t, json.Unmarshal(output.Bytes(), &record))
			require.Equal(t, "http.request_failed", record["event"])
			require.Equal(t, "ERROR", record["level"])
			require.Equal(t, testID, record["request_id"])
			require.Equal(t, "GET", record["method"])
			require.Equal(t, "GET /books", record["route"])
			require.Equal(t, float64(tc.status), record["status_code"])
			require.Equal(t, float64(http.StatusInternalServerError), record["upstream_status_code"])
			require.Equal(t, "internal_error", record["error_code"])
			require.Contains(t, record["error"], "list books")
		})
	}
}
