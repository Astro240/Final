package api
import (
	"bufio"
	"os"
	"strings"
)

func LoadEnv() {
    file, err := os.Open(".env")
    if err != nil {
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        
        if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
            continue
        }
        
        if strings.Contains(line, "=") {
            parts := strings.SplitN(line, "=", 2)
            key := strings.TrimSpace(parts[0])
            value := strings.TrimSpace(parts[1])
            
            value = strings.Trim(value, `"'`)
            
            os.Setenv(key, value)
        }
    }
}