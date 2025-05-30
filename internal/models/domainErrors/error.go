package domainErrors

import "errors"

var (
	ErrValidationFailed   = errors.New("ошибка проверки") //общее	//значения не прошли проверку на соответствие заданным правилам
	ErrOrderNotFound      = errors.New("заказ не найден")
	ErrInvalidInput       = errors.New("неверный ввод")
	ErrDuplicateOrder     = errors.New("заказ с таким ID уже существует")
	ErrOrderAlreadyExists = errors.New("заказ уже есть")
	ErrOrderAlreadyIssued = errors.New("заказ у клиента")
	ErrStorageExpired     = errors.New("время хранения истекло")
	ErrStorageNotExpired  = errors.New("время хранения не истекло")
	ErrReturnTimeExpired  = errors.New("время возврата истекло")
	ErrInvalidAction      = errors.New("непредусмотренное действие")
	ErrInternalError      = errors.New("внутренняя ошибка") //общее
	ErrImportFailed       = errors.New("ошибка импорта")    //общее
	ErrOpenFiled          = errors.New("ошибка открытия")   //общее
	ErrReadFiled          = errors.New("ошибка записи")     //общее
	ErrJsonFiled          = errors.New("ошибка JSON")       //общее
)

// Привязка ошибок к кодам
var ErrorCodes = map[error]string{
	ErrValidationFailed: "VALIDATION_FAILED",
	ErrOrderNotFound:    "ORDER_NOT_FOUND",
	ErrInvalidInput:     "INVALID_INPUT",

	ErrDuplicateOrder:     "DUPLICATE_ORDER",
	ErrOrderAlreadyExists: "ORDER_ALREADY_EXISTS",
	ErrOrderAlreadyIssued: "ORDER_ALREADY_ISSUED",

	ErrStorageExpired:    "STORAGE_EXPIRED",
	ErrStorageNotExpired: "STORAGE_NOT_EXPIRED",
	ErrReturnTimeExpired: "RETURN_TIME_EXPIRED",

	ErrInvalidAction: "INVALID_ACTION",
	ErrInternalError: "INTERNAL_ERROR",
	ErrImportFailed:  "IMPORT_FAILED",
	ErrOpenFiled:     "OPEN_FAILED",
	ErrReadFiled:     "READ_FILE_ERROR",
	ErrJsonFiled:     "JSON_FAILED",
}

/*
	ValidationFailedExpiresAtError = errors.New("ERROR: VALIDATION_FAILED: срок хранения в прошлом")
	OrderAlreadyExistsError        = errors.New("ERROR: ORDER_ALREADY_EXISTS: заказ уже есть")
	OrderAlreadyIssuedError        = errors.New("ERROR: ORDER_ALREADY_ISSUED: заказ у клиента")
	StorageNotExpiredError         = errors.New("ERROR: STORAGE_NOT_EXPIRED: время хранения не истекло")

	ErrOrderNotFound = errors.New("ERROR: ORDER_NOT_FOUND: заказ не найден")
	DuplicateOrder   = errors.New("ERROR: DUPLICATE_ORDER: заказ с таким ID уже существует")
*/

/*
Примеры допустимых кодов ошибок (общие):

INVALID_PACKAGE			недопустимый пакет
WEIGHT_TOO_HEAVY		слишком тяжёлый вес
ORDER_NOT_FOUND			заказ не найден
ORDER_ALREADY_EXISTS	заказ уже существует
STORAGE_EXPIRED			срок хранения истёк
VALIDATION_FAILED		ошибка проверки
INTERNAL_ERROR			внутренняя ошибка
*/

/*
Предложения от чата гпт

Работа с данными / ресурсами

DATA_NOT_FOUND	данные не найдены
DATA_DUPLICATE	запись уже существует
DATA_CONFLICT	конфликт данных (например, версия)
DATA_VALIDATION	данные не прошли валидацию


Системные / внутренняя логика

INTERNAL_ERROR	внутренняя ошибка
INIT_FAILED	ошибка при инициализации


Пользовательский ввод / CLI

INPUT_INVALID	недопустимый ввод
INPUT_MISSING	обязательный аргумент не указан
*/
