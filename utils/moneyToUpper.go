package utils

import (
	"strings"
)

// NumberToChinese 阿拉伯数字转中文大写金额
func NumberToChinese(amount float64) string {
	units := []string{"", "拾", "佰", "仟"}
	largeUnits := []string{"", "万", "亿", "兆"}
	digits := []string{"零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖"}

	integerPart := int64(amount)
	decimalPart := int64((amount - float64(integerPart)) * 100) // 保留小数点后两位

	var result strings.Builder

	// 转换整数部分
	if integerPart == 0 {
		result.WriteString("零")
	} else {
		sectionIndex := 0 // 当前段落计数
		for integerPart > 0 {
			section := integerPart % 10000
			if section > 0 {
				sectionResult := convertSection(int(section), digits, units)
				if sectionIndex > 0 {
					sectionResult += largeUnits[sectionIndex]
				}
				result.WriteString(sectionResult)
			}
			integerPart /= 10000
			sectionIndex++
		}
	}

	// 处理小数部分
	result.WriteString("圆")
	if decimalPart > 0 {
		jiao := decimalPart / 10
		fen := decimalPart % 10
		if jiao > 0 {
			result.WriteString(digits[jiao] + "角")
		}
		if fen > 0 {
			result.WriteString(digits[fen] + "分")
		}
	} else {
		result.WriteString("整")
	}

	return reverseString(result.String())
}

// convertSection 转换每个四位小节
func convertSection(section int, digits, units []string) string {
	var sectionResult strings.Builder
	zero := true // 是否需要写零
	for i := 0; section > 0; i++ {
		digit := section % 10
		if digit == 0 {
			if !zero {
				sectionResult.WriteString(digits[0])
				zero = true
			}
		} else {
			sectionResult.WriteString(units[i] + digits[digit])
			zero = false
		}
		section /= 10
	}
	return reverseString(sectionResult.String())
}

// reverseString 字符串反转
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
