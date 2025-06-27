package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	digits   = "0123456789"
	letters  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specials = "~%^&*_-@!+#$"
)

func main() {
	// Объявление флагов
	var (
		length      = flag.IntP("length", "L", 18, "Длина пароля")
		useDigits   = flag.BoolP("digits", "d", true, "Использовать цифры")
		useLetters  = flag.BoolP("letters", "l", true, "Использовать буквы")
		useSpecials = flag.BoolP("specials", "s", true, "Использовать спецсимволы")
		count       = flag.IntP("count", "c", 1, "Количество паролей")
		exclude     = flag.StringP("exclude", "e", "", "Символы для исключения")
		logLevel    = flag.StringP("log-level", "", "error", "Уровень логирования (trace, debug, info, warn, error)")
	)

	flag.Parse()

	// Настройка логгера Logrus
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger.SetOutput(os.Stderr)

	// Установка уровня логирования
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logger.Errorf("Некорректный уровень логирования: %s", *logLevel)
		os.Exit(1)
	}
	logger.SetLevel(level)

	logger.Info("Запуск генератора паролей")
	logger.Debugf("Уровень логирования: %s", level)

	// Валидация параметров
	if *length <= 0 {
		logger.Error("Длина пароля должна быть положительным числом")
		os.Exit(1)
	}
	if *count <= 0 {
		logger.Error("Количество паролей должно быть положительным числом")
		os.Exit(1)
	}
	if !*useDigits && !*useLetters && !*useSpecials {
		logger.Error("Должен быть выбран хотя бы один тип символов")
		os.Exit(1)
	}

	// Создание набора символов
	charSet := buildCharSet(*useDigits, *useLetters, *useSpecials, *exclude, logger)
	if len(charSet) == 0 {
		logger.Error("Набор символов пуст после исключений")
		os.Exit(1)
	}

	// Генерация паролей
	passwords := make([]string, *count)
	for i := 0; i < *count; i++ {
		pass, err := generatePassword(*length, charSet, logger)
		if err != nil {
			logger.Errorf("Ошибка генерации: %v", err)
			continue
		}
		passwords[i] = pass
	}

	// Вывод результатов
	for _, p := range passwords {
		if p != "" {
			fmt.Println(p)
		}
	}
	logger.Info("Генерация завершена")
}

func buildCharSet(useDigits, useLetters, useSpecials bool, exclude string, logger *logrus.Logger) string {
	var builder strings.Builder

	if useDigits {
		builder.WriteString(digits)
		logger.Debug("Добавлены цифры")
	}
	if useLetters {
		builder.WriteString(letters)
		logger.Debug("Добавлены буквы")
	}
	if useSpecials {
		builder.WriteString(specials)
		logger.Debug("Добавлены спецсимволы")
	}

	charSet := builder.String()
	if charSet == "" {
		return ""
	}

	// Фильтрация исключенных символов
	if exclude != "" {
		originalLen := len(charSet)
		filtered := strings.Map(func(r rune) rune {
			if !strings.ContainsRune(exclude, r) {
				return r
			}
			return -1
		}, charSet)

		if len(filtered) == 0 {
			logger.Warn("Все символы исключены")
		} else if len(filtered) < originalLen {
			logger.Debugf("Исключено символов: %d", originalLen-len(filtered))
		}
		charSet = filtered
	}

	return charSet
}

func generatePassword(length int, charSet string, logger *logrus.Logger) (string, error) {
	if len(charSet) == 0 {
		return "", fmt.Errorf("пустой набор символов")
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	var password strings.Builder

	logger.Debugf("Генерация пароля длиной %d символов", length)
	for i := 0; i < length; i++ {
		idx := rand.Intn(len(charSet))
		password.WriteByte(charSet[idx])
	}

	// Проверка что пароль содержит хотя бы один символ каждого типа
	result := password.String()
	if !isPasswordValid(result) {
		logger.Debug("Пароль не соответствует требованиям, перегенерация")
		return generatePassword(length, charSet, logger) // Рекурсивная перегенерация
	}

	return result, nil
}

func isPasswordValid(password string) bool {
	hasDigit := false
	hasLetter := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsLetter(r):
			hasLetter = true
		case strings.ContainsRune(specials, r):
			hasSpecial = true
		}
	}

	return (hasDigit || !strings.ContainsAny(password, digits)) &&
		(hasLetter || !strings.ContainsAny(password, letters)) &&
		(hasSpecial || !strings.ContainsAny(password, specials))
}
