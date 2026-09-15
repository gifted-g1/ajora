package utils

import "fmt"

// Kobo = smallest Naira unit. 1 Naira = 100 Kobo. NEVER use floats.
func FormatKobo(kobo int64) string {
    negative := kobo < 0
    if negative {
        kobo = -kobo
    }
    naira := kobo / 100
    remainder := kobo % 100
    nairaStr := fmt.Sprintf("%d", naira)
    result := ""
    for i, c := range nairaStr {
        if i > 0 && (len(nairaStr)-i)%3 == 0 {
            result += ","
        }
        result += string(c)
    }
    formatted := fmt.Sprintf("₦%s.%02d", result, remainder)
    if negative {
        formatted = "-" + formatted
    }
    return formatted
}

func ValidateAmount(kobo int64) error {
    if kobo <= 0 {
        return fmt.Errorf("amount must be positive")
    }
    if kobo > 100_000_000_00 {
        return fmt.Errorf("amount exceeds maximum")
    }
    return nil
}
