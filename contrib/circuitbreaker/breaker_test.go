package circuitbreaker

import (
	"errors"
	"testing"
	"time"

	"github.com/quenbyako/core"
)

func succeed(cb core.CircuitBreaker) error {
	reporter, ok := cb.Allow()
	if !ok {
		return errors.New("cb rejected")
	}

	reporter.Success()
	return nil
}

func fail(cb core.CircuitBreaker) error {
	reporter, ok := cb.Allow()
	if !ok {
		return errors.New("cb rejected")
	}

	reporter.Fail()
	return nil
}

func TestCircuitBreaker(t *testing.T) {
	cbInterface, err := New(
		WithMaxRequests(2),
		WithTimeout(time.Hour),
	)
	assertNil(t, err)
	cb := cbInterface.(*CircuitBreaker)
	setTestName(cb, "cb")

	assertEqual(t, cb.Name(), "cb")

	// 1. Тестируем накопление ошибок в Closed
	for i := 0; i < 5; i++ {
		assertNil(t, fail(cb))
	}

	assertEqual(t, cb.State(), core.StateClosed)
	assertEqual(t, cb.Counts(), Counts{Requests: 5, TotalFailures: 5, ConsecutiveFailures: 5})

	// 2. Сброс CF при успехе
	assertNil(t, succeed(cb))
	assertEqual(t, cb.State(), core.StateClosed)
	assertEqual(t, cb.Counts(), Counts{Requests: 6, TotalSuccesses: 1, TotalFailures: 5, ConsecutiveSuccesses: 1})

	// 3. Переход в Open после 6 ошибок подряд
	assertNil(t, fail(cb)) // CF = 1
	for i := 0; i < 5; i++ {
		assertNil(t, fail(cb)) // CF = 6
	}

	assertEqual(t, cb.State(), core.StateOpen)
	assertEqual(t, cb.Counts(), Counts{})
	assertFalse(t, cb.expiry.IsZero())

	// 4. Проверка блокировки в Open
	assertError(t, succeed(cb))
	assertError(t, fail(cb))

	// 5. Переход в HalfOpen по времени
	pseudoSleep(cb, time.Hour + time.Second)
	
	// Первый запрос в HalfOpen (триггерит обновление состояния)
	rep1, ok1 := cb.Allow()
	assertTrue(t, ok1)
	assertEqual(t, cb.State(), core.StateHalfOpen)
	assertTrue(t, cb.expiry.IsZero())

	// Второй запрос
	rep2, ok2 := cb.Allow()
	assertTrue(t, ok2)

	// Третий запрос должен быть отклонен (MaxRequests=2)
	assertError(t, succeed(cb))

	// 6. Переход в Open при ошибке в HalfOpen
	rep1.Fail()
	rep2.Success() // Неважно, так как rep1.Fail перевел в Open
	assertEqual(t, cb.State(), core.StateOpen)

	// 7. Возврат в Closed после успехов в HalfOpen
	pseudoSleep(cb, time.Hour + time.Second)
	
	// Выполняем 2 успешных запроса
	assertNil(t, succeed(cb))
	assertEqual(t, cb.State(), core.StateHalfOpen)
	assertNil(t, succeed(cb))
	
	// Теперь должно быть Closed (так как MaxRequests=2 достигнут и всё ок)
	assertEqual(t, cb.State(), core.StateClosed)
}
