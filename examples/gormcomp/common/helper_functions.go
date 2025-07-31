package common

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ConvertStringToInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func MakeLinkAvatar(domain string, avatar string) string {
	if strings.Contains(avatar, domain) {
		return avatar
	}
	return domain + "/" + avatar
}

func MakeLinkShareProfile(domain string, name string) string {
	if name == "" {
		return ""
	}
	return domain + "/" + name
}

func FormatPhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) > 0 && phone[0] == '+' {
		phone = phone[1:]
	}
	if len(phone) > 0 && phone[0] == '0' {
		phone = "84" + phone[1:]
	}
	return phone
}

// Paging represents pagination parameters for API requests.
type Paging struct {
	Offset int `json:"offset" form:"offset"`
	Limit  int `json:"limit" form:"limit"`
}

func (p *Paging) Process() {
	if p.Offset < 0 {
		p.Offset = 0
	}

	if p.Limit <= 0 {
		p.Limit = 10
	}

	if p.Limit >= 200 {
		p.Limit = 200
	}
}

// ConvertVietnameseToSlug converts a Vietnamese string to a URL-friendly slug.
func ConvertVietnameseToSlug(input string) string {
	// map for Vietnamese characters to their replacements
	vietnameseMap := map[rune]string{
		'à': "a", 'á': "a", 'ạ': "a", 'ả': "a", 'ã': "a",
		'â': "a", 'ầ': "a", 'ấ': "a", 'ậ': "a", 'ẩ': "a", 'ẫ': "a",
		'ă': "a", 'ằ': "a", 'ắ': "a", 'ặ': "a", 'ẳ': "a", 'ẵ': "a",
		'è': "e", 'é': "e", 'ẹ': "e", 'ẻ': "e", 'ẽ': "e",
		'ê': "e", 'ề': "e", 'ế': "e", 'ệ': "e", 'ể': "e", 'ễ': "e",
		'ì': "i", 'í': "i", 'ị': "i", 'ỉ': "i", 'ĩ': "i",
		'ò': "o", 'ó': "o", 'ọ': "o", 'ỏ': "o", 'õ': "o",
		'ô': "o", 'ồ': "o", 'ố': "o", 'ộ': "o", 'ổ': "o", 'ỗ': "o",
		'ơ': "o", 'ờ': "o", 'ớ': "o", 'ợ': "o", 'ở': "o", 'ỡ': "o",
		'ù': "u", 'ú': "u", 'ụ': "u", 'ủ': "u", 'ũ': "u",
		'ư': "u", 'ừ': "u", 'ứ': "u", 'ự': "u", 'ử': "u", 'ữ': "u",
		'ỳ': "y", 'ý': "y", 'ỵ': "y", 'ỷ': "y", 'ỹ': "y",
		'đ': "d",
		'À': "A", 'Á': "A", 'Ạ': "A", 'Ả': "A", 'Ã': "A",
		'Â': "A", 'Ầ': "A", 'Ấ': "A", 'Ậ': "A", 'Ẩ': "A", 'Ẫ': "A",
		'Ă': "A", 'Ằ': "A", 'Ắ': "A", 'Ặ': "A", 'Ẳ': "A", 'Ẵ': "A",
		'È': "E", 'É': "E", 'Ẹ': "E", 'Ẻ': "E", 'Ẽ': "E",
		'Ê': "E", 'Ề': "E", 'Ế': "E", 'Ệ': "E", 'Ể': "E", 'Ễ': "E",
		'Ì': "I", 'Í': "I", 'Ị': "I", 'Ỉ': "I", 'Ĩ': "I",
		'Ò': "O", 'Ó': "O", 'Ọ': "O", 'Ỏ': "O", 'Õ': "O",
		'Ô': "O", 'Ồ': "O", 'Ố': "O", 'Ộ': "O", 'Ổ': "O", 'Ỗ': "O",
		'Ơ': "O", 'Ờ': "O", 'Ớ': "O", 'Ợ': "O", 'Ở': "O", 'Ỡ': "O",
		'Ù': "U", 'Ú': "U", 'Ụ': "U", 'Ủ': "U", 'Ũ': "U",
		'Ư': "U", 'Ừ': "U", 'Ứ': "U", 'Ự': "U", 'Ử': "U", 'Ữ': "U",
		'Ỳ': "Y", 'Ý': "Y", 'Ỵ': "Y", 'Ỷ': "Y", 'Ỹ': "Y",
		'Đ': "D",
	}

	var result strings.Builder
	for _, char := range input {
		if replacement, exists := vietnameseMap[char]; exists {
			result.WriteString(replacement)
		} else {
			result.WriteRune(char)
		}
	}

	// lowercase the result and replace spaces with hyphens
	output := strings.ToLower(result.String())
	output = strings.ReplaceAll(output, " ", "-")

	return output
}

// GetPathParamUUID retrieves a UUID from the path parameters of a gin context.
func GetPathParamUUID(c *gin.Context, paramName string) (uuid.UUID, error) {
	paramValue := c.Params.ByName(paramName)
	if paramValue == "" {
		return uuid.Nil, fmt.Errorf("missing path parameter: %s", paramName)
	}

	id, err := uuid.Parse(paramValue)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format for parameter %s: %v", paramName, err)
	}

	return id, nil
}

// RandomString generates a random string of the specified length.
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
