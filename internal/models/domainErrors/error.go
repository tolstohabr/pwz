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
	//todo: новые ошибки
	ErrWeightTooHeavy = errors.New("вес слишком большой")
	ErrInvalidPackage = errors.New("неизвестная упаковка или другая ошибка упаковки") //можно просто VALIDATION_FAILED
	//todo: ошибки gRPC и HTTP
	//
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

	//todo: новые ошибки
	ErrWeightTooHeavy: "WEIGHT_TOO_HEAVY",
	ErrInvalidPackage: "INVALID_PACKAGE", //можно просто VALIDATION_FAILED
}
