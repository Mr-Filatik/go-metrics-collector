package common

// Константы заголовков запроса и ответа для клиентов и серверов.

const (
	// Типы данных.

	HeaderAccept                          = "Accept"           // поддерживаемый тип данных ответа
	HeaderContentType                     = "Content-Type"     // тип данных запроса
	HeaderContentTypeValueApplicationJSON = "application/json" // тип данных application/json
	HeaderContentTypeValueTextHTML        = "text/html"        // тип данных text/html

	// Форматы сжатия.

	HeaderAcceptEncoding    = "Accept-Encoding"  // поддерживаемый формат сжатия ответа
	HeaderContentEncoding   = "Content-Encoding" // формат сжатия запроса
	HeaderEncodingValueGZIP = "gzip"             // формат сжатия gzip

	// Другое.

	HeaderHashSHA256 = "HashSHA256"   // хэш-сумма контента запроса
	HeaderXRealIP    = "X-Real-IP"    // IP сети клиента
	HeaderXRequestID = "X-Request-Id" // ID запроса
)
