package web

import (
	"errors"
	"log/slog"
	"net/http"

	"bookshop/frontend/internal/adapters/httpcore"
)

type errorDescription struct {
	Field   string
	Message string
}

var errorDescriptions = map[string]errorDescription{
	"invalid_input":                     {Message: "Проверьте заполнение полей."},
	"conflict":                          {Message: "Не удалось сохранить изменения: запись уже существует или используется."},
	"not_found":                         {Message: "Запись не найдена. Возможно, она была удалена."},
	"internal_error":                    {Message: "Не удалось выполнить запрос к серверу. Попробуйте позже."},
	"invalid_response":                  {Message: "Не удалось выполнить запрос к серверу. Попробуйте позже."},
	"book_author_required":              {Field: "author", Message: "Укажите автора."},
	"book_name_required":                {Field: "name", Message: "Укажите название книги."},
	"book_year_invalid":                 {Field: "year", Message: "Укажите год выхода от 1 до 9999."},
	"book_price_negative":               {Field: "price", Message: "Цена не может быть отрицательной."},
	"book_publisher_required":           {Field: "publisherId", Message: "Выберите издательство."},
	"book_publication_year_invalid":     {Field: "publicationYear", Message: "Укажите год выпуска версии от 1 до 9999."},
	"book_publication_year_before_year": {Field: "publicationYear", Message: "Год выпуска версии не может быть раньше года выхода."},
	"book_quantity_negative":            {Field: "quantity", Message: "Остаток не может быть отрицательным."},
	"publisher_name_required":           {Field: "name", Message: "Укажите название издательства."},
	"publisher_name_duplicate":          {Field: "name", Message: "Издательство с таким названием уже существует."},
	"book_duplicate":                    {Field: "name", Message: "Такая книга уже существует."},
	"book_publisher_not_found":          {Field: "publisherId", Message: "Выберите существующее издательство."},
	"publisher_referenced":              {Message: "Нельзя удалить издательство, пока на него ссылаются книги."},
	"invalid_request":                   {Message: "Некорректный формат запроса."},
	"invalid_json":                      {Message: "Проверьте формат JSON."},
	"trailing_json_data":                {Message: "Удалите лишние данные после JSON."},
	"invalid_id":                        {Message: "Некорректный идентификатор записи."},
	"invalid_author":                    {Field: "author", Message: "Укажите автора текстом."},
	"invalid_name":                      {Field: "name", Message: "Укажите название текстом."},
	"invalid_year":                      {Field: "year", Message: "Укажите год выхода целым числом от 1 до 9999."},
	"invalid_price":                     {Field: "price", Message: "Укажите цену целым числом в рублях. Дробные и слишком большие значения не поддерживаются."},
	"invalid_publisher_id":              {Field: "publisherId", Message: "Выберите издательство из списка."},
	"invalid_publication_year":          {Field: "publicationYear", Message: "Укажите год выпуска версии целым числом от 1 до 9999."},
	"invalid_quantity":                  {Field: "quantity", Message: "Укажите остаток целым числом. Дробные и слишком большие значения не поддерживаются."},
}

func translateErrorCode(code string) (string, map[string]string) {
	description, ok := errorDescriptions[code]
	if !ok {
		description = errorDescriptions["invalid_request"]
	}
	if description.Field != "" {
		return errorDescriptions["invalid_input"].Message, map[string]string{description.Field: description.Message}
	}
	return description.Message, nil
}

func classifyPublicError(err error) (int, string, map[string]string) {
	var formError *formValidationError
	if errors.As(err, &formError) {
		message, fields := translateErrorCode(formError.Code)
		return http.StatusBadRequest, message, fields
	}
	if errors.Is(err, errInvalidForm) {
		return http.StatusBadRequest, errorDescriptions["invalid_request"].Message, nil
	}
	var apiError *httpcore.APIError
	if errors.As(err, &apiError) && apiError.Code != "invalid_response" {
		switch apiError.StatusCode {
		case http.StatusBadRequest, http.StatusConflict:
			message, fields := translateErrorCode(apiError.Code)
			return apiError.StatusCode, message, fields
		case http.StatusNotFound:
			return http.StatusNotFound, errorDescriptions["not_found"].Message, nil
		}
	}
	slog.Error("web request failed", "error", err)
	return http.StatusBadGateway, errorDescriptions["internal_error"].Message, nil
}

func renderFormError(w http.ResponseWriter, err error, data page) {
	code, message, fields := classifyPublicError(err)
	data.Error = message
	data.Fields = fields
	renderPage(w, code, data)
}
