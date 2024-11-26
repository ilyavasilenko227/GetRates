package optel

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// SetUpOTelSDK инициализирует OpenTelemetry SDK и возвращает функцию завершения работы (shutdown).
// Эта функция настраивает провайдер трассировки, экспортирует трассы через OTLP HTTP и регистрирует провайдер.
func SetUpOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	// Список функций, которые необходимо вызвать для корректного завершения работы (например, закрытие провайдера).
	var shutdownFuncs []func(context.Context) error

	// Функция завершения работы, вызывающая все зарегистрированные shutdown-функции.
	shutdown = func(ctx context.Context) error {
		var err error

		// Выполнение всех зарегистрированных shutdown-функций в порядке добавления.
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		// Обнуление списка после выполнения.
		shutdownFuncs = nil
		return err
	}

	// Локальная функция для обработки ошибок.
	// Если возникает ошибка, она вызывает shutdown и сохраняет ошибки.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}
	// Создание провайдера трассировки.
	tracerProvider, err := newTraceProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	// Добавляем функцию завершения работы провайдера в список shutdown-функций.
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)

	// Устанавливаем новый провайдер трассировки как текущий в глобальном реестре.
	otel.SetTracerProvider(tracerProvider)
	return
}

// newTraceProvider создает новый провайдер трассировки с OTLP HTTP экспортером.
func newTraceProvider(ctx context.Context) (*trace.TracerProvider, error) {
	// Создание экспортера трассировки, который отправляет данные через OTLP по HTTP.
	traceExporter, err := otlptracehttp.New(ctx) // Возврат ошибки, если экспортер не удалось создать.

	if err != nil {
		return nil, err
	}

	// Создание провайдера трассировки с указанными параметрами:
	// - Использование батчевого экспорта (Batcher) для оптимизации отправки данных.
	// - Добавление ресурса с атрибутами (имени сервиса).
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Устанавливаем время ожидания для отправки батчей.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(resource.NewWithAttributes(
			// Используем стандартную схему OpenTelemetry для атрибутов.
			semconv.SchemaURL,
			// Указываем имя сервиса как атрибут ресурса.
			semconv.ServiceNameKey.String("GetRates"),
		)),
	)
	return traceProvider, nil
}
