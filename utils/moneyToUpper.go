package utils

import (
	"strings"
)

// 数字对应的中文大写
var digitUpper = []string{"零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖"}
var unitUpper = []string{"", "拾", "佰", "仟"}
var sectionUpper = []string{"", "万", "亿", "兆"}

// 转换整数部分
func convertIntegerPart(num int64) string {
	if num == 0 {
		return "零元"
	}

	var result strings.Builder
	sectionIndex := 0
	zeroFlag := false // 标记连续零

	for num > 0 {
		section := num % 10000
		num /= 10000

		if section == 0 {
			if !zeroFlag && sectionIndex > 0 {
				result.WriteString(sectionUpper[sectionIndex])
			}
			zeroFlag = true
		} else {
			zeroFlag = false
			sectionStr := convertSection(int(section))
			if sectionIndex > 0 {
				sectionStr += sectionUpper[sectionIndex]
			}
			result.WriteString(sectionStr)
		}

		sectionIndex++
	}

	runes := []rune(result.String())
	// 反转字符串
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes) + "元"
}

// 转换小节（4位数）
func convertSection(section int) string {
	var result strings.Builder
	zeroFlag := true

	for i := 0; section > 0; i++ {
		digit := section % 10
		section /= 10

		if digit == 0 {
			if !zeroFlag {
				result.WriteString(digitUpper[0])
				zeroFlag = true
			}
		} else {
			zeroFlag = false
			result.WriteString(unitUpper[i])
			result.WriteString(digitUpper[digit])
		}
	}

	runes := []rune(result.String())
	// 反转字符串
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// 转换小数部分
func convertDecimalPart(num float64) string {
	decimal := int((num - float64(int(num))) * 100)
	if decimal == 0 {
		return "整"
	}

	jiao := decimal / 10
	fen := decimal % 10

	var result strings.Builder
	if jiao > 0 {
		result.WriteString(digitUpper[jiao] + "角")
	}
	if fen > 0 {
		result.WriteString(digitUpper[fen] + "分")
	}

	return result.String()
}

// MoneyToUpper 金额转大写
func MoneyToUpper(amount float64) string {
	integerPart := int64(amount)
	decimalPart := amount - float64(integerPart)

	return convertIntegerPart(integerPart) + convertDecimalPart(decimalPart)
}
