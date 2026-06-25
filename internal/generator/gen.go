package generator

import (
	"fmt"
	"time"

	"github.com/Vladroon22/TaskTracker/internal/models"
)

// GenerateRecurringDates генерирует все даты выпадений для повторяющейся задачи
func GenerateRecurringDates(task models.Task, from, to time.Time) []time.Time {
	return generateRecurringDates(task, from, to)
}

func generateRecurringDates(task models.Task, from, to time.Time) []time.Time {
	var dates []time.Time

	// Определяем начальную дату
	current := task.DueDate.Time()
	if from.After(current) {
		current = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	}

	// Защита:
	maxDates := 20

	for !current.After(to) && len(dates) < maxDates {
		var shouldAdd bool

		tp := task.RecurrenceType
		switch *tp {
		case models.RecurrenceDaily:
			shouldAdd = shouldAddDaily(task, current)

		case models.RecurrenceMonthly:
			shouldAdd = shouldAddMonthly(task, current)

		case models.RecurrenceSpecific:
			shouldAdd = shouldAddSpecific(task, current)

		case models.RecurrenceEvenOdd:
			shouldAdd = shouldAddEvenOdd(task, current)
		}

		if shouldAdd {
			dates = append(dates, current)
		}

		current = current.AddDate(0, 0, 1)
	}

	return dates
}

// shouldAddDaily проверяет ежедневные задачи
func shouldAddDaily(task models.Task, date time.Time) bool {
	if *task.RecurrenceInterval <= 0 {
		*task.RecurrenceInterval = 1
	}

	// Вычисляем разницу в днях от начальной даты
	daysDiff := int(date.Sub(task.DueDate.Time()).Hours() / 24)

	// Проверяем, что прошло нужное количество интервалов
	interval := (daysDiff % *task.RecurrenceInterval)
	return interval == 0
}

// shouldAddMonthly проверяет ежемесячные задачи на определенные числа
func shouldAddMonthly(task models.Task, date time.Time) bool {
	if len(task.RecurrenceDays) == 0 {
		return false
	}

	currentDay := date.Day()

	// Получаем последний день месяца
	lastDay := daysInMonth(date.Year(), int(date.Month()))

	for _, day := range task.RecurrenceDays {
		// Для дней, которые могут не существовать в месяце (31 февраля)
		checkDay := day
		if checkDay > lastDay {
			checkDay = lastDay
		}

		if currentDay == checkDay {
			return true
		}
	}

	return false
}

// shouldAddSpecific проверяет конкретные даты
func shouldAddSpecific(task models.Task, date time.Time) bool {
	if len(task.RecurrenceDays) == 0 {
		return false
	}

	currentDay := date.Day()

	for _, day := range task.RecurrenceDays {

		// Проверяем только число месяца
		if currentDay == day {
			// Проверяем, что дата входит в диапазон
			if !date.Before(task.DueDate.Time()) {
				return true
			}
		}
	}

	return false
}

// shouldAddEvenOdd проверяет четные/нечетные дни
func shouldAddEvenOdd(task models.Task, date time.Time) bool {
	currentDay := date.Day()
	isEven := currentDay%2 == 0

	// recurrence_interval: 0 = четные, 1 = нечетные
	if *(task.RecurrenceInterval) == 0 {
		return isEven
	}
	return !isEven
}

// daysInMonth возвращает количество дней в месяце
func daysInMonth(year int, month int) int {
	if month == 2 {
		if isLeapYear(year) {
			return 29
		}
		return 28
	}

	if month == 4 || month == 6 || month == 9 || month == 11 {
		return 30
	}

	return 31
}

// isLeapYear проверяет високосный год
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// ValidateRecurrenceRule валидирует правило повторения
func ValidateRecurrenceRule(recurrenceType string, interval int, days []int) error {
	switch recurrenceType {
	case models.RecurrenceDaily:
		if interval < 1 {
			return fmt.Errorf("daily interval must be >= 1")
		}

	case models.RecurrenceMonthly:
		if len(days) == 0 {
			return fmt.Errorf("monthly recurrence requires days of month")
		}
		for _, day := range days {
			if day < 1 || day > 31 {
				return fmt.Errorf("invalid day of month: %d", day)
			}
		}

	case models.RecurrenceSpecific:
		if len(days) == 0 {
			return fmt.Errorf("specific recurrence requires days")
		}
		for _, day := range days {
			if day < 1 || day > 31 {
				return fmt.Errorf("invalid day: %d", day)
			}
		}

	case models.RecurrenceEvenOdd:
		if interval != 0 && interval != 1 {
			return fmt.Errorf("even_odd interval must be 0 (even) or 1 (odd)")
		}

	default:
		return fmt.Errorf("unknown recurrence type: %s", recurrenceType)
	}

	return nil
}
