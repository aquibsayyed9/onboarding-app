package utils

import (
    "fmt"
    "time"
    "strings"
)

// TemplateFunctions returns a map of functions available to templates
func TemplateFunctions() map[string]interface{} {
    return map[string]interface{}{
        // Existing functions
        "statusBadgeClass": GetStatusBadgeClass,
        "subtract":         Subtract,
        "add":             Add,
        
        // New formatting functions
        "formatNumber":     FormatNumber,
        "formatMoney":     FormatMoney,
        "formatDate":  FormatDate,
        "truncate":        Truncate,
    }
}

// GetStatusBadgeClass returns the CSS classes for status badges
func GetStatusBadgeClass(status string) string {
    switch status {
    case "pending":
        return "bg-yellow-100 text-yellow-800"
    case "approved":
        return "bg-green-100 text-green-800"
    case "rejected":
        return "bg-red-100 text-red-800"
    default:
        return "bg-gray-100 text-gray-800"
    }
}

// Subtract performs subtraction for template pagination
func Subtract(a, b int) int {
    return a - b
}

// Add performs addition for template pagination
func Add(a, b int) int {
    return a + b
}

// FormatNumber formats large numbers with thousand separators
func FormatNumber(n interface{}) string {
    var val float64
    switch v := n.(type) {
    case int:
        val = float64(v)
    case int64:
        val = float64(v)
    case float64:
        val = v
    default:
        return "0"
    }
    
    parts := strings.Split(fmt.Sprintf("%.0f", val), ".")
    intPart := parts[0]
    
    // Add thousand separators
    if len(intPart) > 3 {
        for i := len(intPart) - 3; i > 0; i -= 3 {
            intPart = intPart[:i] + "," + intPart[i:]
        }
    }
    
    return intPart
}

// FormatMoney formats monetary values with 2 decimal places and currency symbol
func FormatMoney(n interface{}) string {
    var val float64
    switch v := n.(type) {
    case int:
        val = float64(v)
    case int64:
        val = float64(v)
    case float64:
        val = v
    default:
        return "$0.00"
    }
    
    return fmt.Sprintf("$%.2f", val)
}

// FormatDateTime formats time.Time into a readable string
func FormatDate(t interface{}) string {
    switch v := t.(type) {
    case time.Time:
        return v.Format("Jan 02, 2006 15:04")
    case *time.Time:
        if v == nil {
            return ""
        }
        return v.Format("Jan 02, 2006 15:04")
    default:
        return ""
    }
}

// Truncate limits string length with ellipsis
func Truncate(s string, length int) string {
    if len(s) <= length {
        return s
    }
    return s[:length] + "..."
}