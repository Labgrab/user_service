package service_test

import (
	"labgrab/user_service/internal/service"
	"testing"
)

func TestValidateAlphabeticString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "cyrillic surname", input: "Иванов", want: true},
		{name: "cyrillic full name", input: "Иван Петров", want: true},
		{name: "cyrillic full name with patronymic", input: "Иван Петрович Сидоров", want: true},
		{name: "cyrillic hyphenated surname", input: "Петров-Водкин", want: true},
		{name: "cyrillic name with initials", input: "Иванов И.П.", want: true},
		{name: "cyrillic name with underscore", input: "Иван_Петрович", want: true},
		{name: "cyrillic and latin mixed name", input: "Ivanov Иван", want: true},
		{name: "ukrainian cyrillic name", input: "Шевченко Тарас", want: true},
		{name: "cyrillic uppercase name", input: "ИВАНОВ ИВАН", want: true},
		{name: "cyrillic lowercase name", input: "иванов иван", want: true},
		{name: "cyrillic with 'ё' letter", input: "Семён Алёшин", want: true},
		{name: "latin surname", input: "Smith", want: true},
		{name: "latin full name", input: "John Smith", want: true},
		{name: "latin hyphenated surname", input: "Smith-Johnson", want: true},

		{name: "empty string", input: "", want: false},
		{name: "cyrillic name with digits", input: "Иванов123", want: false},
		{name: "only digits", input: "12345", want: false},
		{name: "name with at symbol", input: "Иванов@mail", want: false},
		{name: "name with parentheses", input: "Иванов()", want: false},
		{name: "name with quotes", input: "Иванов\"", want: false},
		{name: "name with slash", input: "Иванов/Петров", want: false},
		{name: "name with asterisk", input: "Иванов*", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.ValidateAlphabeticString(tt.input)
			if got != tt.want {
				t.Errorf("ValidateAlphabeticString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateGroupCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "cyrillic 2 letters", input: "ИТ-1-1", want: true},
		{name: "cyrillic 3 letters", input: "МАТ-12-34", want: true},
		{name: "cyrillic 2 letters double digits", input: "ФИ-99-99", want: true},
		{name: "cyrillic uppercase", input: "ПМ-5-7", want: true},
		{name: "ukrainian cyrillic", input: "КИ-1-2", want: true},
		{name: "latin 2 letters", input: "IT-1-1", want: true},
		{name: "latin 3 letters", input: "MAT-12-34", want: true},

		{name: "empty string", input: "", want: false},
		{name: "one letter only", input: "И-1-1", want: false},
		{name: "four letters", input: "ИТИТ-1-1", want: false},
		{name: "no hyphens", input: "ИТ11", want: false},
		{name: "three digits first group", input: "ИТ-111-1", want: false},
		{name: "three digits second group", input: "ИТ-1-111", want: false},
		{name: "letters instead of digits", input: "ИТ-А-Б", want: false},
		{name: "wrong order", input: "1-ИТ-1", want: false},
		{name: "single hyphen only", input: "ИТ-11", want: false},
		{name: "extra hyphens", input: "ИТ-1-1-1", want: false},
		{name: "with spaces", input: "ИТ -1-1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.ValidateGroupCode(tt.input)
			if got != tt.want {
				t.Errorf("ValidateGroupCode(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "russia minimal", input: "+71234567890", want: true},
		{name: "russia full", input: "+79991234567", want: true},
		{name: "usa", input: "+12025551234", want: true},
		{name: "ukraine", input: "+380123456789", want: true},
		{name: "kazakhstan", input: "+77001234567", want: true},
		{name: "germany", input: "+4915112345678", want: true},
		{name: "china", input: "+8613812345678", want: true},
		{name: "minimal length", input: "+123", want: true},
		{name: "maximum length 15", input: "+123456789012345", want: true},

		{name: "empty string", input: "", want: false},
		{name: "no plus sign", input: "79991234567", want: false},
		{name: "starts with zero", input: "+09991234567", want: false},
		{name: "too short", input: "+1", want: false},
		{name: "too long", input: "+1234567890123456", want: false},
		{name: "with spaces", input: "+7 999 123 45 67", want: false},
		{name: "with hyphens", input: "+7-999-123-45-67", want: false},
		{name: "with parentheses", input: "+7(999)1234567", want: false},
		{name: "with letters", input: "+7999ABC4567", want: false},
		{name: "only plus", input: "+", want: false},
		{name: "plus in middle", input: "7+9991234567", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.ValidatePhoneNumber(tt.input)
			if got != tt.want {
				t.Errorf("ValidatePhoneNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateTelegramID(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  bool
	}{
		{name: "valid ID", input: 1234567890, want: true},

		{name: "less than 1", input: 0, want: false},
		{name: "negative number", input: -1, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.ValidateTelegramID(tt.input)
			if got != tt.want {
				t.Errorf("ValidateTelegramID(%d) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
