package api

import (
	"github.com/asaskevich/govalidator/v12"
	"github.com/google/uuid"
)

func init() {
	// UUID уже разобран JSON-декодером. Проверяем нулевое значение массива,
	// поскольку стандартный валидатор uuid предназначен для строк.
	govalidator.CustomTypeTagMap.Set("nonzeroUUID", func(value any, _ any) bool {
		id, ok := value.(uuid.UUID)
		return ok && id != uuid.Nil
	})
}
