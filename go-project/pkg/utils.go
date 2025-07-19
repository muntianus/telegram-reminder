package pkg

// Utility function to check if a string is empty
func IsEmpty(s string) bool {
    return len(s) == 0
}

// Utility function to concatenate two strings
func Concat(a, b string) string {
    return a + b
}

// Utility function to generate a greeting message
func Greet(name string) string {
    if IsEmpty(name) {
        return "Hello, Guest!"
    }
    return "Hello, " + name + "!"
}