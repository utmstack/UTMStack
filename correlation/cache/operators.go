package cache

import (
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

func inCIDR(addr, network string) bool {
	_, subnet, err := net.ParseCIDR(network)
	if err == nil {
		ip := net.ParseIP(addr)
		if ip != nil {
			if subnet.Contains(ip) {
				return true
			}
		}
	}
	return false
}

func equal(val1, val2 string) bool {
	return val1 == val2
}

func lowerEqual(val1, val2 string) bool {
	return equal(strings.ToLower(val1), strings.ToLower(val2))
}

func contain(str, substr string) bool {
	return strings.Contains(str, substr)
}

func in(str, list string) bool {
	l := strings.Split(list, ",")
	for _, value := range l {
		if str == value {
			return true
		}
	}
	return false
}

func startWith(str, pref string) bool {
	return strings.HasPrefix(str, pref)
}

func endWith(str, suff string) bool {
	return strings.HasSuffix(str, suff)
}

func expresion(exp, str string) bool {
	re, err := regexp.Compile(exp)
	if err == nil {
		if re.MatchString(str) {
			return true
		}
	}
	return false
}

func minThan(min, may string) bool {
	minN, err := strconv.ParseFloat(min, 64)
	if err != nil {
		return false
	}
	mayN, err := strconv.ParseFloat(may, 64)
	if err != nil {
		return false
	}

	return minN < mayN
}

func compare(operator, val1, val2 string) bool {
	switch operator {
	case "==":
		return equal(val1, val2)
	case "!=":
		return !equal(val1, val2)
	case "<>":
		return !equal(val1, val2)
	case "::":
		return lowerEqual(val1, val2)
	case "!!":
		return !lowerEqual(val1, val2)
	case "contains":
		return contain(val1, val2)
	case "not contain":
		return !contain(val1, val2)
	case "in":
		return in(val1, val2)
	case "not in":
		return !in(val1, val2)
	case "start with":
		return startWith(val1, val2)
	case "not start with":
		return !startWith(val1, val2)
	case "end with":
		return endWith(val1, val2)
	case "not end with":
		return !endWith(val1, val2)
	case "regexp":
		return expresion(val2, val1)
	case "not regexp":
		return !expresion(val2, val1)
	case "<":
		return minThan(val1, val2)
	case ">":
		return !minThan(val1, val2)
	case "<=":
		return equal(val1, val2) || minThan(val1, val2)
	case ">=":
		return equal(val1, val2) || !minThan(val1, val2)
	case "exist":
		return true
	case "in cidr":
		return inCIDR(val1, val2)
	case "not in cidr":
		return !inCIDR(val1, val2)
	default:
		return false
	}
}

func evalElement(elem, field, operator, value string) bool {
	if gjson.Get(elem, field).Exists() {
		return compare(operator, gjson.Get(elem, field).String(), value)
	} else if operator == "not exist" {
		return true
	}
	return false
}
